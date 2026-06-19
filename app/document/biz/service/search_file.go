package service

import (
	"context"

	docutils "github.com/MoScenix/ai-code/app/document/utils"
	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"
)

type SearchFileService struct {
	ctx context.Context
} // NewSearchFileService new SearchFileService
func NewSearchFileService(ctx context.Context) *SearchFileService {
	return &SearchFileService{ctx: ctx}
}

// Run create note info
func (s *SearchFileService) Run(req *document.SearchFileReq) (resp *document.SearchFileResp, err error) {
	dir := projectFileDir(req.ProjectId, req.FileId)
	hits, err := docutils.SearchIndexedFile(s.ctx, req.ProjectId, req.FileId, dir, req.Query, req.TopK)
	if err != nil {
		return nil, err
	}
	resp = &document.SearchFileResp{
		Hits: make([]*document.SearchHit, 0, len(hits)),
	}
	for _, hit := range hits {
		resp.Hits = append(resp.Hits, &document.SearchHit{
			FileId:   hit.FileID,
			ChunkId:  hit.ChunkID,
			ParentId: hit.ParentID,
			Content:  hit.Content,
			Score:    hit.Score,
		})
	}
	return resp, nil
}
