package mysql

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"web_app/settings"
)

var(
	db *sqlx.DB
)

func Init(cfg *settings.MySQLConfig) (err error) {
	dsn:=fmt.Sprintf("%s:%s@tcp(%s:%v)/%s",cfg.User,cfg.Password,cfg.Host,cfg.Port,cfg.DbName)

	db, err = sqlx.Connect("mysql", dsn) //此处的connect相当于两个步骤  Open ping
	if err != nil {
		zap.L().Error("connect db failed",zap.Error(err))
		return err
	}


	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	return nil
}

func Close()  {
	db.Close()
}
