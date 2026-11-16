// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Guestbook usage (added in v0.6.0): the launcher counts which screens are opened and which apps
// are launched from it, per hour, and nothing else. No text, no finer timestamps, no dwell time, no
// identifiers of the guest's. Control planes keep the counts anonymous: nothing about a stay or a
// guest is attached to them.

// PathLauncherUsage takes a UsageReport with the launcher's device token.
const PathLauncherUsage = "/v1/launcher/usage"

// Usage counters. An announcement's counter carries its ID; an app's, its package name.
const (
	UsageHome               = "home"        // the home screen came to the front
	UsageWiFi               = "wifi"        // the Wi-Fi card was opened
	UsageHouseRules         = "house_rules" // the house rules card was opened
	UsageAmenities          = "amenities"   // the amenities card was opened
	UsageAnnouncementPrefix = "announcement:"
	UsageAppPrefix          = "app:"
)

// Bounds a control plane enforces on a report.
const (
	MaxUsageHours    = 168 // a week of hours in one report
	MaxUsageCounters = 64  // counters in one hour
	MaxUsageCount    = 10000
)

// UsageReport carries counts for one or more hours.
//
// Each count is the TOTAL for its counter in that hour so far, not what changed since the last
// report. A control plane keeps the larger of what it has and what it receives, so sending the
// same hour again (after a timeout, or as it fills up) never counts twice. A launcher can drop an
// hour once a report containing its final total has been accepted.
type UsageReport struct {
	Hours []UsageHour `json:"hours"`
}

// UsageHour is the counts for one hour. Hour is the start of a UTC hour, by the launcher's
// best estimate of real time (its clock corrected by the server's).
type UsageHour struct {
	Hour   time.Time      `json:"hour"`
	Counts map[string]int `json:"counts"`
}

var packageName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*(\.[A-Za-z][A-Za-z0-9_]*)+$`)

// ValidUsageCounter reports whether name is a counter this version defines.
func ValidUsageCounter(name string) bool {
	switch name {
	case UsageHome, UsageWiFi, UsageHouseRules, UsageAmenities:
		return true
	}
	if id, ok := strings.CutPrefix(name, UsageAnnouncementPrefix); ok {
		n, err := strconv.ParseInt(id, 10, 64)
		return err == nil && n > 0 && strconv.FormatInt(n, 10) == id
	}
	if pkg, ok := strings.CutPrefix(name, UsageAppPrefix); ok {
		return len(pkg) <= 255 && packageName.MatchString(pkg)
	}
	return false
}
