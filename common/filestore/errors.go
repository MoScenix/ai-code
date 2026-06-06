package filestore

import "errors"

var (
	ErrNotFound         = errors.New("filestore: object not found")
	ErrInvalidKey       = errors.New("filestore: invalid key")
	ErrCacheExpired     = errors.New("filestore: cache expired")
	ErrTextNotFound     = errors.New("filestore: text not found")
	ErrTextNotUnique    = errors.New("filestore: text not unique")
	ErrEmptyProjectID   = errors.New("filestore: empty project id")
	ErrInvalidProjectID = errors.New("filestore: invalid project id")
)
