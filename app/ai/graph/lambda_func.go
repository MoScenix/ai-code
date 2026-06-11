package graph

import (
	"context"

	codernode "github.com/MoScenix/ai-code/app/ai/node/coder"
	designernode "github.com/MoScenix/ai-code/app/ai/node/designer"
)

// newLambda component initialization function of node 'Designer' in graph 'aicode'
func newLambda(ctx context.Context, input map[string]any) (output map[string]any, err error) {
	return designernode.Run(ctx)
}

// newLambda1 component initialization function of node 'Coder' in graph 'aicode'
func newLambda1(ctx context.Context, input map[string]any) (output map[string]any, err error) {
	return codernode.Run(ctx, input)
}
