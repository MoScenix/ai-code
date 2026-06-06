package filestore

import "time"

type ObjectInfo struct {
	Key         string
	Name        string
	IsDir       bool
	Size        int64
	ModTime     time.Time
	ETag        string
	ContentType string
}
