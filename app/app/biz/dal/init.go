package dal

import (
	"github.com/MoScenix/ai-code/app/app/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/app/biz/dal/redis"
)

func Init() {
	redis.Init()
	mysql.Init()
}
