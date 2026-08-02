package coverage

import (
	"os"
	"testing"
)

const (
	moduleCoverageThreshold  = 0.90
	packageCoverageThreshold = 0.80
)

func TestEnforceThresholdsFromProfile(t *testing.T) {
	t.Parallel()

	profilePath := os.Getenv("UNDERSCORE_COVERAGE_PROFILE")
	if profilePath == "" {
		t.Skip("UNDERSCORE_COVERAGE_PROFILE is not set")
	}

	file, err := os.Open(profilePath)
	if err != nil {
		t.Fatalf("open coverage profile: %v", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Fatalf("close coverage profile: %v", closeErr)
		}
	}()

	profiles, err := ParseProfiles(file)
	if err != nil {
		t.Fatalf("parse coverage profile: %v", err)
	}

	if err := Check(moduleCoverageThreshold, packageCoverageThreshold, profiles); err != nil {
		t.Fatal(err)
	}
}
