package service

import (
	"context"
	"strconv"

	"github.com/MoScenix/ai-code/app/bff/biz/utils"
	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
)

type DeployAppService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewDeployAppService(Context context.Context, RequestContext *app.RequestContext) *DeployAppService {
	return &DeployAppService{RequestContext: RequestContext, Context: Context}
}

func (h *DeployAppService) Run(req *lapp.AppDeployRequest) (resp *lapp.BaseResponseString, err error) {
	deployKey := uuid.New().String()
	utils.CopyDir("/static/project/"+strconv.FormatInt(req.AppId, 10), "/deploy/"+deployKey)
	err = utils.ScreenshotViewport("/static/project/"+strconv.FormatInt(req.AppId, 10)+"/", "/static/cover/"+deployKey+"/deploy.png")
	if err != nil {
		return &lapp.BaseResponseString{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	_, err = rpc.AppClient.UpdateApp(h.Context, &rpcapp.UpdateAppReq{
		Id:        req.AppId,
		DeployKey: deployKey,
		Cover:     "/static/cover/" + deployKey + "/deploy.png",
	})
	if err != nil {
		return &lapp.BaseResponseString{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	return &lapp.BaseResponseString{
		Code:    0,
		Message: "success",
		Data:    "/deploy/" + deployKey + "/",
	}, nil
}
