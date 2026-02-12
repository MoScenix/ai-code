package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/bytedance/gopkg/cloud/metainfo"
)

const (
	metaUserIDKey   = "x-user-id"
	metaUserRoleKey = "x-user-role"
	adminRole       = "admin"
)

var (
	errUnauthorized = errors.New("unauthorized: missing operator identity")
	errForbidden    = errors.New("forbidden: no permission")
	errDBNotReady   = errors.New("database not initialized")
)

type operator struct {
	userID uint
	role   string
}

func (o operator) isAdmin() bool {
	return o.role == adminRole
}

func getOperator(ctx context.Context) (operator, error) {
	userIDStr, ok := metainfo.GetPersistentValue(ctx, metaUserIDKey)
	if !ok {
		return operator{}, errUnauthorized
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		return operator{}, errUnauthorized
	}
	role, _ := metainfo.GetPersistentValue(ctx, metaUserRoleKey)
	return operator{
		userID: uint(userID),
		role:   role,
	}, nil
}

func mustOwnerOrAdmin(op operator, ownerID uint) error {
	if op.isAdmin() || op.userID == ownerID {
		return nil
	}
	return errForbidden
}
