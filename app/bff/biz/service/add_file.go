package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MoScenix/ai-code/app/bff/biz/utils"
	"github.com/MoScenix/ai-code/app/bff/conf"
	lapp "github.com/MoScenix/ai-code/app/bff/hertz_gen/bff/app"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	rpcapp "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/app"
	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"
	hertzapp "github.com/cloudwego/hertz/pkg/app"
)

type AddFileService struct {
	RequestContext *hertzapp.RequestContext
	Context        context.Context
}

func NewAddFileService(Context context.Context, RequestContext *hertzapp.RequestContext) *AddFileService {
	return &AddFileService{RequestContext: RequestContext, Context: Context}
}

func (h *AddFileService) Run(req *lapp.AddFileRequest) (resp *lapp.BaseResponseString, err error) {
	userID, ok := utils.UserIDFromContext(h.Context)
	if !ok {
		return nil, utils.ErrUnauthorizedUserID
	}
	fileHeader, err := h.RequestContext.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("missing upload file: %w", err)
	}

	fileID := time.Now().UnixNano()
	filename := safeFilename(fileHeader.Filename)
	if filename == "" {
		return nil, fmt.Errorf("missing upload filename")
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".pdf" && ext != ".txt" {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	dir := filepath.Join(conf.StaticRoot(), "document", fmt.Sprintf("%d", req.AppId), fmt.Sprintf("%d", fileID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	savePath := filepath.Join(dir, filename)
	if err := h.RequestContext.SaveUploadedFile(fileHeader, savePath); err != nil {
		return nil, err
	}

	meta := fileMessageContent{
		FileID:      fileID,
		Filename:    filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Size:        fileHeader.Size,
	}
	if stat, statErr := os.Stat(savePath); statErr == nil {
		meta.Size = stat.Size()
	}

	textPath := savePath
	textSize := meta.Size
	if ext == ".pdf" {
		parsed, err := rpc.DocumentClient.ParsePDFToText(h.Context, &document.ParsePDFToTextReq{
			ProjectId: req.AppId,
			FileId:    fileID,
		})
		if err != nil {
			return nil, err
		}
		meta.TextFilename = parsed.TextFilename
		textSize = parsed.TextSize
		textPath = filepath.Join(dir, parsed.TextFilename)
	} else {
		meta.TextFilename = filename
	}
	meta.TextSize = textSize

	threshold := conf.GetConf().File.BigThresholdBytes
	meta.IsBig = threshold > 0 && textSize > threshold
	if meta.IsBig {
		chunkResp, err := rpc.DocumentClient.IndexTextFile(h.Context, &document.IndexTextFileReq{
			ProjectId: req.AppId,
			FileId:    fileID,
			MinSize:   conf.GetConf().File.ChunkMinSize,
			MaxSize:   conf.GetConf().File.ChunkMaxSize,
		})
		if err != nil {
			return nil, err
		}
		meta.ChunkCount = chunkResp.ChunkCount
		meta.ParentCount = chunkResp.ParentCount
	} else {
		text, err := os.ReadFile(textPath)
		if err != nil {
			return nil, err
		}
		meta.Text = string(text)
	}

	content, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	addResp, err := rpc.AppClient.AddMessage(h.Context, &rpcapp.AddMessageReq{
		AppId:   req.AppId,
		UserId:  userID,
		Role:    "user",
		Content: string(content),
		IsFile:  true,
	})
	if err != nil {
		return nil, err
	}
	return &lapp.BaseResponseString{
		Code:    0,
		Message: "success",
		Data:    fmt.Sprintf("%d", addResp.Id),
	}, nil
}

type fileMessageContent struct {
	FileID       int64  `json:"fileId"`
	Filename     string `json:"filename"`
	ContentType  string `json:"contentType"`
	Size         int64  `json:"size"`
	TextFilename string `json:"textFilename"`
	TextSize     int64  `json:"textSize"`
	IsBig        bool   `json:"isBig"`
	ChunkCount   int64  `json:"chunkCount,omitempty"`
	ParentCount  int64  `json:"parentCount,omitempty"`
	Text         string `json:"text,omitempty"`
}

func safeFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, "\x00", "")
	return name
}
