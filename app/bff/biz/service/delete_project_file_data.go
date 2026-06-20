package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MoScenix/ai-code/app/bff/conf"
	"github.com/MoScenix/ai-code/app/bff/infra/rpc"
	document "github.com/MoScenix/ai-code/rpc_gen/kitex_gen/document"
)

func deleteProjectFileData(ctx context.Context, projectID int64) error {
	if projectID <= 0 {
		return nil
	}
	if _, err := rpc.DocumentClient.DeleteProjectFileData(ctx, &document.DeleteProjectFileDataReq{
		ProjectId: projectID,
	}); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(conf.StaticRoot(), "document", fmt.Sprintf("%d", projectID))); err != nil {
		return err
	}
	return nil
}
