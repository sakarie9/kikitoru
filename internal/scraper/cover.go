package scraper

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	log "github.com/sirupsen/logrus"
	"io"
	"kikitoru/config"
	"kikitoru/logs"
	"kikitoru/util"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var CoverTypes = []string{"main", "sam", "240x240"}

/*
GetCoverImage
从 DLsite 下载封面图片，并保存到 Images 文件夹，
返回一个 Promise 对象，处理结果: 'added' or 'failed'
@param {number} id work id
@param {Array} types img types: ['main', 'sam', 'sam@2x', 'sam@3x', '240x240', '360x360']
*/
func GetCoverImage(rj string) {

	rjCode, rjCode2 := getCoverRJCodes(rj)
	//fmt.Println(rjCode, rjCode2)

	p, _ := ants.NewPool(config.C.MaxParallelism)
	defer p.Release()

	for _, t := range CoverTypes {
		coverUrl := fmt.Sprintf(
			"https://img.dlsite.jp/modpub/images2/work/doujin/%s/%s_img_%s.jpg", rjCode2, rjCode, t)
		if t == "240x240" || t == "360x360" {
			coverUrl = fmt.Sprintf(
				"https://img.dlsite.jp/resize/images2/work/doujin/%s/%s_img_main_%s.jpg", rjCode2, rjCode, t)
		}
		//fmt.Println(coverUrl)
		err := DownloadFile(config.C.CoverFolderDir, coverUrl)
		if err != nil {
			log.Warnf("%s: 封面下载错误，开始重试", rj)
			var retryTimes int
			for retryTimes = 0; retryTimes < config.C.Retry; retryTimes++ {
				time.Sleep(config.C.RetryDelay) // 失败等待
				log.Warnf("%s: 封面下载错误，第 %d 次重试", rj, retryTimes)
				err1 := DownloadFile(config.C.CoverFolderDir, coverUrl)
				if err1 == nil {
					break
				}
			}
			log.Warnf("%s: 封面下载错误，重试失败", rj)
		}
	}

}

func getCoverRJCodes(rjCode string) (string, string) {

	id := util.RJToID(rjCode)

	var id2 int
	if id%1000 == 0 {
		id2 = id
	} else {
		id2 = (id/1000)*1000 + 1000
	}

	var rjCode2 string
	if id2 >= 1000000 {
		rjCode2 = fmt.Sprintf("RJ%08d", id2)
	} else {
		rjCode2 = fmt.Sprintf("RJ%06d", id2)
	}
	return rjCode, rjCode2
}

func DownloadFile(filePath string, fullURL string) error {

	err := os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		return err
	}

	// Build fileName from fullPath
	fileURL, err := url.Parse(fullURL)
	if err != nil {
		return err
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	fullPath, err := url.JoinPath(filePath, fileName)
	if err != nil {
		return err
	}

	// Check if Exists
	fileInfo, err := os.Stat(fullPath)
	if err == nil && fileInfo.Size() != 0 { // path/to/whatever exists
		return nil
	}
	// Create blank file
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}

	// 设置代理
	var client http.Client
	if config.C.HTTPProxyHost != "" && config.C.HTTPProxyPort != 0 {
		proxyUrl := fmt.Sprintf("http://%s:%d", config.C.HTTPProxyHost, config.C.HTTPProxyPort)
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(proxyUrl)
		}
		httpTransport := &http.Transport{
			Proxy: proxy,
		}

		client = http.Client{Transport: httpTransport,
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}
	} else { // 未设置代理
		client = http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}
	}

	// Put content on file
	resp, err := client.Get(fullURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)

	defer file.Close()

	log.Infof("%s: 成功下载封面 大小: %d", fileName, size)
	logs.ScanLogs.Details.Enqueue(fmt.Sprintf("%s: 成功下载封面 大小: %d", fileName, size))

	return nil

}
