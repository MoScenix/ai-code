package rpc

import (
	"context"
	"strconv"
	"sync"

	"github.com/MoScenix/ai-code/app/bff/conf"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai/aiservice"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app/appservice"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/user/userservice"
	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/transmeta"
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

const (
	ctxUserIDKey   = "user_id"
	ctxUserRoleKey = "user_role"

	metaUserIDKey   = "x-user-id"
	metaUserRoleKey = "x-user-role"
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
		client.WithMetaHandler(transmeta.MetainfoClientHandler),
		client.WithMiddleware(injectIdentityMetaMiddleware()),
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
		client.WithMetaHandler(transmeta.MetainfoClientHandler),
		client.WithMiddleware(injectIdentityMetaMiddleware()),
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
		client.WithMetaHandler(transmeta.MetainfoClientHandler),
		client.WithMiddleware(injectIdentityMetaMiddleware()),
	)
	if err != nil {
		hlog.Fatal(err)
	}
}

func injectIdentityMetaMiddleware() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) error {
			if userID, ok := parseUserID(ctx.Value(ctxUserIDKey)); ok {
				ctx = metainfo.WithPersistentValue(ctx, metaUserIDKey, strconv.FormatInt(userID, 10))
			}
			if userRole, ok := ctx.Value(ctxUserRoleKey).(string); ok && userRole != "" {
				ctx = metainfo.WithPersistentValue(ctx, metaUserRoleKey, userRole)
			}
			return next(ctx, req, resp)
		}
	}
}

func parseUserID(v interface{}) (int64, bool) {
	switch value := v.(type) {
	case int64:
		return value, true
	case int32:
		return int64(value), true
	case int:
		return int64(value), true
	case float64:
		return int64(value), true
	case float32:
		return int64(value), true
	case uint64:
		return int64(value), true
	case uint32:
		return int64(value), true
	case uint:
		return int64(value), true
	default:
		return 0, false
	}
}
