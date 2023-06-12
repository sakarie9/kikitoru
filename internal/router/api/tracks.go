package api

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"kikitoru/internal/filesystem"
	"kikitoru/util"
	"net/http"
	"strconv"
)

func GetTracks(c *gin.Context) {
	rj := c.Param("id")
	rjInt, _ := strconv.Atoi(rj)
	if rj == "undefined" {
		log.Error("Get undefined tracks")
		return
	}

	tree := filesystem.ScanTracks(util.IDToRJ(rjInt))
	if tree == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件列表失败，请检查文件是否存在或重新扫描清理"})
		return
	}
	c.JSON(http.StatusOK, tree)
}
