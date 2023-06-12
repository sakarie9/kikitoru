package api

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"kikitoru/config"
	"net/http"
)

func GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"config": config.ConvertConfigMatchFrontend(config.C)})
}

func PutConfig(c *gin.Context) {
	var confC struct {
		Config config.StructConfig
	}

	err := c.BindJSON(&confC)
	if err != nil {
		log.Error(err)
		return
	}

	conf := confC.Config
	config.C = config.ConvertFrontendMatchConfig(conf)

	err = config.WriteConfig(config.PathConfig, config.C)
	if err != nil {
		log.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "保存成功",
	})
}

func GetVersion(c *gin.Context) {
	type version struct {
		Current string `json:"current"`
		//LatestStable    string      `json:"latest_stable"`
		//LatestRelease   string      `json:"latest_release"`
		//UpdateAvailable bool        `json:"update_available"`
		//NotifyUser      bool        `json:"notifyUser"`
		//LockFileExists  bool        `json:"lockFileExists"`
		//LockReason      interface{} `json:"lockReason"`
	}
	c.JSON(200, version{Current: config.VERSION})
}

func GetHealth(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, "OK")
}
