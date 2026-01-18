package dal

import (
	"github.com/MoScenix/ai-code/app/ai/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/ai/biz/dal/redis"
)

func Init() {
	redis.Init()
	mysql.Init()
}
