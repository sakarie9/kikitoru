package scraper

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gofrs/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"io"
	"kikitoru/config"
	"kikitoru/logs"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type StructIDName struct {
	ID   int
	Name string
}

type StructStringName struct {
	ID   string
	Name string
}

type ScrapedWorkMetadata struct {
	ID              string
	Title           string
	Nsfw            bool
	Release         string
	DLCount         int
	Price           int
	ReviewCount     int
	RateCount       int
	RateAverage2dp  float32
	RateCountDetail string
	Rank            string
	Circle          StructStringName
	Vas             []StructStringName
	Tags            []StructIDName
	Series          StructStringName
	RootFolder      string
	Dir             string
	Lrc             bool
}

//var work ScrapedWorkMetadata

func ScrapeStaticWorkMetadataFromDLsite(work *ScrapedWorkMetadata, rj string) {

	var AGE_RATINGS, VA, GENRE, RELEASE, SERIES, COOKIE_LOCALE string

	switch config.C.TagLanguage {
	case "ja-jp":
		COOKIE_LOCALE = "locale=ja-jp"
		AGE_RATINGS = "年齢指定"
		GENRE = "ジャンル"
		VA = "声優"
		RELEASE = "販売日"
		SERIES = "シリーズ名"
	case "zh-tw":
		COOKIE_LOCALE = "locale=zh-tw"
		AGE_RATINGS = "年齡指定"
		GENRE = "分類"
		VA = "聲優"
		RELEASE = "販賣日"
		SERIES = "系列名"
	default:
		COOKIE_LOCALE = "locale=zh-cn"
		AGE_RATINGS = "年龄指定"
		GENRE = "分类"
		VA = "声优"
		RELEASE = "贩卖日"
		SERIES = "系列名"
	}

	//_, _, _, _, _, _ = AGE_RATINGS, VA, GENRE, RELEASE, SERIES, COOKIE_LOCALE

	work.ID = rj

	c := colly.NewCollector()

	c.Limit(&colly.LimitRule{
		RandomDelay: 500 * time.Millisecond, // 两次请求 随机延迟5s 内
	})

	// 标题
	c.OnHTML("meta[property=\"og:title\"]", func(e *colly.HTMLElement) {
		title := e.Attr("content")
		work.Title = strings.Replace(title, " | DLsite", "", -1)
		//fmt.Println(work.Title)
	})
	// 社团
	c.OnHTML("span[class=\"maker_name\"]", func(e *colly.HTMLElement) {
		circleUrl := e.ChildAttr("a", "href")
		circleName := e.ChildText("a")
		if circleUrl != "" && circleName != "" {
			var err error
			work.Circle.ID, _, err = matchFromString(circleUrl, `RG(\d+)`)
			if err != nil {
				log.Error(err)
			}
			//work.CircleID = work.Circle.ID
			work.Circle.Name = circleName
			//work.Name = circleName
			//fmt.Println(work.Circle.name, work.Circle.id)
		}
	})
	// 各种标签
	c.OnHTML("#work_outline", func(e *colly.HTMLElement) {
		e.ForEach("tbody > tr", func(_ int, e *colly.HTMLElement) {
			th := e.ChildText("th")
			switch th {
			// NSFW
			case AGE_RATINGS:
				td := e.DOM.Find("span").Text()
				if td == "18禁" {
					work.Nsfw = true
				} else {
					work.Nsfw = false
				}
				//fmt.Println("全年龄：", !work.Nsfw)

			// 贩卖日 (YYYY-MM-DD)
			case RELEASE:
				td := e.DOM.Find("a[href]").Text()
				//const TIME_LAYOUT = "2006年01月02日"
				//t, _ := time.Parse(TIME_LAYOUT, td)
				//work.releaseDate = t.Format("2006-01-02")
				date := fmt.Sprintf("%s-%s-%s", td[:4], td[7:9], td[12:14])
				work.Release = date
				//fmt.Println(work.Release)

			// 系列
			case SERIES:
				seriesElement := e.DOM.Find("a[href]")
				seriesUrl, _ := seriesElement.Attr("href")
				id, _, err := matchFromString(seriesUrl, `SRI(\d+)`)
				if err != nil {
					log.Error(err)
				}
				work.Series.ID = id
				work.Series.Name = seriesElement.Text()
				//fmt.Println(work.Series)

			// 标签
			case GENRE:
				genreElements := e.DOM.Find(".main_genre > a")
				genreElements.Each(func(_ int, el *goquery.Selection) {
					link, _ := el.Attr("href")
					_, id, err := matchFromString(link, `genre/(\d+)`)
					if err != nil {
						log.Error(err)
					}
					text := el.Text()
					work.Tags = append(work.Tags, StructIDName{ID: id, Name: text})

				})
				//fmt.Println(work.Tags)

			// 声优
			case VA:
				vaElements := e.DOM.Find("a[href]")
				vaElements.Each(func(_ int, el *goquery.Selection) {
					vaName := el.Text()
					id := nameToUUID(vaName)
					work.Vas = append(work.Vas, StructStringName{ID: id, Name: vaName})
				})
				//fmt.Println(work.Vas)
			}
		})
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("cookie", COOKIE_LOCALE)
		log.Debugf("Visiting %s", r.URL)
	})

	err := c.Visit(fmt.Sprintf("https://www.dlsite.com/maniax/work/=/product_id/%s.html", rj))
	if err != nil {
		log.Errorf("%s: %s", rj, err)
	}

	// 声优不存在时
	if work.Vas == nil {
		work.Vas = []StructStringName{{ID: nameToUUID("undefined"), Name: "undefined"}}
	}
}

func ScrapeDynamicWorkMetadataFromDLsite(work *ScrapedWorkMetadata, rj string) {
	url := fmt.Sprintf("https://www.dlsite.com/maniax-touch/product/info/ajax?product_id=%s", rj)
	log.Debugf("Visiting %s", url)
	// 发起 HTTP 请求
	resp, err := http.Get(url)
	if err != nil {
		log.Warn(rj+" 发起请求时出错:", err)
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warn(rj+" 读取响应时出错:", err)
		return
	}
	//workRoot := gjson.GetBytes(body, rj)
	getResult := func(body []byte, path string, isNum bool) string {
		value := gjson.GetBytes(body, rj+"."+path)
		if value.Exists() {
			return value.String()
		} else if isNum {
			return "0"
		} else {
			return ""
		}
	}

	work.DLCount, _ = strconv.Atoi(getResult(body, "dl_count", true))

	float, _ := strconv.ParseFloat(getResult(body, "rate_average_2dp", true), 32)
	work.RateAverage2dp = float32(float)

	work.RateCount, _ = strconv.Atoi(getResult(body, "rate_count", true))

	//var rateCountDetail []model.RateCountDetail
	//_ = json.Unmarshal([]byte(getResult(body, "rate_count_detail", false)), &rateCountDetail)
	//work.RateCountDetail = rateCountDetail
	work.RateCountDetail = getResult(body, "rate_count_detail", false)

	work.ReviewCount, _ = strconv.Atoi(getResult(body, "review_count", true))

	work.Price, _ = strconv.Atoi(getResult(body, "price", true))

	//var rankDetail []model.Rank
	//_ = json.Unmarshal([]byte(getResult(body, "rank", false)), &rankDetail)
	//work.Rank = rankDetail
	work.Rank = getResult(body, "rank", false)

}

func GetScrapedWork(work ScrapedWorkMetadata) ScrapedWorkMetadata {
	ScrapeStaticWorkMetadataFromDLsite(&work, work.ID)
	ScrapeDynamicWorkMetadataFromDLsite(&work, work.ID)
	log.Infof("%s: 成功获取元数据", work.ID)
	logs.ScanLogs.Details.Enqueue(fmt.Sprintf("%s: 成功获取元数据", work.ID))
	return work
}

func matchFromString(str string, pattern string) (string, int, error) {
	regex := regexp.MustCompile(pattern)
	match := regex.FindStringSubmatch(str)
	var id int
	var err error
	if len(match) > 1 {
		id, err = strconv.Atoi(match[1])
	} else if len(match) == 0 {
		return "", 0, err
	}
	return match[0], id, err
}

func nameToUUID(name string) string {
	namespace, _ := uuid.FromString("699d9c07-b965-4399-bafd-18a3cacf073c")
	u1 := uuid.NewV5(namespace, name)
	return u1.String()
}
