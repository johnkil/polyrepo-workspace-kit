package telemetry_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/johnkil/polyrepo-workspace-kit/internal/telemetry"
	"github.com/johnkil/polyrepo-workspace-kit/internal/workspace"
)

func TestEnableRecordExportAndDisable(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	status, err := telemetry.ReadStatus(root)
	if err != nil {
		t.Fatal(err)
	}
	if status.Enabled || status.EventCount != 0 {
		t.Fatalf("unexpected initial status: %#v", status)
	}

	if _, err := telemetry.Enable(root, time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if err := telemetry.RecordIfEnabled(root, telemetry.Event{
		Timestamp:  "2026-04-21T12:00:01Z",
		Command:    "wkit status",
		Args:       []string{"--context", "app-flow"},
		ExitCode:   0,
		DurationMS: 12,
	}); err != nil {
		t.Fatal(err)
	}

	data, err := telemetry.Export(root)
	if err != nil {
		t.Fatal(err)
	}
	var event telemetry.Event
	if err := json.Unmarshal(data[:len(data)-1], &event); err != nil {
		t.Fatal(err)
	}
	if event.Command != "wkit status" || event.Workspace != root || event.ExitCode != 0 {
		t.Fatalf("unexpected event: %#v", event)
	}
	status, err = telemetry.ReadStatus(root)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Enabled || status.EventCount != 1 {
		t.Fatalf("unexpected enabled status: %#v", status)
	}

	if _, err := telemetry.Disable(root, time.Date(2026, 4, 21, 12, 1, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if err := telemetry.RecordIfEnabled(root, telemetry.Event{
		Command:    "wkit doctor",
		ExitCode:   1,
		DurationMS: 5,
	}); err != nil {
		t.Fatal(err)
	}
	status, err = telemetry.ReadStatus(root)
	if err != nil {
		t.Fatal(err)
	}
	if status.Enabled || status.EventCount != 1 {
		t.Fatalf("disabled telemetry should not append events: %#v", status)
	}
}

func TestRecordIfEnabledDoesNotCountExistingEvents(t *testing.T) {
	root := filepath.Join(t.TempDir(), "workspace")
	if err := workspace.Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := telemetry.Enable(root, time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	eventsPath := filepath.Join(root, telemetry.EventsFile)
	if err := os.WriteFile(eventsPath, append(bytes.Repeat([]byte("x"), 70*1024), '\n'), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := telemetry.RecordIfEnabled(root, telemetry.Event{
		Timestamp:  "2026-04-21T12:00:01Z",
		Command:    "wkit info",
		ExitCode:   0,
		DurationMS: 3,
	}); err != nil {
		t.Fatal(err)
	}
	data, err := telemetry.Export(root)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(`"command":"wkit info"`)) {
		t.Fatalf("expected appended telemetry event, got:\n%s", string(data))
	}
}
