package dal

import (
	"github.com/MoScenix/ai-code/app/document/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/document/biz/dal/redis"
)

func Init() {
	redis.Init()
	mysql.Init()
}
