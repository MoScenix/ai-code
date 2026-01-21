package service

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/MoScenix/ai-code/app/app/biz/dal/mysql"
	"github.com/MoScenix/ai-code/app/app/biz/model"
	"github.com/MoScenix/ai-code/app/app/conf"
	app "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
)

type AddAppService struct {
	ctx context.Context
} // NewAddAppService new AddAppService
func NewAddAppService(ctx context.Context) *AddAppService {
	return &AddAppService{ctx: ctx}
}

// Run create note info
func (s *AddAppService) Run(req *app.AddAppReq) (resp *app.AddAppResp, err error) {
	// Finish your business logic.
	rs := []rune(req.InitPrompt)
	res, err := model.NewAppQuery(s.ctx, mysql.DB).CreateApp(model.App{
		Name:       string(rs[:min(len(rs), 12)]),
		InitPrompt: req.InitPrompt,
		UserId:     uint(req.UserId),
		Priority:   1,
	})
	path := filepath.Join(conf.GetConf().ShareDir.ShareDir, strconv.FormatInt(int64(res.ID), 10))
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &app.AddAppResp{
		Id: int64(res.ID),
	}, err
}
