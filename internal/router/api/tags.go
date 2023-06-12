package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"kikitoru/internal/database"
	"net/http"
	"strconv"
)

func GetTags(c *gin.Context) {
	var tags []struct {
		ID    int    `json:"id" db:"id"`
		Name  string `json:"name" db:"name"`
		Count int    `json:"count" db:"count"`
	}
	db := database.GetDB()
	err := db.Select(&tags,
		`SELECT t_tag.id,t_tag.name,count(r_tag_work.work_id)
					FROM t_tag
					LEFT JOIN r_tag_work on t_tag.id = r_tag_work.tag_id
					GROUP BY t_tag.id, t_tag.name;`)
	if err != nil {
		log.Error(err)
	}

	c.JSON(http.StatusOK, tags)
}

func GetTagsByID(c *gin.Context) {
	id := c.Param("id")
	tag, _ := strconv.Atoi(id)

	db := database.GetDB()
	var tags struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	err := db.Get(&tags, "SELECT * FROM t_tag WHERE id=$1", tag)
	if err != nil {
		return
	}
	c.JSON(200, tags)
}

func GetTagsWorksByID(c *gin.Context) {
	id := c.Param("id")
	tag, _ := strconv.Atoi(id)
	queryField := QueryField{tag: tag, field: "tag", userName: c.GetString("username")}

	err, order, sort, page := GetWorkQuery(c)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.JSON(200, queryWorks(queryField, order, sort, page))
}
