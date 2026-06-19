package rpcmeta

import (
	"context"
	"strconv"

	"github.com/bytedance/gopkg/cloud/metainfo"
)

const (
	OperatorIDKey   = "X_USER_ID"
	OperatorRoleKey = "X_USER_ROLE"
	AdminRole       = "admin"
)

type Identity struct {
	OperatorID   string
	OperatorRole string
}

func FromContext(ctx context.Context) Identity {
	operatorID, _ := metainfo.GetPersistentValue(ctx, OperatorIDKey)
	operatorRole, _ := metainfo.GetPersistentValue(ctx, OperatorRoleKey)
	return Identity{
		OperatorID:   operatorID,
		OperatorRole: operatorRole,
	}
}

func WithIdentity(ctx context.Context, identity Identity) context.Context {
	if identity.OperatorID != "" {
		ctx = metainfo.WithPersistentValue(ctx, OperatorIDKey, identity.OperatorID)
	}
	if identity.OperatorRole != "" {
		ctx = metainfo.WithPersistentValue(ctx, OperatorRoleKey, identity.OperatorRole)
	}
	return ctx
}

func WithOperator(ctx context.Context, operatorID int64, operatorRole string) context.Context {
	if operatorID > 0 {
		ctx = metainfo.WithPersistentValue(ctx, OperatorIDKey, strconv.FormatInt(operatorID, 10))
	}
	if operatorRole != "" {
		ctx = metainfo.WithPersistentValue(ctx, OperatorRoleKey, operatorRole)
	}
	return ctx
}

func OperatorIDFromContext(ctx context.Context) (int64, bool) {
	operatorID, ok := metainfo.GetPersistentValue(ctx, OperatorIDKey)
	if !ok || operatorID == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(operatorID, 10, 64)
	return id, err == nil
}
