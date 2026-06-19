package service

import (
	"context"

	docutils "github.com/MoScenix/ai-code/app/document/utils"
	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"
)

type IndexTextFileService struct {
	ctx context.Context
} // NewIndexTextFileService new IndexTextFileService
func NewIndexTextFileService(ctx context.Context) *IndexTextFileService {
	return &IndexTextFileService{ctx: ctx}
}

// Run create note info
func (s *IndexTextFileService) Run(req *document.IndexTextFileReq) (resp *document.IndexTextFileResp, err error) {
	dir := projectFileDir(req.ProjectId, req.FileId)
	textPath, err := findFileByExt(dir, ".txt")
	if err != nil {
		return nil, err
	}
	result, err := docutils.IndexTextFile(s.ctx, req.ProjectId, req.FileId, textPath, req.MinSize, req.MaxSize)
	if err != nil {
		return nil, err
	}
	return &document.IndexTextFileResp{
		FileId:      req.FileId,
		ChunkCount:  result.ChunkCount,
		ParentCount: result.ParentCount,
	}, nil
}
