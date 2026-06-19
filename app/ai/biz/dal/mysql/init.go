package mysql

import (
	"fmt"
	"os"
	"strings"

	"github.com/MoScenix/ai-code/app/ai/conf"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	err error
)

func Init() {
	dsn := conf.GetConf().MySQL.DSN
	if strings.Contains(dsn, "%s") {
		dsn = fmt.Sprintf(dsn, os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_DATABASE"))
	}
	DB, err = gorm.Open(mysql.Open(dsn),
		&gorm.Config{
			PrepareStmt:            true,
			SkipDefaultTransaction: true,
		},
	)
	if err != nil {
		panic(err)
	}
}
