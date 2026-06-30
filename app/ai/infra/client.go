package infra

import (
	"sync"

	"github.com/MoScenix/ai-code/app/ai/conf"
	"github.com/MoScenix/ai-code/common/clientsuit"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app/appservice"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document/documentservice"
	"github.com/cloudwego/kitex/client"
)

var (
	appClient          appservice.Client
	appClientOnce      sync.Once
	appClientErr       error
	documentClient     documentservice.Client
	documentClientOnce sync.Once
	documentClientErr  error
)

func AppClient() (appservice.Client, error) {
	appClientOnce.Do(func() {
		appClient, appClientErr = appservice.NewClient("app", newCommonClientOptions(false)...)
	})
	return appClient, appClientErr
}

func DocumentClient() (documentservice.Client, error) {
	documentClientOnce.Do(func() {
		documentClient, documentClientErr = documentservice.NewClient("document", newCommonClientOptions(false)...)
	})
	return documentClient, documentClientErr
}

func newCommonClientOptions(enableGRPC bool) []client.Option {
	return clientsuit.CommonGrpcClientSuite{
		CurrentServiceName: conf.GetConf().Kitex.Service,
		RegistryAddr:       conf.GetConf().Registry.RegistryAddress[0],
		EnableGRPC:         enableGRPC,
	}.Options()
}
