package fsutil_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/johnkil/polyrepo-workspace-kit/internal/fsutil"
)

func TestCopyFilePreservesContentAndMode(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "nested", "target")
	if err := os.WriteFile(source, []byte("canonical\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := fsutil.CopyFile(source, target); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "canonical\n" {
		t.Fatalf("unexpected target content: %q", string(data))
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		return
	}
	if got := info.Mode().Perm(); got != 0o755 {
		t.Fatalf("expected target mode 0755, got %o", got)
	}
}

func TestBackupExistingDoesNotOverwriteExistingBackupFile(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	backup := filepath.Join(root, "source.bak.20260419T120000Z")
	if err := os.WriteFile(source, []byte("new backup\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backup, []byte("older backup\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := fsutil.BackupExisting(source, backup); err == nil {
		t.Fatal("expected existing backup path to be rejected")
	}

	data, err := os.ReadFile(backup)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "older backup\n" {
		t.Fatalf("existing backup was overwritten: %q", string(data))
	}
}

func TestBackupExistingDoesNotOverwriteExistingBackupDirectory(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source-dir")
	backup := filepath.Join(root, "source-dir.bak.20260419T120000Z")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "file.txt"), []byte("new backup\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(backup, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backup, "file.txt"), []byte("older backup\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := fsutil.BackupExisting(source, backup); err == nil {
		t.Fatal("expected existing backup directory to be rejected")
	}

	data, err := os.ReadFile(filepath.Join(backup, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "older backup\n" {
		t.Fatalf("existing backup directory was overwritten: %q", string(data))
	}
}

func TestBackupExistingDoesNotReplaceExistingEmptyBackupDirectory(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source-dir")
	backup := filepath.Join(root, "source-dir.bak.20260419T120000Z")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "file.txt"), []byte("new backup\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(backup, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := fsutil.BackupExisting(source, backup); err == nil {
		t.Fatal("expected existing empty backup directory to be rejected")
	}
	if _, err := os.Stat(filepath.Join(backup, "file.txt")); !os.IsNotExist(err) {
		t.Fatalf("existing empty backup directory was replaced or populated: %v", err)
	}
}

func TestBackupExistingCleansPartialBackupDirectoryOnError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod-based unreadable file setup is not reliable on Windows")
	}
	root := t.TempDir()
	source := filepath.Join(root, "source-dir")
	backup := filepath.Join(root, "source-dir.bak.20260419T120000Z")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "a-readable.txt"), []byte("copied first\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	unreadable := filepath.Join(source, "z-unreadable.txt")
	if err := os.WriteFile(unreadable, []byte("cannot copy\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(unreadable, 0); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chmod(unreadable, 0o644)
	}()

	if err := fsutil.BackupExisting(source, backup); err == nil {
		t.Fatal("expected backup copy to fail")
	}
	if _, err := os.Stat(backup); !os.IsNotExist(err) {
		t.Fatalf("partial backup path still exists: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(root, ".source-dir.bak.20260419T120000Z.tmp-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary backup path was not cleaned up: %v", matches)
	}
}
