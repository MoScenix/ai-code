package service

import (
	"context"
	"encoding/json"
	"io"
	"strconv"
	"time"

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
	res, err := q.ListAppMessage(h.Context, &rpcapp.ListAppMessageReq{
		AppId:          req.AppId,
		PageSize:       20,
		LastCreateTime: time.Now().Add(20 * time.Second).Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		return SendErr(w, err)
	}
	var Queryc = ai.AiReq{
		ProjectId: strconv.FormatInt(req.AppId, 10),
	}
	for _, v := range res.MessageList {
		Queryc.History = append(Queryc.History, &ai.HistoryItem{
			Question: v.Content,
			Role:     v.Role,
		})
	}
	stream, err := rpc.AiClient.Chat(h.Context, &Queryc)
	defer stream.Close()
	var ans = ""
	for {
		data, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return SendErr(w, err)
		}
		Msg, err := json.Marshal(lapp.ServerSentEventString{
			D: data.Answer,
		})
		w.WriteEvent("", "message", []byte(Msg))
		ans += data.Answer
	}
	w.WriteEvent("", "done", []byte("1"))
	_, err = q.AddMessage(h.Context, &rpcapp.AddMessageReq{
		AppId:   req.AppId,
		UserId:  int64(h.Context.Value(utils.UserIdKey).(float64)),
		Content: ans,
		Role:    "assistant",
	})
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
