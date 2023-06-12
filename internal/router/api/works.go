package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx/types"
	log "github.com/sirupsen/logrus"
	"kikitoru/config"
	"kikitoru/internal/database"
	"net/http"
	"strconv"
)

type HttpResponse struct {
	Message     string
	Status      int
	Description string
}

type WorkDB struct {
	ID              int            `json:"id"`
	Title           string         `json:"title"`
	CircleID        int            `json:"circle_id" db:"circle_id"`
	Name            string         `json:"name"`
	Nsfw            bool           `json:"nsfw"`
	Release         string         `json:"release" time_format:"2006-01-02"`
	DLCount         int            `json:"dl_count" db:"dl_count"`
	Price           int            `json:"price"`
	ReviewCount     int            `json:"review_count" db:"review_count"`
	RateCount       int            `json:"rate_count" db:"rate_count"`
	RateAverage2dp  float32        `json:"rate_average_2dp" db:"rate_average_2dp"`
	RateCountDetail types.JSONText `json:"rate_count_detail" db:"rate_count_detail"`
	Rank            types.JSONText `json:"rank"`
	CreateDate      string         `json:"create_date" db:"create_date" time_format:"2006-01-02"`
	UserRating      *int           `json:"userRating" db:"rating"`
	Circle          types.JSONText `json:"circle" db:"circleobj"`
	Vas             types.JSONText `json:"vas" db:"vaobj"`
	Tags            types.JSONText `json:"tags" db:"tagobj"`
	HasSubtitle     bool           `json:"has_subtitle" db:"has_subtitle"`
	//Series          StructStringIDName   `json:"series"`
}

type Pagination struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
	TotalCount  int `json:"totalCount"`
}

type Works struct {
	Works      []WorkDB   `json:"works"`
	Pagination Pagination `json:"pagination"`
}

type QueryField struct {
	tag      int
	rj       int
	rg       int
	va       uuid.UUID
	userName string
	field    string
}

func GetWorks(c *gin.Context) {
	err, order, sort, page := GetWorkQuery(c)
	if err != nil {
		log.Error(err)
		return
	}

	c.JSON(200, queryWorks(QueryField{userName: c.GetString("username")}, order, sort, page))
	//queryWorks(order, sort, page, seed)
}

func queryWorks(queryField QueryField, order string, sort string, page int) Works {
	db := database.GetDB()
	var works []WorkDB

	limit := config.C.PageSize

	var count int
	var err error

	var sqlOrder string
	switch order {
	//case "id":
	//sqlOrder = "substring(m.%s FROM '\\d+')::integer"
	case "random":
		sqlOrder = "%s()"
	case "betterRandom": // 随心听
		queryField.field = "betterRandom"
	case "rating":
		sort = "desc NULLS LAST"
		sqlOrder = "%s"
	default:
		sqlOrder = "%s"
	}
	//fmt.Println(order)

	sqlBase := `
	SELECT m.*,r.rating
	FROM "staticMetadata" as m
	LEFT JOIN (
		SELECT * FROM t_review
		JOIN t_work on t_review.work_id = t_work.id
		WHERE t_review.user_name='%s'
    ) as r on substring(r.work_id,3)::integer = m.id`

	username := queryField.userName
	var sqlText string
	switch queryField.field {
	case "circle":
		sqlText = sqlBase + " WHERE m.circle_id=%d ORDER BY " + sqlOrder + " %s LIMIT %d OFFSET %d"
		err = db.Select(&works,
			fmt.Sprintf(sqlText, username, queryField.rg, order, sort, limit, (page-1)*limit))
		if err != nil {
			log.Error(err)
		}
		err := db.QueryRow("SELECT COUNT(*) FROM \"staticMetadata\" WHERE circle_id=$1", queryField.rg).Scan(&count)
		if err != nil {
			log.Error(err)
		}
	case "tag":
		sqlText = sqlBase + " WHERE m.id in (SELECT substring(work_id,3)::integer as work_id FROM r_tag_work WHERE tag_id=%d) ORDER BY " + sqlOrder + " %s LIMIT %d OFFSET %d"
		err = db.Select(&works,
			fmt.Sprintf(sqlText, username, queryField.tag, order, sort, limit, (page-1)*limit))
		if err != nil {
			log.Error(err)
		}

		err := db.QueryRow("SELECT COUNT(*) FROM r_tag_work WHERE tag_id=$1", queryField.tag).Scan(&count)
		if err != nil {
			log.Error(err)
		}
	case "va":
		sqlText = sqlBase + " WHERE m.id in (SELECT substring(work_id,3)::integer as work_id FROM r_va_work WHERE va_id='%s') ORDER BY " + sqlOrder + " %s LIMIT %d OFFSET %d"
		err = db.Select(&works,
			fmt.Sprintf(sqlText, username, queryField.va, order, sort, limit, (page-1)*limit))
		if err != nil {
			log.Error(err)
		}

		err := db.QueryRow("SELECT COUNT(*) FROM r_va_work WHERE va_id=$1", queryField.va).Scan(&count)
		if err != nil {
			log.Error(err)
		}
	case "betterRandom":
		sqlText = sqlBase + " ORDER BY random() LIMIT 1"
		err = db.Select(&works,
			fmt.Sprintf(sqlText, username))
		if err != nil {
			log.Error(err)
		}
		count = 1
	default:
		sqlText = sqlBase + " ORDER BY " + sqlOrder + " %s LIMIT %d OFFSET %d"
		err = db.Select(&works,
			fmt.Sprintf(sqlText, username, order, sort, limit, (page-1)*limit))
		if err != nil {
			log.Error(err)
		}

		err := db.QueryRow("SELECT COUNT(*) FROM \"staticMetadata\"").Scan(&count)
		if err != nil {
			log.Error(err)
		}
	}

	//fmt.Println(works)
	//fmt.Printf("%+v\n", works)
	// handle nil
	if works == nil {
		works = []WorkDB{}
	}

	return Works{Works: works, Pagination: Pagination{CurrentPage: page, PageSize: limit, TotalCount: count}}
}

func GetWorkQuery(c *gin.Context) (error, string, string, int) {
	var err error

	// Validity check
	order := c.Query("order")
	switch order {
	case "release", "create_date", "rating", "dl_count", "price", "rate_average_2dp", "review_count", "id", "nsfw", "random":
	case "betterRandom":
		return err, order, "", 0
	case "insert_time": // 兼容
		order = "create_date"
	default:
		err = errors.New("Illegal order parameter: " + order)
		resp := HttpResponse{Message: "Bad Request", Status: http.StatusBadRequest, Description: err.Error()}
		c.AbortWithStatusJSON(http.StatusBadRequest, resp)
		return err, "", "", 0
	}

	sort := c.Query("sort")
	switch sort {
	case "asc", "desc":
	default:
		err = errors.New("illegal sort parameter: " + sort)
		resp := HttpResponse{Message: "Bad Request", Status: http.StatusBadRequest, Description: err.Error()}
		c.AbortWithStatusJSON(http.StatusBadRequest, resp)
		return err, "", "", 0
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		resp := HttpResponse{Message: "Bad Request", Status: http.StatusBadRequest, Description: err.Error()}
		c.AbortWithStatusJSON(http.StatusBadRequest, resp)
		return err, "", "", 0
	}

	//seed, err := strconv.Atoi(c.Query("seed"))
	//if err != nil {
	//	resp := HttpResponse{Message: "Bad Request", Status: http.StatusBadRequest, Description: err.Error()}
	//	c.AbortWithStatusJSON(http.StatusBadRequest, resp)
	//	return err, "", "", 0, 0
	//}

	return err, order, sort, page
}
