package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Path  int
}

func (v *Version) withoutLabel() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Path)
}

func (v *Version) withLabel(label string) string {
	return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Path, label)
}

func (v *Version) withLabelAndBuildNumber(label string, buildNumber int) string {
	return fmt.Sprintf("%d.%d.%d-%s.%d", v.Major, v.Minor, v.Path, label, buildNumber)
}

func (v *Version) nextPath() *Version {
	return &Version{
		Major: v.Major,
		Minor: v.Minor,
		Path:  v.Path + 1,
	}
}

func (v *Version) nextMinorVersion() *Version {
	return &Version{
		Major: v.Major,
		Minor: v.Minor + 1,
		Path:  0,
	}
}

func (v *Version) nextMajorVersion() *Version {
	return &Version{
		Major: v.Major + 1,
		Minor: 0,
		Path:  0,
	}
}

func (v *Version) branchBase() string {
	return fmt.Sprintf("%d.%d.x", v.Major, v.Minor)
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
	v.Path, err = strconv.Atoi(numbers[2])
	if err != nil {
		return nil, errors.New("path is not a number")
	}
	return v, nil
}
