package utils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func CopyDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	// 创建目标根目录
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		info, err := d.Info()
		if err != nil {
			return err
		}

		// 符号链接：这里选择跳过（按需改成复制链接本身）
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		if d.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		// 普通文件：复制内容 + 权限
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(srcFile, dstFile string, perm fs.FileMode) error {
	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dstFile), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(dstFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
