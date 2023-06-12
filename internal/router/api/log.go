package api

import (
	"github.com/gin-gonic/gin"
	"kikitoru/internal/filesystem"
	"kikitoru/logs"
	"net/http"
)

// api/logs
func GetLogs(c *gin.Context) {

	c.JSON(http.StatusOK, logs.ScanLogs)
}

func StartScan(c *gin.Context) {
	filesystem.StartScanWorks()
	c.JSON(http.StatusOK, gin.H{"message": "开始扫描"})
}
