package dal

import (
	"github.com/MoScenix/ai-code/app/bff/biz/dal/redis"
)

func Init() {
	redis.Init()
	//mysql.Init()
}
