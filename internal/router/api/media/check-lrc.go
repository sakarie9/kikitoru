package media

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"kikitoru/internal/database"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

/*
GetCheckLrc
/api/media/check-lrc/RJ01018155/16
{"result":true,"message":"找到歌词文件","hash":"RJ1018155/0"}
*/
func GetCheckLrc(c *gin.Context) {
	rj := c.Param("id")
	i := c.Param("index")
	index, err := strconv.Atoi(i)
	if err != nil {
		log.Warn(err)
	}

	db := database.GetDB()
	var lrcIndex int
	var audioPath string
	_ = db.Select(&audioPath, "SELECT path FROM t_file_nodes WHERE id=$1 and index=$2 LIMIT 1", rj, index)

	if audioPath == "" {
		//audioPath = updateFileNodesAndGetPath(db, rj, index)
		updateFileNodesDB(db, rj)
	}

	fileName := filepath.Base(audioPath)
	ext := path.Ext(audioPath)
	fileNameLrc := fmt.Sprintf("%%%s.lrc", strings.TrimSuffix(fileName, ext))
	err = db.Select(&lrcIndex, "SELECT index FROM t_file_nodes WHERE id=$1 and path LIKE $2 LIMIT 1", rj, fileNameLrc)
	if err != nil {
		log.Warn(err)
	}

	hash := fmt.Sprintf("%s/%d", rj, lrcIndex)
	c.JSON(http.StatusOK, gin.H{"result": true, "message": "找到歌词文件", "hash": hash})
}
