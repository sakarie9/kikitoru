package filesystem

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"kikitoru/config"
	"kikitoru/internal/database"
	"kikitoru/internal/scraper"
	"kikitoru/logs"
	"kikitoru/util"
	"strings"
)

func SaveWorks(works []scraper.ScrapedWorkMetadata) int {
	db := database.GetDB()

	tx, err := db.Beginx()
	if err != nil {
		log.Error(err)
		return 0
	}

	// 删除旧的未扫描到但存在于数据库中的作品
	nonExistWorks := CompareScannedWorksWithDB(db, works)
	//fmt.Println(nonExistWorks)
	deleteWorks(db, nonExistWorks)

	// 清理封面
	if config.C.SkipCleanup {
		cleanCovers(nonExistWorks)
	}

	for i, v := range works {
		logs.ScanLogs.Position = i + 1
		logs.ScanLogs.MainLog.Enqueue(fmt.Sprintf("======== 插入数据库（%d/%d） ========", logs.ScanLogs.Position, logs.ScanLogs.Total))
		logs.ScanLogs.Details.Enqueue(fmt.Sprintf("%s: 正在插入数据库", v.ID))
		saveWork(tx, v)
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
		return 0
	}
	log.Info("成功写入数据库")
	return len(works)
}

func CompareScannedWorksWithDB(db *sqlx.DB, works []scraper.ScrapedWorkMetadata) []string {

	type structDBWorks struct {
		ID string `db:"id"`
	}
	var dbWorks []structDBWorks
	err := db.Select(&dbWorks, "SELECT id FROM t_work")
	if err != nil {
		log.Error(err)
	}

	dbWorksSlice := make([]string, len(dbWorks))
	for i, s := range dbWorks {
		dbWorksSlice[i] = s.ID
	}

	// 预先分配 id 切片的容量
	scWorks := make([]string, 0, len(works))
	// 迭代 test1 中的每个元素
	for _, t := range works {
		// 将每个元素的 id 添加到切片中
		scWorks = append(scWorks, t.ID)
	}

	return util.DifferenceSlice(dbWorksSlice, scWorks)
}

func saveWork(tx *sqlx.Tx, work scraper.ScrapedWorkMetadata) {

	// Insert to t_work
	sqlInsertWork := `
	INSERT INTO t_work (id,root_folder,dir,title,nsfw,release,dl_count,price,review_count,rate_count,rate_average_2dp,rate_count_detail,rank,has_subtitle)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	ON CONFLICT (id) 
	DO UPDATE SET (root_folder,dir,title,nsfw,release,dl_count,price,review_count,rate_count,rate_average_2dp,rate_count_detail,rank,has_subtitle)
	   = (excluded.root_folder,excluded.dir,excluded.title,excluded.nsfw,excluded.release,excluded.dl_count,excluded.price,excluded.review_count,excluded.rate_count,excluded.rate_average_2dp,excluded.rate_count_detail,excluded.rank,excluded.has_subtitle)
`
	_, err := tx.Exec(
		sqlInsertWork,
		work.ID,
		work.RootFolder,
		work.Dir,
		work.Title,
		work.Nsfw,
		work.Release,
		work.DLCount,
		work.Price,
		work.ReviewCount,
		work.RateCount,
		work.RateAverage2dp,
		work.RateCountDetail,
		work.Rank,
		work.Lrc,
	)
	if err != nil {
		log.Errorf("%s: %s", work.ID, err)
	}

	// Insert to t_circle
	sqlInsertCircle := `
	INSERT INTO t_circle (id,name)
	VALUES (:id,:name)
	ON CONFLICT (id) 
	DO UPDATE SET name = EXCLUDED.name
`
	_, err = tx.NamedExec(
		sqlInsertCircle,
		work.Circle,
	)
	if err != nil {
		log.Errorf("%s: %s", work.ID, err)
	}

	// Insert to r_circle_work
	sqlInsertCircleWork := `
	INSERT INTO r_circle_work (circle_id,work_id)
	VALUES ($1,$2)
	ON CONFLICT DO NOTHING
`
	_, err = tx.Exec(
		sqlInsertCircleWork,
		work.Circle.ID,
		work.ID,
	)
	if err != nil {
		log.Errorf("%s: %s", work.ID, err)
	}

	// Insert to t_va
	sqlInsertVA := `
	INSERT INTO t_va (id,name)
	VALUES (:id,:name)
	ON CONFLICT (id) 
	DO UPDATE SET name = EXCLUDED.name
`
	_, err = tx.NamedExec(
		sqlInsertVA,
		work.Vas,
	)
	if err != nil {
		log.Errorf("%s: %s", work.ID, err)
	}

	// Insert to r_va_work
	sqlInsertVAWork := `
	INSERT INTO r_va_work (va_id,work_id)
	VALUES ($1,$2)
	ON CONFLICT DO NOTHING
`
	for _, v := range work.Vas {
		_, err = tx.Exec(
			sqlInsertVAWork,
			v.ID,
			work.ID,
		)
		if err != nil {
			log.Errorf("%s: %s", work.ID, err)
		}
	}

	if len(work.Tags) != 0 {

		// Insert to t_tag
		sqlInsertTags := `
	INSERT INTO t_tag (id,name)
	VALUES (:id,:name)
	ON CONFLICT (id) 
	DO UPDATE SET name = EXCLUDED.name
`
		_, err = tx.NamedExec(
			sqlInsertTags,
			work.Tags,
		)
		if err != nil {
			log.Errorf("%s: %s", work.ID, err)
		}

		// Insert to r_tag_work
		sqlInsertTagWork := `
	INSERT INTO r_tag_work (tag_id,work_id)
	VALUES ($1,$2)
	ON CONFLICT DO NOTHING
`
		for _, v := range work.Tags {
			_, err = tx.Exec(
				sqlInsertTagWork,
				v.ID,
				work.ID,
			)
			if err != nil {
				log.Errorf("%s: %s", work.ID, err)
			}
		}
	} else {
		log.Warnf("%s: Tag 不存在", work.ID)
	}

	if work.Series.ID != "" {
		// Insert to t_series
		sqlInsertSeries := `
		INSERT INTO t_series (id,name)
		VALUES (:id,:name)
		ON CONFLICT (id) 
	    DO UPDATE SET name = EXCLUDED.name
		`
		_, err = tx.NamedExec(
			sqlInsertSeries,
			work.Series,
		)
		if err != nil {
			log.Errorf("%s: %s", work.ID, err)
		}

		// Insert to r_series_work
		sqlInsertSeriesWork := `
		INSERT INTO r_series_work (series_id,work_id)
		VALUES ($1,$2)
		ON CONFLICT DO NOTHING
		`
		_, err = tx.Exec(
			sqlInsertSeriesWork,
			work.Series.ID,
			work.ID,
		)
		if err != nil {
			log.Errorf("%s: %s", work.ID, err)
		}
	} else {
		log.Warnf("%s: 系列不存在", work.ID)
	}
}

func deleteWorks(db *sqlx.DB, worksID []string) {
	log.Warn("开始清理数据库 " + strings.Join(worksID, ","))
	tx, err := db.Beginx()
	if err != nil {
		log.Error(err)
	}
	for _, v := range worksID {
		_, err := tx.Exec("DELETE FROM t_work WHERE id=$1", v)
		if err != nil {
			log.Error(err)
			tx.Rollback()
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}
	log.Warnf("完成清理数据库 共清理 %d 条数据", len(worksID))
}
