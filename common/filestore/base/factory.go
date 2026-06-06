package base

import "github.com/MoScenix/ai-code/common/filestore"

func NewStore() (Store, error) {
	return NewLocalStore(filestore.GetConf().ShareDir.ShareDir)
}
