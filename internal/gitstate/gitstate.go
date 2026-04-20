package gitstate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
)

type State struct {
	Commit         string
	Short          string
	Branch         string
	Clean          bool
	DirtyPaths     []string
	UntrackedPaths []string
	Lockfiles      []model.LockfileHint
}

type CheckoutStatus struct {
	Git            bool
	Commit         string
	Short          string
	Branch         string
	Detached       bool
	Clean          bool
	DirtyPaths     []string
	UntrackedPaths []string
	Upstream       string
	Ahead          int
	Behind         int
	HasUpstream    bool
	HasDivergence  bool
}

func Version() string {
	out, err := run("", "git", "--version")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(out)
}

func Capture(repoPath string) (State, error) {
	commit, err := run(repoPath, "git", "rev-parse", "HEAD")
	if err != nil {
		return State{}, err
	}
	branch, err := run(repoPath, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return State{}, err
	}
	dirty, untracked, err := status(repoPath)
	if err != nil {
		return State{}, err
	}
	commit = strings.TrimSpace(commit)
	short := commit
	if len(short) > 8 {
		short = short[:8]
	}
	return State{
		Commit:         commit,
		Short:          short,
		Branch:         strings.TrimSpace(branch),
		Clean:          len(dirty) == 0 && len(untracked) == 0,
		DirtyPaths:     dirty,
		UntrackedPaths: untracked,
		Lockfiles:      LockfileHints(repoPath),
	}, nil
}

func Inspect(repoPath string) (CheckoutStatus, error) {
	inside, err := run(repoPath, "git", "rev-parse", "--is-inside-work-tree")
	if err != nil {
		if isNotGitError(err) {
			return CheckoutStatus{Git: false}, nil
		}
		return CheckoutStatus{}, err
	}
	if strings.TrimSpace(inside) != "true" {
		return CheckoutStatus{Git: false}, nil
	}

	state, err := Capture(repoPath)
	if err != nil {
		return CheckoutStatus{}, err
	}
	status := CheckoutStatus{
		Git:            true,
		Commit:         state.Commit,
		Short:          state.Short,
		Branch:         state.Branch,
		Detached:       state.Branch == "HEAD",
		Clean:          state.Clean,
		DirtyPaths:     state.DirtyPaths,
		UntrackedPaths: state.UntrackedPaths,
	}

	upstream, err := run(repoPath, "git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err != nil {
		return status, nil
	}
	status.Upstream = strings.TrimSpace(upstream)
	status.HasUpstream = status.Upstream != ""
	if !status.HasUpstream {
		return status, nil
	}

	counts, err := run(repoPath, "git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if err != nil {
		return status, nil
	}
	ahead, behind, ok := parseAheadBehind(counts)
	if !ok {
		return status, nil
	}
	status.Ahead = ahead
	status.Behind = behind
	status.HasDivergence = true
	return status, nil
}

func Head(repoPath string) (string, error) {
	out, err := run(repoPath, "git", "rev-parse", "HEAD")
	return strings.TrimSpace(out), err
}

func status(repoPath string) ([]string, []string, error) {
	out, err := run(repoPath, "git", "status", "--porcelain=v2", "-z")
	if err != nil {
		return nil, nil, err
	}
	dirty, untracked := parseStatusPorcelainV2Z([]byte(out))
	return dirty, untracked, nil
}

func parseStatusPorcelainV2Z(out []byte) ([]string, []string) {
	var dirty []string
	var untracked []string
	records := bytes.Split(out, []byte{0})
	for index := 0; index < len(records); index++ {
		record := string(records[index])
		if strings.TrimSpace(record) == "" {
			continue
		}
		switch record[0] {
		case '?':
			if path := strings.TrimPrefix(record, "? "); path != "" {
				untracked = append(untracked, path)
			}
		case '!':
			continue
		case '1':
			dirty = appendIfNotEmpty(dirty, porcelainV2Path(record, 9))
		case '2':
			path := porcelainV2Path(record, 10)
			original := ""
			if index+1 < len(records) {
				original = string(records[index+1])
				index++
			}
			if path != "" && original != "" {
				dirty = append(dirty, original+" -> "+path)
			} else {
				dirty = appendIfNotEmpty(dirty, path)
			}
		case 'u':
			dirty = appendIfNotEmpty(dirty, porcelainV2Path(record, 11))
		default:
			dirty = appendIfNotEmpty(dirty, strings.TrimSpace(record))
		}
	}
	return dirty, untracked
}

func porcelainV2Path(record string, fieldsWithPath int) string {
	parts := strings.SplitN(record, " ", fieldsWithPath)
	if len(parts) < fieldsWithPath {
		return strings.TrimSpace(record)
	}
	return parts[fieldsWithPath-1]
}

func appendIfNotEmpty(items []string, value string) []string {
	if value == "" {
		return items
	}
	return append(items, value)
}

func parseAheadBehind(out string) (int, int, bool) {
	fields := strings.Fields(out)
	if len(fields) != 2 {
		return 0, 0, false
	}
	ahead, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, false
	}
	behind, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, false
	}
	return ahead, behind, true
}

func isNotGitError(err error) bool {
	return strings.Contains(err.Error(), "not a git repository")
}

func run(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s: %s", name, msg)
	}
	return stdout.String(), nil
}

func LockfileHints(root string) []model.LockfileHint {
	names := []string{
		"package-lock.json",
		"pnpm-lock.yaml",
		"yarn.lock",
		"go.sum",
		"Cargo.lock",
		"poetry.lock",
		"uv.lock",
		"Gemfile.lock",
	}
	var out []model.LockfileHint
	for _, name := range names {
		path := filepath.Join(root, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		sum := sha256.Sum256(data)
		encoded := hex.EncodeToString(sum[:])
		out = append(out, model.LockfileHint{Path: name, SHA256: &encoded})
	}
	return out
}
