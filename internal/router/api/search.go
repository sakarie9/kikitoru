package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"kikitoru/config"
	"kikitoru/internal/database"
)

func GetSearch(c *gin.Context) {
	err, order, sort, page := GetWorkQuery(c)
	if err != nil {
		log.Error(err)
		return
	}
	keyword := c.Param("keyword")
	log.Debug(keyword)

	c.JSON(200, searchWorks(keyword, c.GetString("username"), order, sort, page))
}

func searchWorks(keyword string, username string, order string, sort string, page int) Works {
	db := database.GetDB()
	var works []WorkDB

	limit := config.C.PageSize

	var count int
	var err error

	query := `
SELECT m.*,r.rating
FROM "staticMetadata" AS m
LEFT JOIN (
    SELECT * FROM t_review
    JOIN t_work ON t_review.work_id = t_work.id
    WHERE t_review.user_name='%s'
) AS r ON substring(r.work_id,3)::integer = m.id
WHERE m.id in (SELECT substring(id,3)::integer FROM search_view WHERE record_to_text(search_view) ~ '%s')`

	var sqlOrder string
	switch order {
	case "random":
		sqlOrder = "%s()"
	default:
		sqlOrder = "%s"
	}

	err = db.Select(&works,
		fmt.Sprintf(query+" ORDER BY "+sqlOrder+" %s LIMIT %d OFFSET %d",
			username, keyword, order, sort, limit, (page-1)*limit))
	if err != nil {
		log.Error(err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM \"staticMetadata\" WHERE id in (SELECT substring(id,3)::integer FROM search_view WHERE record_to_text(search_view) ~ $1)",
		keyword).Scan(&count)
	if err != nil {
		log.Error(err)
	}

	if works == nil {
		works = []WorkDB{}
	}
	return Works{Works: works, Pagination: Pagination{CurrentPage: page, PageSize: limit, TotalCount: count}}
}
