package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MoScenix/ai-code/app/bff/biz/utils"
	"github.com/MoScenix/ai-code/app/bff/conf"
	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/pkg/klog"
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
	ctx := utils.WithIdentityMeta(h.Context)
	current, err := rpc.AppClient.GetApp(ctx, &rpcapp.GetAppReq{
		Id: req.AppId,
	})
	if err != nil {
		return &lapp.BaseResponseString{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	if current.GetApp() == nil {
		err = fmt.Errorf("app not found")
		return &lapp.BaseResponseString{
			Code:    1,
			Message: err.Error(),
		}, err
	}

	deployKey := uuid.New().String()
	projectDir := conf.ProjectDir(req.AppId)
	deployDir := conf.DeployDir(deployKey)
	if err = utils.CopyDir(projectDir, deployDir); err != nil {
		klog.CtxErrorf(ctx, "deploy copy failed: app_id=%d deploy_key=%s err=%v", req.AppId, deployKey, err)
		cleanupDeployArtifacts(deployKey)
		return &lapp.BaseResponseString{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	err = utils.ScreenshotViewport(deployDir+"/", conf.CoverPath(deployKey))
	if err != nil {
		klog.CtxErrorf(ctx, "deploy screenshot failed: app_id=%d deploy_key=%s err=%v", req.AppId, deployKey, err)
		cleanupDeployArtifacts(deployKey)
		return &lapp.BaseResponseString{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	_, err = rpc.AppClient.UpdateApp(ctx, &rpcapp.UpdateAppReq{
		Id:        req.AppId,
		DeployKey: deployKey,
		Cover:     conf.CoverURL(deployKey),
	})
	if err != nil {
		klog.CtxErrorf(ctx, "deploy update app failed: app_id=%d deploy_key=%s err=%v", req.AppId, deployKey, err)
		cleanupDeployArtifacts(deployKey)
		return &lapp.BaseResponseString{
			Code:    1,
			Message: err.Error(),
		}, err
	}
	cleanupOldDeployArtifacts(current.GetApp().GetDeployKey(), deployKey)
	klog.CtxInfof(ctx, "app deployed: app_id=%d deploy_key=%s", req.AppId, deployKey)
	return &lapp.BaseResponseString{
		Code:    0,
		Message: "success",
		Data:    conf.DeployURL(deployKey),
	}, nil
}

func cleanupOldDeployArtifacts(oldDeployKey string, newDeployKey string) {
	if oldDeployKey == "" || oldDeployKey == newDeployKey {
		return
	}
	cleanupDeployArtifacts(oldDeployKey)
}

func cleanupDeployArtifacts(deployKey string) {
	if deployKey == "" {
		return
	}
	if _, err := uuid.Parse(deployKey); err != nil {
		return
	}
	removePath(conf.DeployDir(deployKey))
	removePath(filepath.Dir(conf.CoverPath(deployKey)))
}

func removePath(path string) {
	if path == "" || path == "." || path == "/" {
		return
	}
	if err := os.RemoveAll(path); err != nil {
		klog.Warnf("remove deploy artifact failed: err=%v", err)
	}
}
