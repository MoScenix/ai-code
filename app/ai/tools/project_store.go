package tools

import (
	"context"
	"fmt"

	lutils "github.com/MoScenix/ai-code/app/ai/utils"
	"github.com/MoScenix/ai-code/common/filestore/project"
)

func projectStoreFromContext(ctx context.Context) (project.Store, error) {
	store, ok := ctx.Value(lutils.ProjectFileStore).(project.Store)
	if !ok || store == nil {
		return nil, fmt.Errorf("ProjectFileStore is missing in context")
	}
	return store, nil
}
