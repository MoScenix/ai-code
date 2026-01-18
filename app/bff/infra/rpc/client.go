package rpc

import (
	"sync"

	"github.com/MoScenix/ai-code/app/bff/conf"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user/userservice"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
)

var (
	UserClient userservice.Client
	once       sync.Once
)

func Init() {
	once.Do(initUserClient)
}
func initUserClient() {
	r, err := consul.NewConsulResolver(conf.GetConf().Consul.Address)
	if err != nil {
		hlog.Fatal(err)
	}

	UserClient, err = userservice.NewClient(
		"user",
		client.WithResolver(r),
	)
	if err != nil {
		hlog.Fatal(err)
	}
}
