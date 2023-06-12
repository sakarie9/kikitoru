package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"kikitoru/config"
	"kikitoru/util"
	"path"
	"strconv"
)

func GetCover(c *gin.Context) {
	rj := c.Param("id")
	t := "main"
	var coverPath string
	if rj[0] == 'R' {
		coverPath = path.Join(config.C.CoverFolderDir, fmt.Sprintf("%s_img_%s.jpg", rj, t))
	} else {
		rjInt, _ := strconv.Atoi(rj)
		coverPath = path.Join(config.C.CoverFolderDir, fmt.Sprintf("%s_img_%s.jpg", util.IDToRJ(rjInt), t))
	}
	log.Debug(coverPath)
	c.File(coverPath)
}
