package steps

import (
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

func TestMain(m *testing.M) {
	// Suppress service-layer logs during tests for cleaner output.
	// Chi middleware logs to stdout separately — those are fine.
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "unified-service-scheduler",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed")
	}
}
