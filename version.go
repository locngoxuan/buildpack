package buildpack

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v *Version) withoutLabel() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Version) withLabel(label string) string {
	return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, label)
}

func (v *Version) withLabelAndBuildNumber(label string, buildNumber int) string {
	return fmt.Sprintf("%d.%d.%d-%s.%d", v.Major, v.Minor, v.Patch, label, buildNumber)
}

func (v *Version) nextPatch() {
	v.Patch = v.Patch + 1
}

func (v *Version) nextMinorVersion() {
	v.Patch = 0
	v.Minor = v.Minor + 1
}

func (v *Version) nextMajorVersion() {
	v.Patch = 0
	v.Minor = 0
	v.Major = v.Major + 1
}

func (v *Version) branchBaseMinor() string {
	return fmt.Sprintf("%d.%d.x", v.Major, v.Minor)
}

func (v *Version) branchBaseMajor() string {
	return fmt.Sprintf("%d.x.x", v.Major)
}

// convert from major.minor.path-label.buildNumber to version
func fromString(version string) (*Version, error) {
	numbers := strings.Split(version, ".")
	if len(numbers) != 3 {
		return nil, errors.New("invalid number format")
	}

	v := &Version{}
	var err error
	v.Major, err = strconv.Atoi(numbers[0])
	if err != nil {
		return nil, errors.New("major is not a number")
	}
	v.Minor, err = strconv.Atoi(numbers[1])
	if err != nil {
		return nil, errors.New("minor is not a number")
	}
	v.Patch, err = strconv.Atoi(numbers[2])
	if err != nil {
		return nil, errors.New("path is not a number")
	}
	return v, nil
}
