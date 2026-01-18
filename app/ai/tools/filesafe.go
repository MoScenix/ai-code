package tools

import (
	"path/filepath"
	"runtime"
	"strings"
)

func IsSubPathAbs(child, parent string, ok bool) bool {
	if !filepath.IsAbs(child) || !filepath.IsAbs(parent) {
		return false
	}
	childN := normAbs(child)
	parentN := normAbs(parent)
	if samePath(childN, parentN) && ok {
		return true
	}
	parentWithSep := ensureTrailingSep(parentN)
	childN = filepath.Clean(childN)
	return strings.HasPrefix(childN, parentWithSep)
}

func normAbs(p string) string {
	p = filepath.Clean(p)

	if real, err := filepath.EvalSymlinks(p); err == nil {
		p = filepath.Clean(real)
	}
	if runtime.GOOS == "windows" {
		p = strings.ToLower(p)
	}
	return p
}

func ensureTrailingSep(p string) string {
	p = filepath.Clean(p)
	if !strings.HasSuffix(p, string(filepath.Separator)) {
		p += string(filepath.Separator)
	}
	return p
}

func samePath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}
