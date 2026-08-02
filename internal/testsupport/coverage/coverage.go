package coverage

import (
	"bufio"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

type PackageProfile struct {
	Package    string
	Statements int
	Covered    int
	Excluded   bool
}

func Check(moduleThreshold float64, packageThreshold float64, profiles []PackageProfile) error {
	moduleStatements := 0
	moduleCovered := 0

	for _, profile := range profiles {
		if profile.Excluded {
			continue
		}

		if profile.Statements == 0 {
			continue
		}

		packageCoverage := float64(profile.Covered) / float64(profile.Statements)
		if packageCoverage < packageThreshold {
			return fmt.Errorf(
				"package %s coverage %.2f%% is below %.2f%%",
				profile.Package,
				packageCoverage*100,
				packageThreshold*100,
			)
		}

		moduleStatements += profile.Statements
		moduleCovered += profile.Covered
	}

	if moduleStatements == 0 {
		return fmt.Errorf("module coverage is undefined")
	}

	moduleCoverage := float64(moduleCovered) / float64(moduleStatements)
	if moduleCoverage < moduleThreshold {
		return fmt.Errorf(
			"module coverage %.2f%% is below %.2f%%",
			moduleCoverage*100,
			moduleThreshold*100,
		)
	}

	return nil
}

func ParseProfiles(reader io.Reader) ([]PackageProfile, error) {
	scanner := bufio.NewScanner(reader)
	lineNumber := 0
	byPackage := map[string]*PackageProfile{}

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if lineNumber == 1 {
			if !strings.HasPrefix(line, "mode: ") {
				return nil, fmt.Errorf("invalid coverage header %q", line)
			}
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid coverage line %q", line)
		}

		filename, statements, count, err := parseCoverageLine(fields)
		if err != nil {
			return nil, err
		}

		pkg := path.Dir(filename)
		profile := byPackage[pkg]
		if profile == nil {
			profile = &PackageProfile{
				Package:  pkg,
				Excluded: IsExcludedPackage(pkg),
			}
			byPackage[pkg] = profile
		}

		profile.Statements += statements
		if count > 0 {
			profile.Covered += statements
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	profiles := make([]PackageProfile, 0, len(byPackage))
	for _, profile := range byPackage {
		profiles = append(profiles, *profile)
	}

	return profiles, nil
}

func IsExcludedPackage(pkg string) bool {
	return strings.HasPrefix(pkg, "cmd/") ||
		strings.Contains(pkg, "/cmd/") ||
		pkg == "internal/testsupport/commandtest" ||
		strings.HasSuffix(pkg, "/internal/testsupport/commandtest")
}

func parseCoverageLine(fields []string) (string, int, int, error) {
	fileAndRange := fields[0]
	file, _, found := strings.Cut(fileAndRange, ":")
	if !found {
		return "", 0, 0, fmt.Errorf("invalid coverage location %q", fileAndRange)
	}

	statements, err := strconv.Atoi(fields[1])
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid statement count %q: %w", fields[1], err)
	}

	count, err := strconv.Atoi(fields[2])
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid execution count %q: %w", fields[2], err)
	}

	return file, statements, count, nil
}
