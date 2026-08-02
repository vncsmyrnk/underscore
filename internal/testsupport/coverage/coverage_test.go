package coverage

import (
	"strings"
	"testing"
)

func TestCheckFailsWhenModuleCoverageIsBelowThreshold(t *testing.T) {
	t.Parallel()

	profiles := []PackageProfile{
		{Package: "internal/foo", Statements: 100, Covered: 89},
	}

	err := Check(0.90, 0.80, profiles)
	if err == nil {
		t.Fatal("expected module coverage failure")
	}
}

func TestCheckFailsWhenPackageCoverageIsBelowThreshold(t *testing.T) {
	t.Parallel()

	profiles := []PackageProfile{
		{Package: "internal/foo", Statements: 50, Covered: 50},
		{Package: "internal/bar", Statements: 100, Covered: 79},
	}

	err := Check(0.90, 0.80, profiles)
	if err == nil {
		t.Fatal("expected package coverage failure")
	}
}

func TestCheckIgnoresExcludedPackages(t *testing.T) {
	t.Parallel()

	profiles := []PackageProfile{
		{Package: "internal/foo", Statements: 90, Covered: 90},
		{Package: "internal/testsupport/commandtest", Statements: 100, Covered: 0, Excluded: true},
	}

	if err := Check(0.90, 0.80, profiles); err != nil {
		t.Fatalf("expected excluded package to be ignored: %v", err)
	}
}

func TestCheckPassesWhenThresholdsAreMet(t *testing.T) {
	t.Parallel()

	profiles := []PackageProfile{
		{Package: "internal/foo", Statements: 90, Covered: 90},
		{Package: "internal/bar", Statements: 100, Covered: 85},
	}

	if err := Check(0.90, 0.80, profiles); err != nil {
		t.Fatalf("expected thresholds to pass: %v", err)
	}
}

func TestCheckFailsWhenNoIncludedStatementsExist(t *testing.T) {
	t.Parallel()

	profiles := []PackageProfile{
		{Package: "internal/testsupport/commandtest", Statements: 10, Covered: 0, Excluded: true},
	}

	err := Check(0.90, 0.80, profiles)
	if err == nil {
		t.Fatal("expected undefined module coverage failure")
	}
}

func TestParseProfilesGroupsStatementsByPackage(t *testing.T) {
	t.Parallel()

	report := strings.NewReader(`mode: set
internal/testsupport/coverage/coverage.go:10.1,12.2 2 1
internal/testsupport/coverage/coverage.go:14.1,18.2 3 0
cmd/coveragecheck/main.go:5.1,6.2 1 1
internal/testsupport/commandtest/main.go:5.1,6.2 1 0
`)

	profiles, err := ParseProfiles(report)
	if err != nil {
		t.Fatalf("ParseProfiles returned error: %v", err)
	}

	if len(profiles) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(profiles))
	}

	for _, profile := range profiles {
		switch profile.Package {
		case "internal/testsupport/coverage":
			if profile.Statements != 5 || profile.Covered != 2 {
				t.Fatalf("unexpected coverage profile: %+v", profile)
			}
			if profile.Excluded {
				t.Fatalf("coverage package should not be excluded: %+v", profile)
			}
		case "cmd/coveragecheck":
			if !profile.Excluded {
				t.Fatalf("cmd entry point should be excluded: %+v", profile)
			}
		case "internal/testsupport/commandtest":
			if !profile.Excluded {
				t.Fatalf("command test helper should be excluded: %+v", profile)
			}
		default:
			t.Fatalf("unexpected package %q", profile.Package)
		}
	}
}

func TestParseProfilesRejectsInvalidHeader(t *testing.T) {
	t.Parallel()

	_, err := ParseProfiles(strings.NewReader("invalid\n"))
	if err == nil {
		t.Fatal("expected invalid header error")
	}
}

func TestParseProfilesRejectsInvalidCoverageLine(t *testing.T) {
	t.Parallel()

	report := strings.NewReader(`mode: set
internal/testsupport/coverage/coverage.go:10.1,12.2 nope 1
`)

	_, err := ParseProfiles(report)
	if err == nil {
		t.Fatal("expected invalid coverage line error")
	}
}

func TestIsExcludedPackage(t *testing.T) {
	t.Parallel()

	if !IsExcludedPackage("cmd/coveragecheck") {
		t.Fatal("expected cmd package to be excluded")
	}

	if !IsExcludedPackage("internal/testsupport/commandtest") {
		t.Fatal("expected commandtest helper to be excluded")
	}

	if !IsExcludedPackage("github.com/vncsmyrnk/underscore/internal/testsupport/commandtest") {
		t.Fatal("expected full import path for commandtest helper to be excluded")
	}

	if IsExcludedPackage("internal/testsupport/coverage") {
		t.Fatal("expected coverage package to remain included")
	}

}
