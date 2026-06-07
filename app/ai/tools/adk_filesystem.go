package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/MoScenix/ai-code/common/filestore"
	"github.com/MoScenix/ai-code/common/filestore/project"
	"github.com/bmatcuk/doublestar/v4"
	adkfs "github.com/cloudwego/eino/adk/filesystem"
)

type ProjectFilesystemBackend struct {
	store project.Store
}

func NewProjectFilesystemBackend(store project.Store) *ProjectFilesystemBackend {
	return &ProjectFilesystemBackend{store: store}
}

func (b *ProjectFilesystemBackend) LsInfo(ctx context.Context, req *adkfs.LsInfoRequest) ([]adkfs.FileInfo, error) {
	infos, err := b.store.List(ctx, req.Path)
	if err != nil {
		return nil, err
	}
	result := make([]adkfs.FileInfo, 0, len(infos))
	for _, info := range infos {
		result = append(result, fileInfo(req.Path, info))
	}
	return result, nil
}

func (b *ProjectFilesystemBackend) Read(ctx context.Context, req *adkfs.ReadRequest) (*adkfs.FileContent, error) {
	data, err := b.store.ReadFile(ctx, req.FilePath)
	if err != nil {
		return nil, err
	}
	return &adkfs.FileContent{
		Content: readLines(string(data), req.Offset, req.Limit),
	}, nil
}

func (b *ProjectFilesystemBackend) Write(ctx context.Context, req *adkfs.WriteRequest) error {
	return b.store.WriteFile(ctx, req.FilePath, []byte(req.Content))
}

func (b *ProjectFilesystemBackend) Edit(ctx context.Context, req *adkfs.EditRequest) error {
	if req.ReplaceAll {
		return b.replaceAll(ctx, req.FilePath, req.OldString, req.NewString)
	}
	return b.store.EditFile(ctx, req.FilePath, req.OldString, req.NewString)
}

func (b *ProjectFilesystemBackend) GrepRaw(ctx context.Context, req *adkfs.GrepRequest) ([]adkfs.GrepMatch, error) {
	if req.Pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}
	pattern := req.Pattern
	if req.CaseInsensitive {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid grep pattern: %w", err)
	}

	files, err := b.listFiles(ctx, req.Path)
	if err != nil {
		return nil, err
	}

	var matches []adkfs.GrepMatch
	for _, info := range files {
		if !matchFileType(info.Key, req.FileType) {
			continue
		}
		if req.Glob != "" {
			matched, err := matchGlob(req.Glob, info.Key)
			if err != nil {
				return nil, err
			}
			if !matched {
				continue
			}
		}

		data, err := b.store.ReadFile(ctx, info.Key)
		if err != nil {
			return nil, err
		}
		matches = append(matches, grepContent(info.Key, string(data), re, req)...)
	}

	sort.Slice(matches, func(i int, j int) bool {
		if matches[i].Path == matches[j].Path {
			return matches[i].Line < matches[j].Line
		}
		return matches[i].Path < matches[j].Path
	})
	return matches, nil
}

func (b *ProjectFilesystemBackend) GlobInfo(ctx context.Context, req *adkfs.GlobInfoRequest) ([]adkfs.FileInfo, error) {
	if req.Pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}

	entries, err := b.listEntries(ctx, req.Path)
	if err != nil {
		return nil, err
	}

	result := make([]adkfs.FileInfo, 0)
	for _, info := range entries {
		matched, err := matchGlob(req.Pattern, info.Key)
		if err != nil {
			return nil, err
		}
		if matched {
			result = append(result, fileInfoFromPath(info))
		}
	}

	sort.Slice(result, func(i int, j int) bool {
		return result[i].Path < result[j].Path
	})
	return result, nil
}

func (b *ProjectFilesystemBackend) replaceAll(ctx context.Context, path string, oldString string, newString string) error {
	if oldString == "" {
		return filestore.ErrTextNotFound
	}
	data, err := b.store.ReadFile(ctx, path)
	if err != nil {
		return err
	}
	content := string(data)
	if !strings.Contains(content, oldString) {
		return filestore.ErrTextNotFound
	}
	next := strings.ReplaceAll(content, oldString, newString)
	return b.store.WriteFile(ctx, path, []byte(next))
}

func fileInfo(parent string, info filestore.ObjectInfo) adkfs.FileInfo {
	path := info.Key
	if info.Name != "" {
		path = filepath.ToSlash(filepath.Join(parent, info.Name))
	}
	return adkfs.FileInfo{
		Path:       path,
		IsDir:      info.IsDir,
		Size:       info.Size,
		ModifiedAt: info.ModTime.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func readLines(content string, offset int, limit int) string {
	if offset <= 1 && limit <= 0 {
		return content
	}
	if offset < 1 {
		offset = 1
	}
	lines := strings.Split(content, "\n")
	start := offset - 1
	if start >= len(lines) {
		return ""
	}
	end := len(lines)
	if limit > 0 && start+limit < end {
		end = start + limit
	}
	return strings.Join(lines[start:end], "\n")
}

func (b *ProjectFilesystemBackend) listFiles(ctx context.Context, root string) ([]filestore.ObjectInfo, error) {
	entries, err := b.listEntries(ctx, root)
	if err != nil {
		return nil, err
	}
	files := make([]filestore.ObjectInfo, 0, len(entries))
	for _, info := range entries {
		if !info.IsDir {
			files = append(files, info)
		}
	}
	return files, nil
}

func (b *ProjectFilesystemBackend) listEntries(ctx context.Context, root string) ([]filestore.ObjectInfo, error) {
	root = cleanProjectPath(root)
	var result []filestore.ObjectInfo
	if err := b.walk(ctx, root, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (b *ProjectFilesystemBackend) walk(ctx context.Context, dir string, result *[]filestore.ObjectInfo) error {
	infos, err := b.store.List(ctx, dir)
	if err != nil {
		return err
	}
	for _, info := range infos {
		path := cleanProjectPath(info.Key)
		if info.Name != "" {
			path = filepath.ToSlash(filepath.Join(dir, info.Name))
		}
		info.Key = path
		*result = append(*result, info)
		if info.IsDir {
			if err := b.walk(ctx, path, result); err != nil {
				return err
			}
		}
	}
	return nil
}

func fileInfoFromPath(info filestore.ObjectInfo) adkfs.FileInfo {
	return adkfs.FileInfo{
		Path:       cleanProjectPath(info.Key),
		IsDir:      info.IsDir,
		Size:       info.Size,
		ModifiedAt: info.ModTime.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func grepContent(path string, content string, re *regexp.Regexp, req *adkfs.GrepRequest) []adkfs.GrepMatch {
	if req.EnableMultiline {
		return grepMultiline(path, content, re)
	}

	lines := strings.Split(content, "\n")
	include := make(map[int]struct{})
	for i, line := range lines {
		if !re.MatchString(line) {
			continue
		}
		start := i - req.BeforeLines
		if start < 0 {
			start = 0
		}
		end := i + req.AfterLines
		if end >= len(lines) {
			end = len(lines) - 1
		}
		for lineIndex := start; lineIndex <= end; lineIndex++ {
			include[lineIndex] = struct{}{}
		}
	}

	result := make([]adkfs.GrepMatch, 0, len(include))
	for i, line := range lines {
		if _, ok := include[i]; !ok {
			continue
		}
		result = append(result, adkfs.GrepMatch{
			Path:    path,
			Line:    i + 1,
			Content: line,
		})
	}
	return result
}

func grepMultiline(path string, content string, re *regexp.Regexp) []adkfs.GrepMatch {
	indexes := re.FindAllStringIndex(content, -1)
	result := make([]adkfs.GrepMatch, 0, len(indexes))
	for _, index := range indexes {
		line := strings.Count(content[:index[0]], "\n") + 1
		match := content[index[0]:index[1]]
		result = append(result, adkfs.GrepMatch{
			Path:    path,
			Line:    line,
			Content: match,
		})
	}
	return result
}

func matchGlob(pattern string, path string) (bool, error) {
	pattern = cleanProjectPath(pattern)
	path = cleanProjectPath(path)
	if strings.Contains(pattern, "/") || strings.Contains(pattern, "**") {
		return doublestar.Match(pattern, path)
	}
	return doublestar.Match(pattern, filepath.Base(path))
}

func matchFileType(path string, fileType string) bool {
	fileType = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(fileType)), ".")
	if fileType == "" {
		return true
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if ext == fileType {
		return true
	}
	switch fileType {
	case "js":
		return ext == "js" || ext == "jsx" || ext == "mjs" || ext == "cjs" || ext == "vue"
	case "ts":
		return ext == "ts" || ext == "tsx"
	case "html":
		return ext == "html" || ext == "htm"
	case "css":
		return ext == "css" || ext == "scss" || ext == "sass" || ext == "less"
	case "go":
		return ext == "go"
	case "json":
		return ext == "json" || ext == "jsonl"
	case "md", "markdown":
		return ext == "md" || ext == "markdown"
	default:
		return false
	}
}

func cleanProjectPath(path string) string {
	path = strings.TrimSpace(filepath.ToSlash(path))
	path = strings.TrimPrefix(path, "/")
	if path == "." {
		return ""
	}
	return path
}
