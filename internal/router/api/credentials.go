package api

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"kikitoru/config"
	"kikitoru/internal/database"
	"kikitoru/util"
	"net/http"
)

type structUsers struct {
	Name  string `json:"name" db:"name"`
	Group string `json:"group" db:"group"`
}

func GetUsers(c *gin.Context) {

	var users []structUsers
	db := database.GetDB()
	err := db.Select(&users, "SELECT name,\"group\" FROM t_user")
	if err != nil {
		log.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func PostNewUser(c *gin.Context) {
	type structNewUser struct {
		Name     string `json:"name" db:"name"`
		Group    string `json:"group" db:"group"`
		Password string `json:"password" db:"password"`
	}
	var newUser structNewUser
	err := c.BindJSON(&newUser)
	if err != nil {
		log.Warn(err)
		return
	}

	db := database.GetDB()
	_, err = db.Exec(`INSERT INTO t_user (name, password, "group") VALUES ($1,$2,$3) ON CONFLICT (name) DO NOTHING`,
		newUser.Name, util.MD5(newUser.Password+config.C.MD5Secret), newUser.Group)
	if err != nil {
		log.Warn(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户 " + newUser.Name + " 创建成功."})
}

func DeleteUsers(c *gin.Context) {
	type structDeletePayload struct {
		Users []structUsers
	}
	var deletePayload structDeletePayload
	err := c.BindJSON(&deletePayload)
	if err != nil {
		log.Warn(err)
		return
	}

	db := database.GetDB()
	for _, v := range deletePayload.Users {
		_, err := db.Exec(`DELETE FROM t_user WHERE name=$1`, v.Name)
		if err != nil {
			log.Warn(err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功."})
}

func PutNewAdminPassword(c *gin.Context) {
	type structNewPassword struct {
		Name     string `json:"name"`
		Password string `json:"newPassword"`
	}
	var newPwd structNewPassword
	err := c.BindJSON(&newPwd)
	if err != nil {
		log.Warn(err)
		return
	}

	db := database.GetDB()
	_, err = db.Exec(`UPDATE t_user SET password=$1 WHERE name=$2`, util.MD5(newPwd.Password+config.C.MD5Secret), newPwd.Name)
	if err != nil {
		log.Warn(err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功."})
	c.Redirect(302, "/")
}
