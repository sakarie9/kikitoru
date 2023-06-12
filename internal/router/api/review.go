package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx/types"
	log "github.com/sirupsen/logrus"
	"kikitoru/config"
	"kikitoru/internal/database"
	"kikitoru/util"
	"net/http"
	"strconv"
)

type ReviewDB struct {
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
	ReviewText      *string        `json:"review_text" db:"review_text"`
	Progress        *string        `json:"progress" db:"progress"`
	Circle          types.JSONText `json:"circle" db:"circleobj"`
	Vas             types.JSONText `json:"vas" db:"vaobj"`
	Tags            types.JSONText `json:"tags" db:"tagobj"`
	HasSubtitle     bool           `json:"has_subtitle" db:"has_subtitle"`
	//Series          StructStringIDName   `json:"series"`
}

type Reviews struct {
	Works      []ReviewDB `json:"works"`
	Pagination Pagination `json:"pagination"`
}

func GetReview(c *gin.Context) {
	// 获取filter参数
	filter := c.Query("filter")
	switch filter {
	case "marked", "listening", "listened", "replay", "postponed":
	case "":
	default:
		err := errors.New("Illegal filter parameter: " + filter)
		resp := HttpResponse{Message: "Bad Request", Status: http.StatusBadRequest, Description: err.Error()}
		c.AbortWithStatusJSON(http.StatusBadRequest, resp)
	}
	// 获取 page 参数
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		err := errors.New("Illegal page parameter: " + strconv.Itoa(page))
		resp := HttpResponse{Message: "Bad Request", Status: http.StatusBadRequest, Description: err.Error()}
		c.AbortWithStatusJSON(http.StatusBadRequest, resp)
	}
	c.JSON(200, queryWorksReview(filter, page, c.GetString("username")))
}

func queryWorksReview(filter string, page int, username string) Reviews {
	db := database.GetDB()
	var works []ReviewDB
	var count int

	limit := config.C.PageSize

	if filter == "" {
		err := db.Select(&works, `SELECT m.*,r.rating,r.progress,r.review_text
		FROM "staticMetadata" as m
		LEFT JOIN (
			SELECT * FROM t_review
			JOIN t_work on t_review.work_id = t_work.id
			WHERE t_review.user_name=$1
		) as r on substring(r.work_id,3)::integer = m.id
		WHERE rating IS NOT NULL OR progress IS NOT NULL OR review_text IS NOT NULL
		ORDER BY updated_at desc
		LIMIT $2 OFFSET $3`,
			username, limit, (page-1)*limit)
		if err != nil {
			log.Error(err)
		}

		err = db.QueryRow("SELECT COUNT(*) FROM t_review").Scan(&count)
		if err != nil {
			log.Error(err)
		}
	} else {
		err := db.Select(&works, `SELECT m.*,r.rating,r.progress,r.review_text
		FROM "staticMetadata" as m
		LEFT JOIN (
			SELECT * FROM t_review
			JOIN t_work on t_review.work_id = t_work.id
			WHERE t_review.user_name=$1
		) as r on substring(r.work_id,3)::integer = m.id WHERE progress=$2 ORDER BY updated_at desc LIMIT $3 OFFSET $4`,
			username, filter, limit, (page-1)*limit)
		if err != nil {
			log.Error(err)
		}
		err = db.QueryRow("SELECT COUNT(*) FROM t_review WHERE progress=$1", filter).Scan(&count)
		if err != nil {
			log.Error(err)
		}
	}

	if works == nil {
		works = []ReviewDB{}
	}
	return Reviews{Works: works, Pagination: Pagination{CurrentPage: page, PageSize: limit, TotalCount: count}}
}

/*
写评论
{"user_name":"admin","work_id":1018155,"rating":0,"review_text":"114514","progress":null}
{"work_id":1058649,"rating":3,"review_text":"123","progress":"listening"}
{"message":"更新成功"}
标记进度
{"work_id":1058649,"progress":"replay"}
点星
{"work_id":1058649,"rating":2}
*/
func PutReview(c *gin.Context) {
	type structReviewReq struct {
		WorkID     int     `json:"work_id"`
		Rating     *int    `json:"rating"`
		ReviewText *string `json:"review_text"`
		Progress   *string `json:"progress"`
	}
	var reviewReq structReviewReq
	err := c.BindJSON(&reviewReq)
	if err != nil {
		log.Error(err)
		return
	}

	username := c.GetString("username")
	workRJ := util.IDToRJ(reviewReq.WorkID)

	db := database.GetDB()
	tx, err := db.Beginx()
	if err != nil {
		log.Error(err)
	}

	if reviewReq.Rating != nil {
		_, err = tx.Exec("INSERT INTO t_review(user_name,work_id,rating) VALUES($1,$2,$3) ON CONFLICT(user_name,work_id) DO UPDATE SET rating=excluded.rating,updated_at=current_timestamp",
			username, workRJ, *reviewReq.Rating)
		if err != nil {
			log.Error(err)
			tx.Rollback()
			return
		}
	}

	if reviewReq.ReviewText != nil {
		_, err = tx.Exec("INSERT INTO t_review(user_name,work_id,review_text) VALUES($1,$2,$3) ON CONFLICT(user_name,work_id) DO UPDATE SET review_text=excluded.review_text,updated_at=current_timestamp",
			username, workRJ, *reviewReq.ReviewText)
		if err != nil {
			log.Error(err)
			tx.Rollback()
			return
		}
	}

	if reviewReq.Progress != nil {
		_, err = tx.Exec("INSERT INTO t_review(user_name,work_id,progress) VALUES($1,$2,$3) ON CONFLICT(user_name,work_id) DO UPDATE SET progress=excluded.progress,updated_at=current_timestamp",
			username, workRJ, *reviewReq.Progress)
		if err != nil {
			log.Error(err)
			tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

/*
review?work_id=1058649
{"message":"删除标记成功"}
*/
func DeleteReview(c *gin.Context) {
	id, _ := strconv.Atoi(c.Query("work_id"))
	rj := util.IDToRJ(id)
	username := c.GetString("username")

	db := database.GetDB()
	_, err := db.Exec("DELETE FROM t_review WHERE user_name=$1 and work_id=$2", username, rj)
	if err != nil {
		log.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除标记成功"})
}
