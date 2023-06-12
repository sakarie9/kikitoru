package filesystem

import (
	"fmt"
	"golang.org/x/exp/slices"
	"kikitoru/config"
	"kikitoru/internal/database"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const MediaStreamPrefix = "/api/media/stream"
const MediaDownloadPrefix = "/api/media/download"

type FileNode struct {
	Type             string      `json:"type"`
	Title            string      `json:"title"`
	Children         []*FileNode `json:"children,omitempty"`
	Hash             string      `json:"hash,omitempty"`
	WorkTitle        string      `json:"workTitle,omitempty"`
	MediaStreamUrl   string      `json:"mediaStreamUrl,omitempty"`
	MediaDownloadUrl string      `json:"mediaDownloadUrl,omitempty"`
	RealPath         string      `json:"-"`
	//StreamLowQualityUrl string         `json:"streamLowQualityUrl,omitempty"`
	//Duration            float64        `json:"duration,omitempty"`
}

func ScanTracks(rj string) []*FileNode {
	db := database.GetDB()
	var structPath struct {
		RootFolder string `db:"root_folder"`
		Dir        string `db:"dir"`
		Title      string `db:"title"`
	}

	err := db.Get(&structPath, "SELECT root_folder,dir,title FROM t_work WHERE id=$1", rj)
	if err != nil {
		fmt.Println(err)
	}

	idx := slices.IndexFunc(config.C.RootFolders, func(c struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}) bool {
		return c.Name == structPath.RootFolder
	})

	workPath := path.Join(config.C.RootFolders[idx].Path, structPath.Dir)

	var fn FileNode
	var index int
	walk(rj, structPath.Title, workPath, &index, nil, &fn)

	//fmt.Println(fn)

	return fn.Children

}

func walk(id string, title string, pathStr string, index *int, fio os.FileInfo, node *FileNode) {
	// 列出当前目录下的所有目录、文件
	files := listFiles(pathStr)
	files = sortFiles(files, pathStr)

	// 遍历这些文件
	for _, filename := range files {
		// 拼接全路径
		fPath := filepath.Join(pathStr, filename)

		// 构造文件结构
		fio, _ = os.Lstat(fPath)

		var typeFile string
		if fio.IsDir() {
			typeFile = "folder"
		} else {
			typeFile = getExtType(path.Ext(fPath))
		}
		// 为空则不是扩展名可合法文件 跳过
		if typeFile == "" {
			continue
		}

		var child FileNode
		// 处理文件夹
		if typeFile == "folder" {
			child = FileNode{
				typeFile, fio.Name(), []*FileNode{}, "", "", "", "", fPath}
		} else { // 非文件夹
			hash := getHash(id, index)
			child = FileNode{
				typeFile, fio.Name(), []*FileNode{}, hash, title, path.Join(MediaStreamPrefix, hash), path.Join(MediaDownloadPrefix, hash), fPath}
		}

		node.Children = append(node.Children, &child)

		// 如果遍历的当前文件是个目录，则进入该目录进行递归
		if fio.IsDir() {
			walk(id, title, fPath, index, fio, &child)
		}
	}

	return
}

func listFiles(dirname string) []string {
	f, _ := os.Open(dirname)

	names, _ := f.Readdirnames(-1)
	f.Close()

	sort.Strings(names)

	return names
}

func getHash(id string, index *int) string {
	hash := id + "/" + strconv.Itoa(*index)
	*index++
	return hash
}

func getExtType(ext string) string {

	ext = strings.ToLower(ext)
	var t string
	switch ext {
	case ".txt", ".lrc", ".srt", ".ass":
		t = "text"
	case ".jpg", ".jpeg", ".png", ".webp":
		t = "image"
	case ".pdf":
		t = "other"
	case ".mp3", ".ogg", ".opus", ".wav", ".aac", ".flac", ".webm", ".mp4", ".m4a":
		t = "audio"
	default:
		t = ""
	}
	//fmt.Println(ext, t)
	return t
}

func sortTree(tree []*FileNode) []*FileNode {
	sort.SliceStable(tree, func(i, j int) bool {
		if tree[i].Type == "folder" && tree[j].Type != "folder" {
			return true
		} else if tree[i].Type != "folder" && tree[j].Type == "folder" {
			return false
		} else {
			return tree[i].Title < tree[j].Title
		}
	})
	return tree
}

func sortFiles(files []string, pathStr string) []string {
	sort.SliceStable(files, func(i, j int) bool {
		fioI, _ := os.Lstat(filepath.Join(pathStr, files[i]))
		fioJ, _ := os.Lstat(filepath.Join(pathStr, files[j]))
		if fioI.IsDir() && !fioJ.IsDir() {
			return true
		} else if !fioI.IsDir() && fioJ.IsDir() {
			return false
		} else {
			return fioI.Name() < fioJ.Name()
		}
	})
	return files
}
