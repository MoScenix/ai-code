package main

import (
	"context"
	"github.com/MoScenix/ai-code/app/document/biz/service"
	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"
)

// DocumentServiceImpl implements the last service interface defined in the IDL.
type DocumentServiceImpl struct{}

// ParsePDFToText implements the DocumentServiceImpl interface.
func (s *DocumentServiceImpl) ParsePDFToText(ctx context.Context, req *document.ParsePDFToTextReq) (resp *document.ParsePDFToTextResp, err error) {
	resp, err = service.NewParsePDFToTextService(ctx).Run(req)

	return resp, err
}

// SearchFile implements the DocumentServiceImpl interface.
func (s *DocumentServiceImpl) SearchFile(ctx context.Context, req *document.SearchFileReq) (resp *document.SearchFileResp, err error) {
	resp, err = service.NewSearchFileService(ctx).Run(req)

	return resp, err
}

// DeleteProjectFileData implements the DocumentServiceImpl interface.
func (s *DocumentServiceImpl) DeleteProjectFileData(ctx context.Context, req *document.DeleteProjectFileDataReq) (resp *document.DeleteProjectFileDataResp, err error) {
	resp, err = service.NewDeleteProjectFileDataService(ctx).Run(req)

	return resp, err
}

// IndexTextFile implements the DocumentServiceImpl interface.
func (s *DocumentServiceImpl) IndexTextFile(ctx context.Context, req *document.IndexTextFileReq) (resp *document.IndexTextFileResp, err error) {
	resp, err = service.NewIndexTextFileService(ctx).Run(req)

	return resp, err
}
