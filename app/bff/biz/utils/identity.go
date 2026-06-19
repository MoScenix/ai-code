package utils

import (
	"context"
	"errors"
	"strconv"

	"github.com/MoScenix/ai-code/common/rpcmeta"
)

var ErrUnauthorizedUserID = errors.New("unauthorized: missing user id")

func UserIDFromContext(ctx context.Context) (int64, bool) {
	return ParseUserID(ctx.Value(UserIdKey))
}

func WithIdentityMeta(ctx context.Context) context.Context {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return ctx
	}
	role, _ := ctx.Value(UserRoleKey).(string)
	return rpcmeta.WithOperator(ctx, userID, role)
}

func ParseUserID(v any) (int64, bool) {
	switch value := v.(type) {
	case int64:
		return value, true
	case int32:
		return int64(value), true
	case int:
		return int64(value), true
	case float64:
		return int64(value), true
	case float32:
		return int64(value), true
	case uint64:
		return int64(value), true
	case uint32:
		return int64(value), true
	case uint:
		return int64(value), true
	case string:
		id, err := strconv.ParseInt(value, 10, 64)
		return id, err == nil
	default:
		return 0, false
	}
}
