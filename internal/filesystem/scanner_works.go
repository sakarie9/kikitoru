package filesystem

import "C"
import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	log "github.com/sirupsen/logrus"
	"kikitoru/config"
	"kikitoru/internal/scraper"
	"kikitoru/logs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// WalkFun 回调
type WalkFun func(dir, name string, isDir bool) error

func WalkDepth(name string, maxDepth int, fn WalkFun) error {
	return walkDirs([]string{name}, 1, maxDepth, fn)
}

// 遍历每个dirs下的文件，并且返回所有dirs下是目录的文件
func walkDirs(dirs []string, dep, maxDepth int, fn WalkFun) error {
	var subDirs []string
	for _, d := range dirs {
		ff, err := os.ReadDir(d)
		if err != nil {
			return err
		}
		for _, f := range ff {
			err = fn(d, f.Name(), f.IsDir())
			if err != nil {
				return err
			}
			if f.IsDir() { // 如果是目录
				subDirs = append(subDirs, filepath.Join(d, f.Name()))
			}
		}
	}
	if len(subDirs) > 0 && dep < maxDepth {
		return walkDirs(subDirs, dep+1, maxDepth, fn)
	}
	return nil
}

func IsLrcExistsInPath(path string) bool {
	found := false

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".lrc" {
			log.Info("找到 lrc 文件:", path)
			found = true
			return filepath.SkipDir // 停止遍历子目录
		}
		return nil
	})
	if err != nil {
		log.Warn("遍历目录出错:", err)
	}
	return found
}

// ScanFileSystem 扫描文件系统获取RJ号
func ScanFileSystem() []scraper.ScrapedWorkMetadata {

	var scannedWorks []scraper.ScrapedWorkMetadata

	for _, v := range config.C.RootFolders {
		root := v.Name
		path := strings.TrimRight(v.Path, "/")
		err := WalkDepth(path, config.C.ScannerMaxRecursionDepth, func(dir, name string, isDir bool) error {
			if !isDir {
				return nil
			}
			pattern := `[Rr][Jj](\d+)`
			re := regexp.MustCompile(pattern)
			match := re.FindStringSubmatch(name)

			shortDir := (dir + "/" + name)[len(path)+1:]
			// 如果匹配到含有RJ号的目录
			if len(match) > 1 {
				log.Infof("扫描到 %s", dir+"/"+name)
				isLrc := IsLrcExistsInPath(dir + "/" + name)
				work := scraper.ScrapedWorkMetadata{ID: match[0], RootFolder: root, Dir: shortDir, Lrc: isLrc}
				scannedWorks = append(scannedWorks, work)
			}
			return nil
		})
		if err != nil {
			log.Warn(err)
		}
	}

	log.Infof("共找到 %d 个音声文件夹", len(scannedWorks))
	logs.ScanLogs.Total = len(scannedWorks)
	logs.ScanLogs.Details.Enqueue(fmt.Sprintf("共找到 %d 个音声文件夹", len(scannedWorks)))

	return scannedWorks
}

// StartScraper 爬取 DLsite 获取作品信息
func StartScraper() []scraper.ScrapedWorkMetadata {
	scannedWorks := ScanFileSystem()

	logs.ScanLogs.Details.Enqueue("开始爬取作品元数据并下载封面")

	p, _ := ants.NewPool(config.C.MaxParallelism)
	defer p.Release()

	// Use the common pool.
	var wg sync.WaitGroup

	// 获取将要写入数据库的数据
	newScannedWorks := sync.Map{}
	for i, v := range scannedWorks {
		tmpV := v
		logs.ScanLogs.Position = i + 1
		logs.ScanLogs.MainLog.Enqueue(fmt.Sprintf("======== 正在扫描（%d/%d） ========", logs.ScanLogs.Position, logs.ScanLogs.Total))
		wg.Add(1)
		_ = p.Submit(func() {
			log.Infof("开始获取 [%s] 的元数据", tmpV.ID)
			// 爬取音声元数据
			sWork := scraper.GetScrapedWork(tmpV)
			if sWork.Title != "" {
				newScannedWorks.Store(tmpV.ID, sWork)
				// 获取封面
				scraper.GetCoverImage(tmpV.ID)
			} else {
				log.Errorf("%s: 标题不存在，刮削失败", sWork.ID)
			}
			wg.Done()
		})
	}
	wg.Wait()

	var sliceWorks []scraper.ScrapedWorkMetadata
	newScannedWorks.Range(func(key, value interface{}) bool {
		sliceWorks = append(sliceWorks, value.(scraper.ScrapedWorkMetadata))
		return true
	})
	return sliceWorks
}

// 扫描入口 扫描、刮削、写入数据库
func StartScanWorks() {

	logs.ScanLogs.MainLog.Enqueue("======== 开始扫描 ========")
	logs.ScanLogs.State = "running"

	start := time.Now()
	var workCount int
	// 开始运行
	works := StartScraper()
	workCount = SaveWorks(works)
	if workCount == 0 {
		log.Error("扫描失败")
		logs.ScanLogs.MainLog.Enqueue("======== 扫描失败 ========")
		logs.ScanLogs.State = "error"
		return
	}

	duration := time.Since(start)
	log.Infof("扫描完成，耗时 %s，共添加 %d 部作品", duration, workCount)
	logs.ScanLogs.MainLog.Enqueue(fmt.Sprintf("扫描完成 耗时 %s 共添加 %d 部作品", duration, workCount))
	logs.ScanLogs.State = "finished"
	logs.ScanLogs.Details.Enqueue(fmt.Sprint("扫描完成"))
}

func cleanCovers(worksID []string) {
	log.Warn("开始清理封面 " + strings.Join(worksID, ","))
	var count int
	coverDir := config.C.CoverFolderDir
	files, err := os.ReadDir(coverDir)
	if err != nil {
		log.Error(err)
		return
	}
	subDirectory := "excluded_subdirectory"
	err = os.Mkdir(filepath.Join(coverDir, subDirectory), 0750)
	if err != nil {
		log.Error(err)
		return
	}

	// 将id和文件名的并集放入子文件夹
	for _, file := range files {
		fileName := file.Name()
		for _, rj := range worksID {
			if strings.HasPrefix(fileName, rj) {
				filePath := filepath.Join(coverDir, fileName)
				filePathNew := filepath.Join(coverDir, subDirectory, fileName)
				//err := os.Remove(filePath)
				err := os.Rename(filePath, filePathNew)
				if err != nil {
					log.Error(err)
				}
			}
		}
	}
	// 删除原文件夹下的所有封面
	for _, file := range files {
		fileName := file.Name()
		if fileName != subDirectory {
			filePath := filepath.Join(coverDir, fileName)
			err := os.Remove(filePath)
			count++
			if err != nil {
				log.Error(err)
			}
		}
	}
	// 将子文件夹的封面移回原位
	files, err = os.ReadDir(filepath.Join(coverDir, subDirectory))
	if err != nil {
		log.Error(err)
		return
	}
	for _, file := range files {
		fileName := file.Name()
		newPath := filepath.Join(coverDir, fileName)
		oldPath := filepath.Join(coverDir, subDirectory, fileName)
		err := os.Rename(oldPath, newPath)
		if err != nil {
			log.Error(err)
		}
	}
	err = os.Remove(filepath.Join(coverDir, subDirectory))
	if err != nil {
		log.Error(err)
	}

	if err == nil {
		log.Warnf("完成封面清理 共清理 %d 个封面", count)
	}

}
