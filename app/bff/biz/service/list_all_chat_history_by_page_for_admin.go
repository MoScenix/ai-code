package service

import (
	"context"
	"time"

	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/hertz/pkg/app"
)

type ListAllChatHistoryByPageForAdminService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewListAllChatHistoryByPageForAdminService(Context context.Context, RequestContext *app.RequestContext) *ListAllChatHistoryByPageForAdminService {
	return &ListAllChatHistoryByPageForAdminService{RequestContext: RequestContext, Context: Context}
}

func (h *ListAllChatHistoryByPageForAdminService) Run(req *lapp.ChatHistoryQueryRequest) (resp *lapp.BaseResponsePageChatHistory, err error) {
	q := rpc.AppClient
	res, err := q.ListAppMessage(h.Context, &rpcapp.ListAppMessageReq{
		AppId:          req.AppId,
		PageSize:       20,
		LastCreateTime: time.Now().Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		return &lapp.BaseResponsePageChatHistory{
			Code:    0,
			Message: err.Error(),
		}, err
	}
	resp = &lapp.BaseResponsePageChatHistory{
		Code:    0,
		Message: "success",
		Data: &lapp.PageChatHistory{
			Records: []*lapp.ChatHistory{},
		},
	}
	for _, v := range res.MessageList {
		resp.Data.Records = append(resp.Data.Records, &lapp.ChatHistory{
			Id:          v.Id,
			Message:     v.Content,
			MessageType: v.Role,
			AppId:       v.AppId,
			UserId:      v.UserId,
			CreateTime:  v.CreateTime,
			UpdateTime:  v.UpdateTime,
			IsDelete:    v.IsDelete,
		})
	}
	return resp, nil
}
