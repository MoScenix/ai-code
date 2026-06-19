package rpc

import (
	"sync"

	"github.com/MoScenix/ai-code/app/bff/conf"
	"github.com/MoScenix/ai-code/common/clientsuit"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai/aiservice"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app/appservice"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document/documentservice"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user/userservice"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/client"
)

var (
	UserClient     userservice.Client
	AppClient      appservice.Client
	AiClient       aiservice.Client
	DocumentClient documentservice.Client
	once           sync.Once
	once1          sync.Once
	once2          sync.Once
	once3          sync.Once
)

func Init() {
	once.Do(initUserClient)
	once1.Do(initAppClient)
	once2.Do(initAiClient)
	once3.Do(initDocumentClient)
}
func initUserClient() {
	opts := newCommonClientOptions(false)
	var err error
	UserClient, err = userservice.NewClient(
		"user",
		opts...,
	)
	if err != nil {
		hlog.Fatal(err)
	}
}
func initAppClient() {
	opts := newCommonClientOptions(false)
	var err error
	AppClient, err = appservice.NewClient(
		"app",
		opts...,
	)
	if err != nil {
		hlog.Fatal(err)
	}
}
func initAiClient() {
	opts := newCommonClientOptions(true)
	var err error
	AiClient, err = aiservice.NewClient(
		"ai",
		opts...,
	)
	if err != nil {
		hlog.Fatal(err)
	}
}
func initDocumentClient() {
	opts := newCommonClientOptions(false)
	var err error
	DocumentClient, err = documentservice.NewClient(
		"document",
		opts...,
	)
	if err != nil {
		hlog.Fatal(err)
	}
}

func newCommonClientOptions(enableGRPC bool) []client.Option {
	opts := clientsuit.CommonGrpcClientSuite{
		CurrentServiceName: conf.GetConf().Hertz.Service,
		RegistryAddr:       conf.GetConf().Consul.Address,
		EnableGRPC:         enableGRPC,
	}.Options()

	return opts
}
