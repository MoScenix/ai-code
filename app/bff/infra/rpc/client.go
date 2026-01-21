package rpc

import (
	"sync"

	"github.com/MoScenix/ai-code/app/bff/conf"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai/aiservice"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app/appservice"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user/userservice"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
)

var (
	UserClient userservice.Client
	AppClient  appservice.Client
	AiClient   aiservice.Client
	once       sync.Once
	once1      sync.Once
	once2      sync.Once
)

func Init() {
	once.Do(initUserClient)
	once1.Do(initAppClient)
	once2.Do(initAiClient)
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
func initAppClient() {
	r, err := consul.NewConsulResolver(conf.GetConf().Consul.Address)
	if err != nil {
		hlog.Fatal(err)
	}

	AppClient, err = appservice.NewClient(
		"app",
		client.WithResolver(r),
	)
	if err != nil {
		hlog.Fatal(err)
	}
}
func initAiClient() {
	r, err := consul.NewConsulResolver(conf.GetConf().Consul.Address)
	if err != nil {
		hlog.Fatal(err)
	}
	AiClient, err = aiservice.NewClient(
		"ai",
		client.WithResolver(r),
	)
	if err != nil {
		hlog.Fatal(err)
	}
}
