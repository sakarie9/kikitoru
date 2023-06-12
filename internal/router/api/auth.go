package api

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"kikitoru/config"
	"kikitoru/internal"
	"kikitoru/internal/database"
	"kikitoru/util"
	"net/http"
)

/*
PostAuthMe
{"name":"admin","password":"admin"}
{"token":"jwt"}
*/
func PostAuthMe(c *gin.Context) {
	db := database.GetDB()
	type structUser struct {
		Name     string
		Password string
		Group    string
	}

	var userPost structUser
	err := c.BindJSON(&userPost)
	if err != nil {
		log.Error(err)
		return
	}

	//logs.Debug(userPost)
	var userDB structUser
	err = db.Get(&userDB, `SELECT name,password,"group" FROM t_user WHERE name=$1`, userPost.Name)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debug(util.MD5(userPost.Password + config.C.MD5Secret))
	if userDB.Password != util.MD5(userPost.Password+config.C.MD5Secret) {
		log.Error(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误."})
		return
	}

	token, err := internal.GenerateToken(userDB.Name, userDB.Group)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

/*
GetAuthMe
{"user":{"name":"admin","group":"administrator"},"auth":true}
*/
func GetAuthMe(c *gin.Context) {
	type User struct {
		Name  string `json:"name"`
		Group string `json:"group"`
	}

	type structAuth struct {
		User User `json:"user"`
		Auth bool `json:"auth"`
	}

	c.JSON(http.StatusOK, structAuth{User: User{
		Name:  c.GetString("username"),
		Group: c.GetString("usergroup"),
	}, Auth: true})

}
