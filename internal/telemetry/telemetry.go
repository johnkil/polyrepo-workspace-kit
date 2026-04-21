package telemetry

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/fsutil"
	"github.com/johnkil/polyrepo-workspace-kit/internal/manifest"
)

const (
	ConfigFile = "local/telemetry/config.yaml"
	EventsFile = "local/telemetry/events.jsonl"
)

type ConfigDocument struct {
	Version   int    `yaml:"version,omitempty"`
	Telemetry Config `yaml:"telemetry"`
}

type Config struct {
	Enabled   bool   `yaml:"enabled"`
	CreatedAt string `yaml:"created_at,omitempty"`
	UpdatedAt string `yaml:"updated_at,omitempty"`
}

type Status struct {
	Enabled    bool
	ConfigPath string
	EventsPath string
	EventCount int
}

type Event struct {
	Version    int      `json:"version"`
	Timestamp  string   `json:"timestamp"`
	Workspace  string   `json:"workspace"`
	Command    string   `json:"command"`
	Args       []string `json:"args,omitempty"`
	ExitCode   int      `json:"exit_code"`
	DurationMS int64    `json:"duration_ms"`
}

func Enable(root string, now time.Time) (Status, error) {
	if now.IsZero() {
		now = time.Now()
	}
	status, err := ReadStatus(root)
	if err != nil {
		return Status{}, err
	}
	createdAt := now.UTC().Format(time.RFC3339)
	if status.Enabled {
		doc, err := loadConfig(root)
		if err != nil {
			return Status{}, err
		}
		createdAt = doc.Telemetry.CreatedAt
	}
	doc := ConfigDocument{
		Version: 1,
		Telemetry: Config{
			Enabled:   true,
			CreatedAt: createdAt,
			UpdatedAt: now.UTC().Format(time.RFC3339),
		},
	}
	if err := writeConfig(root, doc); err != nil {
		return Status{}, err
	}
	return ReadStatus(root)
}

func Disable(root string, now time.Time) (Status, error) {
	if now.IsZero() {
		now = time.Now()
	}
	createdAt := ""
	if doc, err := loadConfig(root); err == nil {
		createdAt = doc.Telemetry.CreatedAt
	} else if !manifest.IsMissing(err) {
		return Status{}, err
	}
	doc := ConfigDocument{
		Version: 1,
		Telemetry: Config{
			Enabled:   false,
			CreatedAt: createdAt,
			UpdatedAt: now.UTC().Format(time.RFC3339),
		},
	}
	if err := writeConfig(root, doc); err != nil {
		return Status{}, err
	}
	return ReadStatus(root)
}

func ReadStatus(root string) (Status, error) {
	status, found, err := readConfigStatus(root)
	if err != nil {
		return status, err
	}
	if !found {
		return status, nil
	}
	count, err := countEvents(status.EventsPath)
	if err != nil {
		return status, err
	}
	status.EventCount = count
	return status, nil
}

func Export(root string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(root, EventsFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}

func RecordIfEnabled(root string, event Event) error {
	status, _, err := readConfigStatus(root)
	if err != nil {
		return err
	}
	if !status.Enabled {
		return nil
	}
	return recordEvent(root, status.EventsPath, event)
}

func Record(root string, event Event) error {
	status, _, err := readConfigStatus(root)
	if err != nil {
		return err
	}
	return recordEvent(root, status.EventsPath, event)
}

func recordEvent(root string, eventsPath string, event Event) error {
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if event.Version == 0 {
		event.Version = 1
	}
	if event.Workspace == "" {
		event.Workspace = root
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if err := fsutil.EnsureDir(filepath.Dir(eventsPath)); err != nil {
		return err
	}
	file, err := os.OpenFile(eventsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func readConfigStatus(root string) (Status, bool, error) {
	status := Status{
		ConfigPath: filepath.Join(root, ConfigFile),
		EventsPath: filepath.Join(root, EventsFile),
	}
	doc, err := loadConfig(root)
	if err != nil {
		if !manifest.IsMissing(err) {
			return status, false, err
		}
		return status, false, nil
	}
	status.Enabled = doc.Telemetry.Enabled
	return status, true, nil
}

func loadConfig(root string) (ConfigDocument, error) {
	var doc ConfigDocument
	err := manifest.LoadYAML(filepath.Join(root, ConfigFile), &doc)
	return doc, err
}

func writeConfig(root string, doc ConfigDocument) error {
	if doc.Version == 0 {
		doc.Version = 1
	}
	if err := fsutil.EnsureDir(filepath.Join(root, "local", "telemetry")); err != nil {
		return err
	}
	return manifest.WriteYAML(filepath.Join(root, ConfigFile), doc)
}

func countEvents(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	defer func() { _ = file.Close() }()
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		if scanner.Text() != "" {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("%s: %w", path, err)
	}
	return count, nil
}
