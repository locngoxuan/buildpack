package core

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func Parse(s string) (v Version, err error) {
	parts := strings.Split(strings.TrimSpace(s), ".")
	if len(parts) != 3 {
		err = fmt.Errorf("can not parse string %s to version", s)
		return
	}
	v.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	v.Minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	v.Patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return
	}
	return
}

func (v *Version) NextPatch() {
	v.Patch = v.Patch + 1
}

func (v *Version) NextMinor() {
	v.Patch = 0
	v.Minor = v.Minor + 1
}

func (v *Version) NextMajor() {
	v.Patch = 0
	v.Minor = 0
	v.Major = v.Major + 1
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Version) MinorBranch() string {
	return fmt.Sprintf("%d.%d.x", v.Major, v.Minor)
}

func (v *Version) MajorBranch() string {
	return fmt.Sprintf("%d.x.x", v.Major)
}
