package service

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/MoScenix/ai-code/app/bff/biz/utils"
	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	"github.com/MoScenix/ai-code/rpc_gen/kitex_gen/ai"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/sse"
)

type ChatToGenCodeService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewChatToGenCodeService(Context context.Context, RequestContext *app.RequestContext) *ChatToGenCodeService {
	return &ChatToGenCodeService{RequestContext: RequestContext, Context: Context}
}

func (h *ChatToGenCodeService) Run(req *lapp.ChatToGenCodeRequest) (resp *lapp.ServerSentEventString, err error) {
	w := sse.NewWriter(h.RequestContext)
	defer w.Close()
	q := rpc.AppClient
	_, err = q.AddMessage(h.Context, &rpcapp.AddMessageReq{
		AppId:   req.AppId,
		UserId:  int64(h.Context.Value(utils.UserIdKey).(float64)),
		Content: req.Message,
		Role:    "user",
	})
	if err != nil {
		return SendErr(w, err)
	}
	var Queryc = ai.AiReq{
		ProjectId: strconv.FormatInt(req.AppId, 10),
	}
	data, err := rpc.AiClient.Chat(utils.WithIdentityMeta(h.Context), &Queryc)
	if err != nil {
		return SendErr(w, err)
	}
	queued := data.GetAnswer() == "true"
	event := "queued"
	message := "true"
	if !queued {
		event = "business-error"
		message = "false"
	}
	Msg, err := json.Marshal(lapp.ServerSentEventString{
		D:       message,
		Message: message,
	})
	if err != nil {
		return SendErr(w, err)
	}
	w.WriteEvent("", event, []byte(Msg))
	w.WriteEvent("", "done", []byte("1"))
	return &lapp.ServerSentEventString{
		Message: "success",
	}, nil
}
func SendErr(w *sse.Writer, err error) (*lapp.ServerSentEventString, error) {
	Msg, err := json.Marshal(lapp.ServerSentEventString{
		Message: err.Error(),
	})
	if err != nil {
		return &lapp.ServerSentEventString{
			Message: "business-error",
		}, err
	}
	w.WriteEvent("", "business-error", Msg)
	return &lapp.ServerSentEventString{
		Message: "business-error",
	}, err
}
