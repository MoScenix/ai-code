package service

import (
	"context"
	"path/filepath"

	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"
)

type ParsePDFToTextService struct {
	ctx context.Context
} // NewParsePDFToTextService new ParsePDFToTextService
func NewParsePDFToTextService(ctx context.Context) *ParsePDFToTextService {
	return &ParsePDFToTextService{ctx: ctx}
}

// Run create note info
func (s *ParsePDFToTextService) Run(req *document.ParsePDFToTextReq) (resp *document.ParsePDFToTextResp, err error) {
	dir := projectFileDir(req.ProjectId, req.FileId)
	pdfPath, err := findFileByExt(dir, ".pdf")
	if err != nil {
		return nil, err
	}
	txtPath, size, err := parsePDFToTextFile(pdfPath)
	if err != nil {
		return nil, err
	}
	return &document.ParsePDFToTextResp{
		FileId:       req.FileId,
		TextFilename: filepath.Base(txtPath),
		TextSize:     size,
	}, nil
}
