package service

import (
	"context"

	docutils "github.com/MoScenix/ai-code/app/document/utils"
	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"
)

type DeleteProjectFileDataService struct {
	ctx context.Context
} // NewDeleteProjectFileDataService new DeleteProjectFileDataService
func NewDeleteProjectFileDataService(ctx context.Context) *DeleteProjectFileDataService {
	return &DeleteProjectFileDataService{ctx: ctx}
}

// Run create note info
func (s *DeleteProjectFileDataService) Run(req *document.DeleteProjectFileDataReq) (resp *document.DeleteProjectFileDataResp, err error) {
	if err := docutils.DeleteProjectData(s.ctx, req.ProjectId); err != nil {
		return nil, err
	}
	return &document.DeleteProjectFileDataResp{
		Success: true,
	}, nil
}
