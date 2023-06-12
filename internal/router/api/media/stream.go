package media

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"kikitoru/internal/database"
	"kikitoru/internal/filesystem"
	"math"
	"strconv"
	"strings"
)

func GetStream(c *gin.Context) {
	id := c.Param("id")
	i := c.Param("index")
	index, err := strconv.Atoi(i)
	if err != nil {
		log.Warn(err)
	}

	db := database.GetDB()
	var streamPath string
	_ = db.Get(&streamPath, "SELECT path FROM t_file_nodes WHERE id=$1 and index=$2 LIMIT 1", id, index)

	if streamPath == "" {
		streamPath = updateFileNodesAndGetPath(db, id, index)
	}

	c.File(streamPath)
}

// 遍历Nodes然后返回跟hash相同的node的path并更新数据库
func updateFileNodesAndGetPath(db *sqlx.DB, rj string, index int) string {
	var fileIndex []FileIndex // hash映射缓存
	hash := fmt.Sprintf("%s/%d", rj, index)
	streamPath := FindMediaStreamUrlByHashInSlice(&fileIndex, filesystem.ScanTracks(rj), hash)
	// 更新数据库
	_, err := db.NamedExec(`INSERT INTO t_file_nodes (id, index, path) VALUES (:id,:index,:path)
		ON CONFLICT (id,index) DO UPDATE SET path=excluded.path`, fileIndex)
	if err != nil {
		log.Warn(err)
	}
	return streamPath
}

// 遍历Nodes并更新数据库
func updateFileNodesDB(db *sqlx.DB, rj string) {
	updateFileNodesAndGetPath(db, rj, math.MaxInt)
}

type FileIndex struct {
	ID    string `db:"id"`
	Index int    `db:"index"`
	Path  string `db:"path"`
}

func FindMediaStreamUrlByHashInSlice(indexes *[]FileIndex, nodes []*filesystem.FileNode, hash string) string {
	for _, node := range nodes {
		result := findMediaStreamUrlByHashRecursive(indexes, node, hash)
		if result != "" {
			return result
		}
	}
	return ""
}

func findMediaStreamUrlByHashRecursive(indexes *[]FileIndex, node *filesystem.FileNode, hash string) string {
	//if node.Type == "text" || node.Type == "audio" {
	if node.Type != "folder" {
		// 遍历压入 index
		id, index := splitHash(node.Hash)
		*indexes = append(*indexes, FileIndex{ID: id, Index: index, Path: node.RealPath})
	}

	// 检查当前节点是否具有指定的 Hash 值
	if node.Hash == hash {
		return node.RealPath
	}

	// 递归遍历子节点
	for _, child := range node.Children {
		result := findMediaStreamUrlByHashRecursive(indexes, child, hash)
		if result != "" {
			return result
		}
	}

	// 如果没有找到匹配的节点，返回空字符串
	return ""
}

func splitHash(hash string) (string, int) {
	parts := strings.Split(hash, "/")
	left := parts[0]
	right := parts[1]
	rightInt, _ := strconv.Atoi(right)
	return left, rightInt
}
