package router

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"kikitoru/config"
	"kikitoru/internal"
	"kikitoru/internal/router/api"
	"kikitoru/internal/router/api/media"
	"path"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	if config.C.EnableGzip {
		r.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	log.Error(config.DataDir)
	log.Error(config.C.LogLevel)
	r.Use(static.Serve("/", static.LocalFile(path.Join(config.DataDir, "dist"), false)))
	r.NoRoute(func(c *gin.Context) {
		c.File(path.Join(config.DataDir, "dist"))
	})

	/* ---------------------------  Public routes  --------------------------- */
	public := r.Group("/api")
	public.POST("/auth/me", api.PostAuthMe)
	public.GET("/health", api.GetHealth)

	/* ---------------------------  Private routes  --------------------------- */

	v1 := r.Group("/api")
	v1.Use(internal.JWT())
	{
		v1.GET("/auth/me", api.GetAuthMe)
		v1.GET("/config/admin", api.GetConfig)
		v1.PUT("/config/admin", api.PutConfig)
		v1.GET("/works", api.GetWorks)
		v1.GET("/work/:id", api.GetWork)
		v1.GET("/cover/:id", api.GetCover)
		v1.GET("/version", api.GetVersion)
		v1.GET("/circles", api.GetCircles)
		v1.GET("/circles/:id", api.GetCircleByID)
		v1.GET("/circles/:id/works", api.GetCircleWorksByID)
		v1.GET("/tags", api.GetTags)
		v1.GET("/tags/:id", api.GetTagsByID)
		v1.GET("/tags/:id/works", api.GetTagsWorksByID)
		v1.GET("/vas", api.GetVAs)
		v1.GET("/vas/:id", api.GetVAsByID)
		v1.GET("/vas/:id/works", api.GetVAsWorksByID)
		v1.GET("/tracks/:id", api.GetTracks)
		//v1.PUT("/history", api.PutHistory)
		//v1.GET("/history/recent", api.GetRecent)
		//v1.GET("/history/getByWorkIdIndex", api.GetByWorkIdIndex)
		v1.GET("/review", api.GetReview)
		v1.PUT("/review", api.PutReview)
		v1.DELETE("/review", api.DeleteReview)
		v1.GET("/search/:keyword", api.GetSearch)
		v1.GET("/logs", api.GetLogs)
		v1.GET("/scan", api.StartScan)

		m := v1.Group("/media")
		{
			m.GET("/stream/:id/:index", media.GetStream)
			//m.GET("/download/:id/:index", media.GetDownload)
			m.GET("/download/:id/:index", media.GetStream)
			m.GET("/check-lrc/:id/:index", media.GetCheckLrc)
		}
	}

	return r
}
