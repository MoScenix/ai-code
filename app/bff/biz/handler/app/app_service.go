package app

import (
	"context"
	"fmt"

	"github.com/MoScenix/ai-code/app/bff/biz/service"
	"github.com/MoScenix/ai-code/app/bff/biz/utils"
	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// AddApp .
// @router /app/add [POST]
func AddApp(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.AppAddRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp := &lapp.BaseResponseLong{}
	resp, err = service.NewAddAppService(ctx, c).Run(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// DeleteApp .
// @router /app/delete [POST]
func DeleteApp(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.DeleteRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp := &lapp.BaseResponseBoolean{}
	resp, err = service.NewDeleteAppService(ctx, c).Run(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// UpdateApp .
// @router /app/update [POST]
func UpdateApp(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.AppUpdateRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp := &lapp.BaseResponseBoolean{}
	resp, err = service.NewUpdateAppService(ctx, c).Run(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// GetAppVOById .
// @router /app/get/vo [GET]
func GetAppVOById(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.GetAppVOByIdRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp := &lapp.BaseResponseAppVO{}
	resp, err = service.NewGetAppVOByIdService(ctx, c).Run(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// ListMyAppVOByPage .
// @router /app/my/list/page/vo [POST]
func ListMyAppVOByPage(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.AppQueryRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp := &lapp.BaseResponsePageAppVO{}
	resp, err = service.NewListMyAppVOByPageService(ctx, c).Run(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// ListGoodAppVOByPage .
// @router /app/good/list/page/vo [POST]
func ListGoodAppVOByPage(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.AppQueryRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp := &lapp.BaseResponsePageAppVO{}
	resp, err = service.NewListGoodAppVOByPageService(ctx, c).Run(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// DeployApp .
// @router /app/deploy [POST]
func DeployApp(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.AppDeployRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp := &lapp.BaseResponseString{}
	resp, err = service.NewDeployAppService(ctx, c).Run(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// DownloadAppCode .
// @router /app/download/:appId [GET]
func DownloadAppCode(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.DownloadAppCodeRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp := &lapp.BaseResponseBytes{}
	resp, err = service.NewDownloadAppCodeService(ctx, c).Run(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	c.Data(200, "application/zip", resp.Data)
}

// DeleteAppByAdmin .
// @router /app/admin/delete [POST]
func DeleteAppByAdmin(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.DeleteRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp := &lapp.BaseResponseBoolean{}
	resp, err = service.NewDeleteAppByAdminService(ctx, c).Run(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// UpdateAppByAdmin .
// @router /app/admin/update [POST]
func UpdateAppByAdmin(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.AppAdminUpdateRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp := &lapp.BaseResponseBoolean{}
	resp, err = service.NewUpdateAppByAdminService(ctx, c).Run(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// GetAppVOByIdByAdmin .
// @router /app/admin/get/vo [GET]
func GetAppVOByIdByAdmin(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.GetAppVOByIdByAdminRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp := &lapp.BaseResponseAppVO{}
	resp, err = service.NewGetAppVOByIdByAdminService(ctx, c).Run(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// ListAppVOByPageByAdmin .
// @router /app/admin/list/page/vo [POST]
func ListAppVOByPageByAdmin(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.AppQueryRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp := &lapp.BaseResponsePageAppVO{}
	resp, err = service.NewListAppVOByPageByAdminService(ctx, c).Run(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// ChatToGenCode .
// @router /app/chat/gen/code [GET]
func ChatToGenCode(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.ChatToGenCodeRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp, err := service.NewChatToGenCodeService(ctx, c).Run(&req)

	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// ListAllChatHistoryByPageForAdmin .
// @router /chatHistory/admin/list/page/vo [POST]
func ListAllChatHistoryByPageForAdmin(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.ChatHistoryQueryRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp, err := service.NewListAllChatHistoryByPageForAdminService(ctx, c).Run(&req)

	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}

// ListAppChatHistory .
// @router /chatHistory/app/:appId [GET]
func ListAppChatHistory(ctx context.Context, c *app.RequestContext) {
	var err error
	var req lapp.ListAppChatHistoryRequest
	err = c.BindAndValidate(&req)
	fmt.Printf("bind err=%v, req=%+v\n", err, req)
	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}

	resp, err := service.NewListAppChatHistoryService(ctx, c).Run(&req)

	if err != nil {
		utils.SendErrResponse(ctx, c, consts.StatusOK, err)
		return
	}
	utils.SendSuccessResponse(ctx, c, consts.StatusOK, resp)
}
