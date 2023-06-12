package api

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"kikitoru/internal/database"
	"net/http"
	"strconv"
	"strings"
)

func GetCircles(c *gin.Context) {
	var circles []struct {
		ID    string `json:"id" db:"id"`
		Name  string `json:"name" db:"name"`
		Count int    `json:"count" db:"count"`
	}
	db := database.GetDB()
	err := db.Select(&circles,
		`SELECT t_circle.id,t_circle.name,count(t_work.id)
					FROM t_circle
					JOIN r_circle_work on r_circle_work.circle_id = t_circle.id
					JOIN t_work on r_circle_work.work_id = t_work.id
					GROUP BY t_circle.id, t_circle.name;`)
	if err != nil {
		log.Error(err)
	}

	c.JSON(http.StatusOK, circles)
}

func GetCircleByID(c *gin.Context) {
	rg := c.Param("id")

	var circle struct {
		ID   string `json:"id" db:"id"`
		Name string `json:"name" db:"name"`
	}

	db := database.GetDB()
	err := db.Get(&circle, "SELECT * FROM t_circle WHERE id=$1", rg)
	if err != nil {
		return
	}
	c.JSON(200, circle)
}

func GetCircleWorksByID(c *gin.Context) {
	rg := c.Param("id")
	rgInt, _ := strconv.Atoi(strings.TrimPrefix(rg, "RG"))
	queryField := QueryField{rg: rgInt, field: "circle", userName: c.GetString("username")}

	err, order, sort, page := GetWorkQuery(c)
	if err != nil {
		log.Error(err)
		return
	}

	c.JSON(200, queryWorks(queryField, order, sort, page))
}
