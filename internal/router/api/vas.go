package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	log "github.com/sirupsen/logrus"
	"kikitoru/internal/database"
	"net/http"
)

func GetVAs(c *gin.Context) {
	var vas []struct {
		ID    string `json:"id" db:"id"`
		Name  string `json:"name" db:"name"`
		Count int    `json:"count" db:"count"`
	}
	db := database.GetDB()
	err := db.Select(&vas,
		`SELECT t_va.id,t_va.name,count(r_va_work.work_id)
					FROM t_va
					LEFT JOIN r_va_work on t_va.id = r_va_work.va_id
					GROUP BY t_va.id, t_va.name;`)
	if err != nil {
		log.Error(err)
	}

	c.JSON(http.StatusOK, vas)
}

func GetVAsByID(c *gin.Context) {
	va := c.Param("id")

	db := database.GetDB()
	type structVAs struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var vas structVAs
	err := db.Get(&vas, "SELECT * FROM t_va WHERE id=$1", va)
	if err != nil {
		return
	}
	c.JSON(200, vas)
}

func GetVAsWorksByID(c *gin.Context) {
	va := c.Param("id")

	uuidVA, err := uuid.FromString(va)
	if err != nil {
		fmt.Println(err)
		return
	}

	queryField := QueryField{va: uuidVA, field: "va", userName: c.GetString("username")}

	err, order, sort, page := GetWorkQuery(c)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.JSON(200, queryWorks(queryField, order, sort, page))
}
