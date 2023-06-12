package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx/types"
	log "github.com/sirupsen/logrus"
	"kikitoru/internal/database"
	"strconv"
)

func GetWork(c *gin.Context) {
	rj := c.Param("id")
	rjInt, _ := strconv.Atoi(rj)

	c.JSON(200, queryWork(rjInt, c.GetString("username")))

}

type workDetail struct {
	ID              int            `db:"id" json:"id"`
	Title           string         `db:"title" json:"title"`
	CircleID        int            `db:"circle_id" json:"circle_id"`
	Name            string         `db:"name" json:"name"`
	Nsfw            bool           `db:"nsfw" json:"nsfw"`
	Release         string         `db:"release" json:"release" time_format:"2006-01-02"`
	DLCount         int            `db:"dl_count" json:"dl_count"`
	Price           int            `db:"price" json:"price"`
	ReviewCount     int            `db:"review_count" json:"review_count"`
	RateCount       int            `db:"rate_count" json:"rate_count"`
	RateAverage2dp  float32        `db:"rate_average_2dp" json:"rate_average_2dp"`
	RateCountDetail types.JSONText `db:"rate_count_detail" json:"rate_count_detail"`
	Rank            types.JSONText `db:"rank" json:"rank"`
	CreateDate      string         `db:"create_date" json:"create_date" time_format:"2006-01-02"`
	UserRating      *int           `db:"rating" json:"userRating"`
	ReviewText      *string        `db:"review_text" json:"review_text"`
	Progress        *string        `db:"progress" json:"progress"`
	UpdatedAt       *string        `db:"updated_at" json:"updated_at"`
	UserName        *string        `db:"user_name" json:"user_name"`
	Circle          types.JSONText `db:"circleobj" json:"circle"`
	VAs             types.JSONText `db:"vaobj" json:"vas"`
	Tags            types.JSONText `db:"tagobj" json:"tags"`
	HasSubtitle     bool           `json:"has_subtitle" db:"has_subtitle"`
}

func queryWork(rjInt int, username string) workDetail {
	db := database.GetDB()
	var work workDetail

	sqlText := `SELECT "staticMetadata".*,r.rating,r.review_text,r.progress,to_char(r.updated_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at,r.user_name
	FROM "staticMetadata"
	LEFT JOIN (
		SELECT * FROM t_review
		JOIN t_work on t_review.work_id = t_work.id
		WHERE t_review.user_name=$1
		) as r on substring(r.work_id,3)::integer = "staticMetadata".id
	WHERE "staticMetadata".id=$2 LIMIT 1`
	err := db.Get(&work, sqlText, username, rjInt)
	if err != nil {
		log.Error(err)
	}
	//fmt.Printf("%+v\n", work)
	//work.UserName = username
	return work
}
