package fsutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func NormalizeText(text string) string {
	return strings.TrimRight(text, "\n") + "\n"
}

func WriteFileAtomic(path string, data []byte) error {
	return WriteFileAtomicMode(path, data, 0o644)
}

func WriteFileAtomicMode(path string, data []byte, mode os.FileMode) error {
	tmpPath, err := writeTempFile(path, data, mode)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpPath)
	}()
	return os.Rename(tmpPath, path)
}

func WriteFileExclusive(path string, data []byte) error {
	return WriteFileExclusiveMode(path, data, 0o644)
}

func WriteFileExclusiveMode(path string, data []byte, mode os.FileMode) error {
	tmpPath, err := writeTempFile(path, data, mode)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpPath)
	}()
	return os.Link(tmpPath, path)
}

func writeTempFile(path string, data []byte, mode os.FileMode) (string, error) {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	cleanup := func(err error) (string, error) {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	if _, err := tmp.Write(data); err != nil {
		return cleanup(err)
	}
	if err := tmp.Chmod(mode.Perm()); err != nil {
		return cleanup(err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return tmpPath, nil
}

func WriteTextIfMissing(path string, text string) error {
	if Exists(path) {
		return nil
	}
	return WriteFileAtomic(path, []byte(NormalizeText(text)))
}

func CopyFile(source string, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	return WriteFileAtomicMode(target, data, info.Mode().Perm())
}

func CopyDir(source string, target string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(target, sourceInfo.Mode().Perm()); err != nil {
		return err
	}
	return copyDirContents(source, target)
}

func CopyDirExclusive(source string, target string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	if err := EnsureDir(filepath.Dir(target)); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(target), "."+filepath.Base(target)+".tmp-*")
	if err != nil {
		return err
	}
	defer func() {
		if tmp != "" {
			_ = os.RemoveAll(tmp)
		}
	}()
	if err := os.Chmod(tmp, sourceInfo.Mode().Perm()); err != nil {
		return err
	}
	if err := copyDirContents(source, tmp); err != nil {
		return err
	}
	if err := os.Mkdir(target, sourceInfo.Mode().Perm()); err != nil {
		return err
	}
	complete := false
	defer func() {
		if !complete {
			_ = os.RemoveAll(target)
		}
	}()
	if err := copyDirContents(tmp, target); err != nil {
		return err
	}
	complete = true
	return nil
}

func copyDirContents(source string, target string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		dst := filepath.Join(target, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(dst, info.Mode().Perm())
		}
		return copyFileWithMode(path, dst, info.Mode().Perm())
	})
}

func copyFileWithMode(source string, target string, mode os.FileMode) error {
	if err := EnsureDir(filepath.Dir(target)); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func BackupExisting(path string, backupPath string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err := CopyDirExclusive(path, backupPath); err != nil {
			if os.IsExist(err) {
				return fmt.Errorf("backup path already exists: %s", backupPath)
			}
			return err
		}
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := WriteFileExclusiveMode(backupPath, data, info.Mode().Perm()); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("backup path already exists: %s", backupPath)
		}
		return err
	}
	return nil
}

func BackupPath(path string, now time.Time) string {
	base := path + ".bak." + now.UTC().Format("20060102T150405Z")
	if !Exists(base) {
		return base
	}
	for index := 1; ; index++ {
		candidate := fmt.Sprintf("%s.%03d", base, index)
		if !Exists(candidate) {
			return candidate
		}
	}
}

func SameText(path string, text string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return string(data) == NormalizeText(text)
}

func SameFile(source string, target string) bool {
	left, err := os.ReadFile(source)
	if err != nil {
		return false
	}
	right, err := os.ReadFile(target)
	if err != nil {
		return false
	}
	return bytes.Equal(left, right)
}

func SameDir(source string, target string) bool {
	left, err := TreeSignature(source)
	if err != nil {
		return false
	}
	right, err := TreeSignature(target)
	if err != nil {
		return false
	}
	return len(left) == len(right) && equalSignatures(left, right)
}

func TreeSignature(root string) (map[string][]byte, error) {
	out := map[string][]byte{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = data
		return nil
	})
	return out, err
}

func equalSignatures(left map[string][]byte, right map[string][]byte) bool {
	for key, leftValue := range left {
		rightValue, ok := right[key]
		if !ok || !bytes.Equal(leftValue, rightValue) {
			return false
		}
	}
	return true
}

func ExpandHome(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func Abs(path string) (string, error) {
	return filepath.Abs(ExpandHome(path))
}
