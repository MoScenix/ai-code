package conf

import (
	"path"
	"path/filepath"
	"strconv"
	"strings"

	filestoreconf "github.com/MoScenix/ai-code/common/filestore"
)

const (
	defaultStaticRoute     = "/static"
	defaultStaticRoot      = "/static"
	defaultStaticProject   = "project"
	defaultStaticAvatar    = "avatar"
	defaultStaticCover     = "cover"
	defaultDeployRoot      = "/static/deploy"
	defaultDeployURLPrefix = "/static/deploy"
	deployImageName        = "deploy.png"
)

func StaticRoute() string {
	if route := GetConf().Static.Route; route != "" {
		return route
	}
	return defaultStaticRoute
}

func StaticRoot() string {
	if root := GetConf().Static.Root; root != "" {
		return root
	}
	return defaultStaticRoot
}

func StaticRouteStripSlashes() int {
	return pathSlashCount(StaticRoute())
}

func ProjectDir(appID int64) string {
	return filepath.Join(filestoreconf.GetConf().ShareDir.ShareDir, strconv.FormatInt(appID, 10))
}

func AvatarPath(userID int64) string {
	return staticFilePath(staticSubDir(GetConf().Static.Avatar, defaultStaticAvatar), strconv.FormatInt(userID, 10)+".jpg")
}

func AvatarURL(userID int64) string {
	return staticURL(staticSubDir(GetConf().Static.Avatar, defaultStaticAvatar), strconv.FormatInt(userID, 10)+".jpg")
}

func CoverPath(deployKey string) string {
	return staticFilePath(staticSubDir(GetConf().Static.Cover, defaultStaticCover), deployKey, deployImageName)
}

func CoverURL(deployKey string) string {
	return staticURL(staticSubDir(GetConf().Static.Cover, defaultStaticCover), deployKey, deployImageName)
}

func DeployRoot() string {
	root := GetConf().Deploy.Root
	if root == "" {
		return defaultDeployRoot
	}
	return root
}

func DeployDir(deployKey string) string {
	return filepath.Join(DeployRoot(), deployKey)
}

func DeployURLPrefix() string {
	prefix := GetConf().Deploy.URLPrefix
	if prefix == "" {
		return defaultDeployURLPrefix
	}
	return prefix
}

func DeployURL(deployKey string) string {
	return joinURL(DeployURLPrefix(), deployKey) + "/"
}

func staticURL(parts ...string) string {
	prefix := GetConf().Static.URLPrefix
	if prefix == "" {
		prefix = StaticRoute()
	}
	return joinURL(append([]string{prefix}, parts...)...)
}

func staticFilePath(parts ...string) string {
	return filepath.Join(append([]string{StaticRoot()}, parts...)...)
}

func staticSubDir(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func pathSlashCount(value string) int {
	count := 0
	for _, part := range strings.Split(value, "/") {
		if part != "" {
			count++
		}
	}
	return count
}

func joinURL(parts ...string) string {
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, "/")
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}
	if len(cleaned) == 0 {
		return "/"
	}
	return "/" + path.Join(cleaned...)
}
