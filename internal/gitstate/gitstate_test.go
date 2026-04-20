package gitstate

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseStatusPorcelainV2ZHandlesSpacesAndRenames(t *testing.T) {
	input := []byte(
		"1 M. N... 100644 100644 100644 abcdef abcdef path with spaces.txt\x00" +
			"? new file.txt\x00" +
			"2 R. N... 100644 100644 100644 abcdef abcdef R100 renamed file.txt\x00old file.txt\x00" +
			"! ignored.txt\x00",
	)

	dirty, untracked := parseStatusPorcelainV2Z(input)

	wantDirty := []string{"path with spaces.txt", "old file.txt -> renamed file.txt"}
	if !reflect.DeepEqual(dirty, wantDirty) {
		t.Fatalf("dirty paths mismatch:\nwant %#v\ngot  %#v", wantDirty, dirty)
	}
	wantUntracked := []string{"new file.txt"}
	if !reflect.DeepEqual(untracked, wantUntracked) {
		t.Fatalf("untracked paths mismatch:\nwant %#v\ngot  %#v", wantUntracked, untracked)
	}
}

func TestInspectReportsLocalOnlyGitState(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "repo")
	initGitRepo(t, repo)

	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "new.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	status, err := Inspect(repo)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Git {
		t.Fatal("expected git checkout")
	}
	if status.Clean {
		t.Fatal("expected dirty checkout")
	}
	if len(status.DirtyPaths) != 1 || status.DirtyPaths[0] != "README.md" {
		t.Fatalf("unexpected dirty paths: %#v", status.DirtyPaths)
	}
	if len(status.UntrackedPaths) != 1 || status.UntrackedPaths[0] != "new.txt" {
		t.Fatalf("unexpected untracked paths: %#v", status.UntrackedPaths)
	}
	if status.HasUpstream {
		t.Fatalf("expected no upstream, got %q", status.Upstream)
	}
}

func TestInspectReportsAheadBehindFromLocalRemoteTrackingRefs(t *testing.T) {
	tmp := t.TempDir()
	remote := filepath.Join(tmp, "remote.git")
	seed := filepath.Join(tmp, "seed")
	local := filepath.Join(tmp, "local")
	other := filepath.Join(tmp, "other")

	gitRun(t, tmp, "git", "init", "--bare", remote)
	initGitRepo(t, seed)
	gitRun(t, seed, "git", "remote", "add", "origin", remote)
	gitRun(t, seed, "git", "push", "-u", "origin", "HEAD")
	gitRun(t, tmp, "git", "clone", remote, local)
	gitRun(t, tmp, "git", "clone", remote, other)
	configUser(t, local)
	configUser(t, other)

	if err := os.WriteFile(filepath.Join(local, "local.txt"), []byte("local\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, local, "git", "add", ".")
	gitRun(t, local, "git", "commit", "-m", "local")

	if err := os.WriteFile(filepath.Join(other, "remote.txt"), []byte("remote\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, other, "git", "add", ".")
	gitRun(t, other, "git", "commit", "-m", "remote")
	gitRun(t, other, "git", "push")
	gitRun(t, local, "git", "fetch", "origin")

	status, err := Inspect(local)
	if err != nil {
		t.Fatal(err)
	}
	if !status.HasUpstream {
		t.Fatalf("expected upstream, got %#v", status)
	}
	if !status.HasDivergence {
		t.Fatalf("expected ahead/behind counts, got %#v", status)
	}
	if status.Ahead != 1 || status.Behind != 1 {
		t.Fatalf("expected ahead=1 behind=1, got ahead=%d behind=%d", status.Ahead, status.Behind)
	}
}

func TestInspectKeepsUpstreamButHidesCountsWhenDivergenceUnavailable(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "repo")
	initGitRepo(t, repo)
	branch := strings.TrimSpace(gitOutput(t, repo, "git", "branch", "--show-current"))
	if branch == "" {
		t.Fatal("expected current branch")
	}
	gitRun(t, repo, "git", "remote", "add", "origin", "https://example.invalid/repo.git")
	gitRun(t, repo, "git", "config", "branch."+branch+".remote", "origin")
	gitRun(t, repo, "git", "config", "branch."+branch+".merge", "refs/heads/main")

	remoteRefs := filepath.Join(repo, ".git", "refs", "remotes", "origin")
	if err := os.MkdirAll(remoteRefs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(remoteRefs, "main"), []byte("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	status, err := Inspect(repo)
	if err != nil {
		t.Fatal(err)
	}
	if !status.HasUpstream || status.Upstream != "origin/main" {
		t.Fatalf("expected upstream to remain visible, got %#v", status)
	}
	if status.HasDivergence {
		t.Fatalf("expected ahead/behind counts to be unavailable, got %#v", status)
	}
}

func TestInspectReturnsNonGitWithoutError(t *testing.T) {
	status, err := Inspect(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if status.Git {
		t.Fatalf("expected non-git status, got %#v", status)
	}
}

func TestInspectSurfacesWorktreeDetectionErrors(t *testing.T) {
	t.Setenv("PATH", "")

	_, err := Inspect(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "executable file not found") {
		t.Fatalf("expected missing git executable error, got %v", err)
	}
}

func TestParseAheadBehind(t *testing.T) {
	ahead, behind, ok := parseAheadBehind("12\t3\n")
	if !ok || ahead != 12 || behind != 3 {
		t.Fatalf("unexpected parse result: ahead=%d behind=%d ok=%v", ahead, behind, ok)
	}
	if _, _, ok := parseAheadBehind("bad"); ok {
		t.Fatal("expected parse failure")
	}
}

func initGitRepo(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "git", "init")
	configUser(t, root)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# repo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "git", "add", ".")
	gitRun(t, root, "git", "commit", "-m", "init")
}

func configUser(t *testing.T, dir string) {
	t.Helper()
	gitRun(t, dir, "git", "config", "user.email", "test@example.com")
	gitRun(t, dir, "git", "config", "user.name", "Test User")
}

func gitRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, string(out))
	}
}

func gitOutput(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, string(out))
	}
	return string(out)
}
