package cli

import (
	"errors"
	"testing"

	"github.com/johnkil/polyrepo-workspace-kit/internal/scenario"
)

func TestScenarioExitErrorPrioritizesCommandFailure(t *testing.T) {
	err := scenarioExitError(scenario.RunResult{
		Failed:  true,
		Blocked: true,
		Drift:   true,
	})
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.Code != 5 {
		t.Fatalf("expected command failure exit code 5, got %d", exitErr.Code)
	}
}
