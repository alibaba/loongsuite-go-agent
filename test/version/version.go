// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// inspired by https://github.com/hashicorp/go-version

package version

import (
	"bytes"
	"fmt"
	"math/rand"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// The compiled regular expression used to test the validity of a version.
var (
	versionRegexp *regexp.Regexp
	semverRegexp  *regexp.Regexp
)

// The raw regular expression string used for testing the validity
// of a version.
const (
	VersionRegexpRaw string = `v?([0-9]+(\.[0-9]+)*?)` +
		`(-([0-9]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)|(-?([A-Za-z\-~]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)))?` +
		`(\+([0-9A-Za-z\-~]+(\.[0-9A-Za-z\-~]+)*))?` +
		`?`

	// SemverRegexpRaw requires a separator between version and pre-release
	SemverRegexpRaw string = `v?([0-9]+(\.[0-9]+)*?)` +
		`(-([0-9]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)|(-([A-Za-z\-~]+[0-9A-Za-z\-~]*(\.[0-9A-Za-z\-~]+)*)))?` +
		`(\+([0-9A-Za-z\-~]+(\.[0-9A-Za-z\-~]+)*))?` +
		`?`
)

// Version represents a single version.
type Version struct {
	metadata string
	pre      string
	segments []int64
	si       int
	original string
}

func init() {
	versionRegexp = regexp.MustCompile("^" + VersionRegexpRaw + "$")
	semverRegexp = regexp.MustCompile("^" + SemverRegexpRaw + "$")
}

// NewVersion parses the given version and returns a new
// Version.
func NewVersion(v string) (*Version, error) {
	return newVersion(v, versionRegexp)
}

func NewGoVersion(v string) (*Version, error) {
	goStart := 0
	for i := range v {
		if i+1 < len(v) && v[i] == 'g' && v[i+1] == 'o' {
			goStart = i + 2
			break
		}
	}
	if goStart >= len(v) {
		return nil, fmt.Errorf("invalid go version: %s", v)
	}
	return NewVersion(v[goStart:])
}

func newVersion(v string, pattern *regexp.Regexp) (*Version, error) {
	matches := pattern.FindStringSubmatch(v)
	if matches == nil {
		return nil, fmt.Errorf("Malformed version: %s", v)
	}
	segmentsStr := strings.Split(matches[1], ".")
	segments := make([]int64, len(segmentsStr))
	for i, str := range segmentsStr {
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, fmt.Errorf(
				"Error parsing version: %s", err)
		}

		segments[i] = val
	}

	// Even though we could support more than three segments, if we
	// got less than three, pad it with 0s. This is to cover the basic
	// default usecase of semver, which is MAJOR.MINOR.PATCH at the minimum
	for i := len(segments); i < 3; i++ {
		segments = append(segments, 0)
	}

	pre := matches[7]
	if pre == "" {
		pre = matches[4]
	}

	return &Version{
		metadata: matches[10],
		pre:      pre,
		segments: segments,
		si:       len(segmentsStr),
		original: v,
	}, nil
}

func (v *Version) Original() string {
	return v.original
}

// Compare compares this version to another version. This
// returns -1, 0, or 1 if this version is smaller, equal,
// or larger than the other version, respectively.
//
// If you want boolean results, use the LessThan, Equal,
// GreaterThan, GreaterThanOrEqual or LessThanOrEqual methods.
func (v *Version) Compare(other *Version) int {
	// A quick, efficient equality check
	if v.String() == other.String() {
		return 0
	}

	// If the segments are the same, we must compare on pre-release info
	if v.equalSegments(other) {
		preSelf := v.PreRelease()
		preOther := other.PreRelease()
		if preSelf == "" && preOther == "" {
			return 0
		}
		if preSelf == "" {
			return 1
		}
		if preOther == "" {
			return -1
		}

		return comparePreReleases(preSelf, preOther)
	}

	segmentsSelf := v.Segments64()
	segmentsOther := other.Segments64()
	// Get the highest specificity (hS), or if they're equal, just use segmentSelf length
	lenSelf := len(segmentsSelf)
	lenOther := len(segmentsOther)
	hS := lenSelf
	if lenSelf < lenOther {
		hS = lenOther
	}
	// Compare the segments
	// Because a constraint could have more/less specificity than the version it's
	// checking, we need to account for a lopsided or jagged comparison
	for i := 0; i < hS; i++ {
		if i > lenSelf-1 {
			// This means Self had the lower specificity
			// Check to see if the remaining segments in Other are all zeros
			if !allZero(segmentsOther[i:]) {
				// if not, it means that Other has to be greater than Self
				return -1
			}
			break
		} else if i > lenOther-1 {
			// this means Other had the lower specificity
			// Check to see if the remaining segments in Self are all zeros -
			if !allZero(segmentsSelf[i:]) {
				// if not, it means that Self has to be greater than Other
				return 1
			}
			break
		}
		lhs := segmentsSelf[i]
		rhs := segmentsOther[i]
		if lhs == rhs {
			continue
		} else if lhs < rhs {
			return -1
		}
		// Otherwise, rhs was > lhs, they're not equal
		return 1
	}

	// if we got this far, they're equal
	return 0
}

func (v *Version) equalSegments(other *Version) bool {
	segmentsSelf := v.Segments64()
	segmentsOther := other.Segments64()

	if len(segmentsSelf) != len(segmentsOther) {
		return false
	}
	for i, v := range segmentsSelf {
		if v != segmentsOther[i] {
			return false
		}
	}
	return true
}

func allZero(segs []int64) bool {
	for _, s := range segs {
		if s != 0 {
			return false
		}
	}
	return true
}

func comparePart(preSelf string, preOther string) int {
	if preSelf == preOther {
		return 0
	}

	var selfInt int64
	selfNumeric := true
	selfInt, err := strconv.ParseInt(preSelf, 10, 64)
	if err != nil {
		selfNumeric = false
	}

	var otherInt int64
	otherNumeric := true
	otherInt, err = strconv.ParseInt(preOther, 10, 64)
	if err != nil {
		otherNumeric = false
	}

	// if a part is empty, we use the other to decide
	if preSelf == "" {
		if otherNumeric {
			return -1
		}
		return 1
	}

	if preOther == "" {
		if selfNumeric {
			return 1
		}
		return -1
	}

	if selfNumeric && !otherNumeric {
		return -1
	} else if !selfNumeric && otherNumeric {
		return 1
	} else if !selfNumeric && !otherNumeric && preSelf > preOther {
		return 1
	} else if selfInt > otherInt {
		return 1
	}

	return -1
}

func comparePreReleases(v string, other string) int {
	// the same pre release!
	if v == other {
		return 0
	}

	// split both pre releases for analyze their parts
	selfPreReleaseMeta := strings.Split(v, ".")
	otherPreReleaseMeta := strings.Split(other, ".")

	selfPreReleaseLen := len(selfPreReleaseMeta)
	otherPreReleaseLen := len(otherPreReleaseMeta)

	biggestLen := otherPreReleaseLen
	if selfPreReleaseLen > otherPreReleaseLen {
		biggestLen = selfPreReleaseLen
	}

	// loop for parts to find the first difference
	for i := 0; i < biggestLen; i = i + 1 {
		partSelfPre := ""
		if i < selfPreReleaseLen {
			partSelfPre = selfPreReleaseMeta[i]
		}

		partOtherPre := ""
		if i < otherPreReleaseLen {
			partOtherPre = otherPreReleaseMeta[i]
		}

		compare := comparePart(partSelfPre, partOtherPre)
		// if parts are equals, continue the loop
		if compare != 0 {
			return compare
		}
	}

	return 0
}

// Equal tests if two versions are equal.
func (v *Version) Equal(o *Version) bool {
	if v == nil || o == nil {
		return v == o
	}

	return v.Compare(o) == 0
}

// GreaterThan tests if this version is greater than another version.
func (v *Version) GreaterThan(o *Version) bool {
	return v.Compare(o) > 0
}

// GreaterThanOrEqual tests if this version is greater than or equal to another version.
func (v *Version) GreaterThanOrEqual(o *Version) bool {
	return v.Compare(o) >= 0
}

// LessThan tests if this version is less than another version.
func (v *Version) LessThan(o *Version) bool {
	return v.Compare(o) < 0
}

// LessThanOrEqual tests if this version is less than or equal to another version.
func (v *Version) LessThanOrEqual(o *Version) bool {
	return v.Compare(o) <= 0
}

// Metadata returns any metadata that was part of the version
// string.
//
// Metadata is anything that comes after the "+" in the version.
// For example, with "1.2.3+beta", the metadata is "beta".
func (v *Version) Metadata() string {
	return v.metadata
}

// PreRelease returns any pre-release data that is part of the version,
// or blank if there is no pre-release data.
//
// PreRelease information is anything that comes after the "-" in the
// version (but before any metadata). For example, with "1.2.3-beta",
// the pre-release information is "beta".
func (v *Version) PreRelease() string {
	return v.pre
}

// Segments returns the numeric segments of the version as a slice of ints.
//
// This excludes any metadata or pre-release information. For example,
// for a version "1.2.3-beta", segments will return a slice of
// 1, 2, 3.
func (v *Version) Segments() []int {
	segmentSlice := make([]int, len(v.segments))
	for i, v := range v.segments {
		segmentSlice[i] = int(v)
	}
	return segmentSlice
}

// Segments64 returns the numeric segments of the version as a slice of int64s.
//
// This excludes any metadata or pre-release information. For example,
// for a version "1.2.3-beta", segments will return a slice of
// 1, 2, 3.
func (v *Version) Segments64() []int64 {
	result := make([]int64, len(v.segments))
	copy(result, v.segments)
	return result
}

// String returns the full version string included pre-release
// and metadata information.
//
// This value is rebuilt according to the parsed segments and other
// information. Therefore, ambiguities in the version string such as
// prefixed zeroes (1.04.0 => 1.4.0), `v` prefix (v1.0.0 => 1.0.0), and
// missing parts (1.0 => 1.0.0) will be made into a canonicalized form
// as shown in the parenthesized examples.
func (v *Version) String() string {
	var buf bytes.Buffer
	fmtParts := make([]string, len(v.segments))
	for i, s := range v.segments {
		// We can ignore err here since we've pre-parsed the values in segments
		str := strconv.FormatInt(s, 10)
		fmtParts[i] = str
	}
	fmt.Fprintf(&buf, strings.Join(fmtParts, "."))
	if v.pre != "" {
		fmt.Fprintf(&buf, "-%s", v.pre)
	}
	if v.metadata != "" {
		fmt.Fprintf(&buf, "+%s", v.metadata)
	}

	return buf.String()
}

func GetRandomVersion(versionNum int, moduleName string, minVersion, maxVersion *Version) ([]*Version, error) {
	versions, err := exec.Command("go", "list", "-m", "-mod=mod", "-versions", moduleName).Output()
	if err != nil {
		return nil, err
	}

	// trim the output and split by space to get individual versions
	versionStr := strings.TrimSpace(string(versions))
	versionArray := strings.Split(versionStr, " ")

	// first filter out all versions that meet the criteria
	validVersions := make([]*Version, 0, len(versionArray))
	for _, vStr := range versionArray {
		if vStr == "" {
			continue
		}
		v, err := NewVersion(vStr)
		if err != nil {
			continue
		}
		// skip pre-release versions (rc, alpha, beta, dep, etc.) to avoid
		// selecting broken releases that may have invalid transitive dependencies
		if v.PreRelease() != "" {
			continue
		}
		if minVersion != nil && v.LessThan(minVersion) {
			continue
		}
		if maxVersion != nil && v.GreaterThan(maxVersion) {
			continue
		}
		validVersions = append(validVersions, v)
	}

	// if the number of versions that meet the criteria is less than the requested quantity, return all versions that meet the criteria
	if len(validVersions) <= versionNum {
		return validVersions, nil
	}

	// randomly select from eligible versions and ensure uniqueness
	selected := make([]*Version, 0, versionNum)
	seen := make(map[string]bool)

	for len(selected) < versionNum {
		randomIndex := rand.Int() % len(validVersions)
		v := validVersions[randomIndex]
		versionStr := v.Original()

		if !seen[versionStr] {
			seen[versionStr] = true
			selected = append(selected, v)
		}
	}

	return selected, nil
}

func GetLatestVersion(moduleName string, minVersion, maxVersion *Version) (*Version, error) {
	versions, err := exec.Command("go", "list", "-m", "-mod=mod", "-versions", moduleName).Output()
	if err != nil {
		return nil, err
	}
	versionArray := strings.Split(string(versions), " ")
	v, err := NewVersion(strings.TrimSpace(versionArray[len(versionArray)-1]))
	if err != nil {
		return nil, err
	}
	if minVersion != nil && v.LessThan(minVersion) {
		return minVersion, nil
	}
	if maxVersion != nil && v.GreaterThan(maxVersion) {
		return maxVersion, nil
	}
	return v, nil
}
