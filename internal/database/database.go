package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"kikitoru/config"
	"kikitoru/util"
	"os"
)

var DB *sqlx.DB

// var DB *gorm.DB

const connStr = "postgres://postgres:114514@localhost/postgres?sslmode=disable"

func InitDataBase() {

	// 优先获取环境变量的数据库链接
	dbURL := os.Getenv("KIKITORU_DATABASE_URL")
	if dbURL == "" {
		dbURL = config.C.DatabaseURL
	}

	var err error
	DB, err = sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatal("DB open error: ", err)
	}

	DB.MustExec(Schema)

	md5 := util.MD5("admin" + config.C.MD5Secret)
	log.Debug(md5)
	_, err = DB.Exec(`insert into t_user (name, password, "group")
	values ('admin',$1,'administrator')
	on conflict do nothing;`, md5)
	if err != nil {
		log.Error("Unable to create admin account")
	}

}

func GetDB() *sqlx.DB {
	return DB
}
