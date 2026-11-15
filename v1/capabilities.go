// SPDX-License-Identifier: Apache-2.0

package v1

import "time"

// Capabilities are decided per TV by evidence, never by platform or brand
// name. A control plane automates an action on a TV only when the matching
// capability is supported there; unknown counts as unsupported.
const (
	// CapGuestbookDisplay: the Hospitium launcher (or guestbook app) runs and
	// shows guest content.
	CapGuestbookDisplay = "guestbook_display"
	// CapCustomLauncher: HOME resolves to the Hospitium launcher.
	CapCustomLauncher = "custom_launcher"
	// CapLauncherPersistsReboot: HOME still resolves to the launcher after a
	// reboot, before the agent re-applies activation.
	CapLauncherPersistsReboot = "launcher_persists_reboot"
	// CapHomeButtonReturnsLauncher: pressing HOME on the remote returns to the
	// launcher. Observed by the launcher or by a person, not by the agent.
	CapHomeButtonReturnsLauncher = "home_button_returns_launcher"
	// CapRemoteAppInstall: the agent can install or update an APK.
	CapRemoteAppInstall = "remote_app_install"
	// CapAppDataCleanup: clearing an app's data works and is verified. One
	// observation per app package (CapabilityObservation.Subject).
	CapAppDataCleanup = "app_data_cleanup"
	// CapRemoteReboot: the agent can reboot the TV.
	CapRemoteReboot = "remote_reboot"
	// CapHealthMonitoring: the agent can reach the TV and read its state.
	CapHealthMonitoring = "health_monitoring"
)

// Capability states.
const (
	CapSupported   = "supported"
	CapUnsupported = "unsupported"
	CapUnknown     = "unknown"
)

// Evidence says where an observation came from. Agents report only
// EvidenceVerified (seen on real hardware) or EvidenceSimulated (seen on a
// simulated TV). The others are recorded by a control plane or a public
// compatibility database.
const (
	EvidenceVerified  = "verified"
	EvidenceSimulated = "simulated"
	// EvidenceCommunity: a compatibility database submission for the same
	// model and build.
	EvidenceCommunity = "community"
	// EvidenceDocumented: vendor documentation. Never enough to automate.
	EvidenceDocumented = "documented"
)

// Drivers name how an agent talks to a TV.
const (
	// DriverAndroidADB covers Google TV, Android TV and Fire OS over ADB.
	DriverAndroidADB = "android-adb"
	// DriverSimulator is a simulated TV for tests and demos.
	DriverSimulator = "simulator"
)

// DeviceFacts is what an agent reads from a TV about the TV itself. It
// identifies the hardware and firmware that capability observations apply
// to: an observation made on one BuildFingerprint says nothing about another.
type DeviceFacts struct {
	Driver       string `json:"driver"`
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
	OSVersion    string `json:"os_version,omitempty"`
	SDKLevel     int    `json:"sdk_level,omitempty"`
	// BuildFingerprint is Android's ro.build.fingerprint, or the platform's
	// equivalent firmware identifier.
	BuildFingerprint string `json:"build_fingerprint,omitempty"`
	// HomeActivity is the component HOME resolves to right now.
	HomeActivity string `json:"home_activity,omitempty"`
}

// CapabilityObservation is one thing seen to work, or not, on one TV.
// Agents report only what they observed directly; they never infer a result
// from the model or platform.
type CapabilityObservation struct {
	Capability string `json:"capability"`
	// Subject narrows the capability, e.g. the app package for
	// CapAppDataCleanup. Empty when the capability is device-wide.
	Subject  string `json:"subject,omitempty"`
	State    string `json:"state"`
	Evidence string `json:"evidence"`
	// Detail is a short, human-readable note. It must never contain guest
	// data, tokens or secrets.
	Detail     string    `json:"detail,omitempty"`
	ObservedAt time.Time `json:"observed_at"`
}
