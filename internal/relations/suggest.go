package relations

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/johnkil/polyrepo-workspace-kit/internal/model"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

type Options struct {
	ContextID string
}

type Report struct {
	Suggestions []Suggestion
	Skipped     []SkippedRepo
}

type Suggestion struct {
	From     string
	To       string
	Kind     string
	Source   string
	Evidence string
	Matched  string
}

type SkippedRepo struct {
	Repo   string
	Reason string
}

type repoCandidate struct {
	ID         string
	Path       string
	Identities map[string]struct{}
	Deps       []dependency
}

type dependency struct {
	Name   string
	Kind   string
	Source string
}

func Suggest(root string, opts Options) (Report, error) {
	doc, err := workspace.LoadWorkspace(root)
	if err != nil {
		return Report{}, err
	}
	repoIDs := doc.Repos
	if strings.TrimSpace(opts.ContextID) != "" {
		contexts, err := workspace.LoadContexts(root)
		if err != nil {
			return Report{}, err
		}
		context, ok := contexts.Contexts[opts.ContextID]
		if !ok {
			return Report{}, fmt.Errorf("unknown context %q", opts.ContextID)
		}
		repoIDs = context.Repos
	}

	report := Report{}
	candidates := make([]repoCandidate, 0, len(repoIDs))
	for _, repoID := range repoIDs {
		checkout, err := workspace.ResolveRepoCheckout(root, repoID)
		if err != nil {
			report.Skipped = append(report.Skipped, SkippedRepo{Repo: repoID, Reason: err.Error()})
			continue
		}
		candidate := repoCandidate{
			ID:         repoID,
			Path:       checkout,
			Identities: map[string]struct{}{},
		}
		addIdentity(candidate.Identities, repoID)
		addIdentity(candidate.Identities, filepath.Base(checkout))
		manifestInfo, err := scanManifests(checkout)
		if err != nil {
			report.Skipped = append(report.Skipped, SkippedRepo{Repo: repoID, Reason: err.Error()})
			continue
		}
		for _, identity := range manifestInfo.Identities {
			addIdentity(candidate.Identities, identity)
		}
		candidate.Deps = manifestInfo.Deps
		candidates = append(candidates, candidate)
	}

	existing := existingRelations(doc.Relations)
	seen := map[string]struct{}{}
	for _, from := range candidates {
		for _, dep := range from.Deps {
			for _, to := range candidates {
				if from.ID == to.ID {
					continue
				}
				matched, ok := matchDependency(dep.Name, to.Identities)
				if !ok {
					continue
				}
				key := relationKey(from.ID, to.ID, dep.Kind)
				if _, ok := existing[key]; ok {
					continue
				}
				suggestionKey := strings.Join([]string{from.ID, to.ID, dep.Kind, dep.Source, dep.Name, matched}, "\x00")
				if _, ok := seen[suggestionKey]; ok {
					continue
				}
				seen[suggestionKey] = struct{}{}
				report.Suggestions = append(report.Suggestions, Suggestion{
					From:     from.ID,
					To:       to.ID,
					Kind:     dep.Kind,
					Source:   dep.Source,
					Evidence: dep.Name,
					Matched:  matched,
				})
			}
		}
	}
	sort.Slice(report.Suggestions, func(i, j int) bool {
		left := report.Suggestions[i]
		right := report.Suggestions[j]
		return strings.Join([]string{left.From, left.To, left.Kind, left.Source, left.Evidence}, "\x00") <
			strings.Join([]string{right.From, right.To, right.Kind, right.Source, right.Evidence}, "\x00")
	})
	sort.Slice(report.Skipped, func(i, j int) bool {
		return report.Skipped[i].Repo < report.Skipped[j].Repo
	})
	return report, nil
}

type manifestInfo struct {
	Identities []string
	Deps       []dependency
}

func scanManifests(root string) (manifestInfo, error) {
	var out manifestInfo
	if info, err := scanGoMod(filepath.Join(root, "go.mod")); err != nil {
		return out, err
	} else {
		out.merge(info)
	}
	if info, err := scanPackageJSON(filepath.Join(root, "package.json")); err != nil {
		return out, err
	} else {
		out.merge(info)
	}
	if info, err := scanCargoToml(filepath.Join(root, "Cargo.toml")); err != nil {
		return out, err
	} else {
		out.merge(info)
	}
	for _, rel := range []string{"settings.gradle", "settings.gradle.kts", "build.gradle", "build.gradle.kts"} {
		if info, err := scanGradle(filepath.Join(root, rel)); err != nil {
			return out, err
		} else {
			out.merge(info)
		}
	}
	return out, nil
}

func (m *manifestInfo) merge(other manifestInfo) {
	m.Identities = append(m.Identities, other.Identities...)
	m.Deps = append(m.Deps, other.Deps...)
}

func scanGoMod(path string) (manifestInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return manifestInfo{}, nil
		}
		return manifestInfo{}, err
	}
	defer func() { _ = file.Close() }()

	var out manifestInfo
	inRequireBlock := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := stripGoComment(strings.TrimSpace(scanner.Text()))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "module ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				out.Identities = append(out.Identities, strings.Trim(fields[1], `"`))
			}
			continue
		}
		if line == "require (" {
			inRequireBlock = true
			continue
		}
		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}
		if strings.HasPrefix(line, "require ") {
			fields := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(fields) >= 1 {
				out.Deps = append(out.Deps, dependency{Name: strings.Trim(fields[0], `"`), Kind: "runtime", Source: "go.mod"})
			}
			continue
		}
		if inRequireBlock {
			fields := strings.Fields(line)
			if len(fields) >= 1 {
				out.Deps = append(out.Deps, dependency{Name: strings.Trim(fields[0], `"`), Kind: "runtime", Source: "go.mod"})
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return out, err
	}
	return out, nil
}

func scanPackageJSON(path string) (manifestInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return manifestInfo{}, nil
		}
		return manifestInfo{}, err
	}
	var doc struct {
		Name                 string            `json:"name"`
		Dependencies         map[string]string `json:"dependencies"`
		DevDependencies      map[string]string `json:"devDependencies"`
		PeerDependencies     map[string]string `json:"peerDependencies"`
		OptionalDependencies map[string]string `json:"optionalDependencies"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return manifestInfo{}, fmt.Errorf("%s: %w", path, err)
	}
	out := manifestInfo{Identities: []string{doc.Name}}
	for name := range doc.Dependencies {
		out.Deps = append(out.Deps, dependency{Name: name, Kind: "runtime", Source: "package.json dependencies"})
	}
	for name := range doc.PeerDependencies {
		out.Deps = append(out.Deps, dependency{Name: name, Kind: "runtime", Source: "package.json peerDependencies"})
	}
	for name := range doc.OptionalDependencies {
		out.Deps = append(out.Deps, dependency{Name: name, Kind: "runtime", Source: "package.json optionalDependencies"})
	}
	for name := range doc.DevDependencies {
		out.Deps = append(out.Deps, dependency{Name: name, Kind: "build", Source: "package.json devDependencies"})
	}
	return out, nil
}

func scanCargoToml(path string) (manifestInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return manifestInfo{}, nil
		}
		return manifestInfo{}, err
	}
	defer func() { _ = file.Close() }()

	var out manifestInfo
	section := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := stripTomlComment(strings.TrimSpace(scanner.Text()))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.Trim(line, "[]")
			continue
		}
		key, _, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.Trim(strings.TrimSpace(key), `"`)
		switch section {
		case "package":
			if key == "name" {
				_, value, _ := strings.Cut(line, "=")
				out.Identities = append(out.Identities, trimQuoted(value))
			}
		case "dependencies":
			out.Deps = append(out.Deps, dependency{Name: key, Kind: "runtime", Source: "Cargo.toml dependencies"})
		case "dev-dependencies", "build-dependencies":
			out.Deps = append(out.Deps, dependency{Name: key, Kind: "build", Source: "Cargo.toml " + section})
		}
	}
	if err := scanner.Err(); err != nil {
		return out, err
	}
	return out, nil
}

func scanGradle(path string) (manifestInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return manifestInfo{}, nil
		}
		return manifestInfo{}, err
	}
	out := manifestInfo{}
	source := filepath.Base(path)
	lines := strings.Split(string(data), "\n")
	for _, raw := range lines {
		line := stripGradleComment(strings.TrimSpace(raw))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "rootProject.name") {
			if _, value, ok := strings.Cut(line, "="); ok {
				out.Identities = append(out.Identities, trimQuoted(value))
			}
			continue
		}
		conf, value, ok := gradleDependency(line)
		if ok {
			out.Deps = append(out.Deps, dependency{Name: value, Kind: gradleRelationKind(conf), Source: source})
			continue
		}
	}
	return out, nil
}

func gradleDependency(line string) (string, string, bool) {
	conf := ""
	value := ""
	if open := strings.Index(line, "("); open > 0 {
		candidate := strings.TrimSpace(line[:open])
		if isGradleConfiguration(candidate) {
			close := strings.LastIndex(line, ")")
			if close <= open {
				return "", "", false
			}
			conf = candidate
			value = line[open+1 : close]
		}
	}
	if conf == "" {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return "", "", false
		}
		conf = fields[0]
		value = strings.TrimSpace(strings.TrimPrefix(line, conf))
	}
	if !isGradleConfiguration(conf) {
		return "", "", false
	}
	value = gradleDependencyValue(value)
	if value == "" {
		return "", "", false
	}
	return conf, value, true
}

func isGradleConfiguration(value string) bool {
	switch value {
	case "implementation", "api", "runtimeOnly", "compileOnly", "testImplementation", "testRuntimeOnly", "annotationProcessor", "kapt":
		return true
	default:
		return false
	}
}

func gradleRelationKind(conf string) string {
	switch conf {
	case "implementation", "api", "runtimeOnly":
		return "runtime"
	default:
		return "build"
	}
}

func gradleDependencyValue(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "project(") {
		return gradleProjectName(value)
	}
	if name, ok := gradleNamedArg(value, "name"); ok {
		return gradleCleanDependencyName(name)
	}
	return gradleCleanDependencyName(gradleFirstArg(value))
}

func gradleProjectName(value string) string {
	name := between(value, "project(", ")")
	if path, ok := gradleNamedArg(name, "path"); ok {
		return gradleCleanDependencyName(path)
	}
	return gradleCleanDependencyName(gradleFirstArg(name))
}

func gradleCleanDependencyName(value string) string {
	return strings.TrimPrefix(trimQuoted(value), ":")
}

func gradleFirstArg(value string) string {
	value = strings.TrimSpace(value)
	if left, _, ok := strings.Cut(value, ","); ok {
		return strings.TrimSpace(left)
	}
	return value
}

func gradleNamedArg(value string, key string) (string, bool) {
	for _, arg := range strings.Split(value, ",") {
		arg = strings.TrimSpace(arg)
		for _, separator := range []string{"=", ":"} {
			left, right, ok := strings.Cut(arg, separator)
			if ok && strings.TrimSpace(left) == key {
				return strings.TrimSpace(right), true
			}
		}
	}
	return "", false
}

func matchDependency(dep string, identities map[string]struct{}) (string, bool) {
	dep = strings.TrimSpace(dep)
	if dep == "" {
		return "", false
	}
	depParts := splitDependencyName(dep)
	for identity := range identities {
		if identity == "" {
			continue
		}
		if dep == identity {
			return identity, true
		}
		if strings.HasSuffix(dep, "/"+identity) {
			return identity, true
		}
		if depParts.artifact != "" && depParts.artifact == identity {
			return identity, true
		}
		if depParts.lastSegment != "" && depParts.lastSegment == identity {
			return identity, true
		}
	}
	return "", false
}

type dependencyParts struct {
	artifact    string
	lastSegment string
}

func splitDependencyName(dep string) dependencyParts {
	out := dependencyParts{}
	if strings.Contains(dep, ":") {
		parts := strings.Split(dep, ":")
		if len(parts) >= 2 {
			out.artifact = strings.TrimSpace(parts[1])
		}
	}
	trimmed := strings.Trim(dep, "/")
	if trimmed != "" {
		parts := strings.Split(trimmed, "/")
		out.lastSegment = parts[len(parts)-1]
	}
	return out
}

func existingRelations(relations []model.Relation) map[string]struct{} {
	out := map[string]struct{}{}
	for _, relation := range relations {
		out[relationKey(relation.From, relation.To, relation.Kind)] = struct{}{}
	}
	return out
}

func relationKey(from string, to string, kind string) string {
	return from + "\x00" + to + "\x00" + kind
}

func addIdentity(identities map[string]struct{}, identity string) {
	identity = strings.TrimSpace(identity)
	if identity == "" {
		return
	}
	identities[identity] = struct{}{}
	parts := splitDependencyName(identity)
	if parts.artifact != "" {
		identities[parts.artifact] = struct{}{}
	}
	if parts.lastSegment != "" {
		identities[parts.lastSegment] = struct{}{}
	}
}

func stripGoComment(line string) string {
	if index := strings.Index(line, "//"); index >= 0 {
		return strings.TrimSpace(line[:index])
	}
	return line
}

func stripTomlComment(line string) string {
	if index := strings.Index(line, "#"); index >= 0 {
		return strings.TrimSpace(line[:index])
	}
	return line
}

func stripGradleComment(line string) string {
	if index := strings.Index(line, "//"); index >= 0 {
		return strings.TrimSpace(line[:index])
	}
	return line
}

func trimQuoted(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ",")
	value = strings.TrimSpace(value)
	return strings.Trim(value, `"'`)
}

func between(value string, prefix string, suffix string) string {
	start := strings.Index(value, prefix)
	if start < 0 {
		return ""
	}
	start += len(prefix)
	end := strings.Index(value[start:], suffix)
	if end < 0 {
		return ""
	}
	return value[start : start+end]
}
