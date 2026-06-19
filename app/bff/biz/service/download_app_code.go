package service

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/MoScenix/ai-code/app/bff/conf"
	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/pkg/klog"
)

type DownloadAppCodeService struct {
	RequestContext *app.RequestContext
	Context        context.Context
}

func NewDownloadAppCodeService(Context context.Context, RequestContext *app.RequestContext) *DownloadAppCodeService {
	return &DownloadAppCodeService{RequestContext: RequestContext, Context: Context}
}

func (h *DownloadAppCodeService) Run(req *lapp.DownloadAppCodeRequest) (resp *lapp.BaseResponseBytes, err error) {
	appDir := conf.ProjectDir(req.AppId)
	if _, err = os.Stat(appDir); err != nil {
		klog.CtxWarnf(h.Context, "download app code not found: app_id=%d err=%v", req.AppId, err)
		return nil, errors.New("app code not found")
	}
	h.RequestContext.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="app-%d.zip"`, req.AppId))
	zipWriter := zip.NewWriter(h.RequestContext)

	err = filepath.Walk(appDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(appDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		_ = zipWriter.Close()
		return nil, err
	}

	if err = zipWriter.Close(); err != nil {
		return nil, err
	}
	return resp, nil
}
