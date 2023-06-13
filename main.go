package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"kikitoru/config"
	"kikitoru/internal/database"
	"kikitoru/internal/router"
	"kikitoru/logs"
)

func main() {
	fmt.Printf("======== Kikitoru %s ========", config.VERSION)
	logs.InitLogger()
	config.InitConfig()
	logs.InitLogger()
	database.InitDataBase()

	if config.C.Production {
		gin.SetMode(gin.ReleaseMode)
	}
	r := router.InitRouter()

	var addr string
	if config.C.BlockRemoteConnection {
		addr = fmt.Sprintf("localhost:%d", config.C.ListenPort)
	} else {
		addr = fmt.Sprintf(":%d", config.C.ListenPort)
	}

	err := r.Run(addr)
	if err != nil {
		log.Fatal(err)
	}

}
