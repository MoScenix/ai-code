package utils

import (
	"sync"

	"github.com/MoScenix/ai-code/app/ai/conf"
	"github.com/MoScenix/ai-code/common/clientsuit"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app/appservice"
	"github.com/cloudwego/kitex/client"
)

var (
	appClient     appservice.Client
	appClientOnce sync.Once
	appClientErr  error
)

func AppClient() (appservice.Client, error) {
	appClientOnce.Do(func() {
		appClient, appClientErr = appservice.NewClient("app", newCommonClientOptions(false)...)
	})
	return appClient, appClientErr
}

func newCommonClientOptions(enableGRPC bool) []client.Option {
	opts := clientsuit.CommonGrpcClientSuite{
		CurrentServiceName: conf.GetConf().Kitex.Service,
		RegistryAddr:       conf.GetConf().Registry.RegistryAddress[0],
		EnableGRPC:         enableGRPC,
	}.Options()

	return opts
}
