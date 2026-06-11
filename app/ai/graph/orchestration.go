package graph

import (
	"context"

	"github.com/cloudwego/eino/compose"
)

func Buildaicode(ctx context.Context) (r compose.Runnable[any, any], err error) {
	g := compose.NewGraph[any, any]()
	_ = g.AddLambdaNode(designerNode, compose.InvokableLambda(newLambda))
	_ = g.AddLambdaNode(coderNode, compose.InvokableLambda(newLambda1))
	_ = g.AddEdge(coderNode, compose.END)
	_ = g.AddEdge(designerNode, coderNode)
	_ = g.AddBranch(compose.START, compose.NewGraphBranch(newBranch, map[string]bool{designerNode: true, coderNode: true}))
	opts := []compose.GraphCompileOption{compose.WithGraphName("aicode")}
	if store, ok := newGraphCheckpointStore(ctx); ok {
		opts = append(opts, compose.WithCheckPointStore(store))
	}
	r, err = g.Compile(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return r, err
}
