// SPDX-License-Identifier: Apache-2.0

// Package v1 defines version 1 of the Hospitium device-management and
// launcher protocols. These types are the wire contract between a Hospitium
// control plane and:
//
//   - the management agent (Raspberry Pi), which pulls desired state for the
//     TVs it manages, reconciles them over ADB, and reports observed state;
//   - the TV launcher, which pulls its own guest-facing configuration.
//
// The protocol is pull-only: devices initiate every request over HTTPS, so no
// inbound connectivity to a property network is ever required.
//
// Compatibility rule: within v1, fields may be added but never removed or
// re-typed. Consumers must ignore unknown fields.
package v1

import (
	"encoding/json"
	"time"
)

// Agent endpoints (relative to the control-plane base URL).
const (
	PathAgentEnroll          = "/v1/agent/enroll"
	PathAgentDesiredState    = "/v1/agent/desired-state"
	PathAgentStatus          = "/v1/agent/status"
	PathAgentOperationResult = "/v1/agent/operations/{id}/result"
)

// EnrollRequest exchanges a single-use enrollment token (created by an
// administrator for a specific property) for a long-lived agent credential.
type EnrollRequest struct {
	EnrollmentToken string `json:"enrollment_token"`
	Hostname        string `json:"hostname"`
	AgentVersion    string `json:"agent_version"`
	Hardware        string `json:"hardware,omitempty"` // e.g. "Raspberry Pi 3 Model B Rev 1.2"
}

// EnrollResponse carries the agent credential. The token is returned exactly
// once; the control plane stores only a hash.
type EnrollResponse struct {
	ControllerID int64  `json:"controller_id"`
	AgentToken   string `json:"agent_token"`
	PropertyID   int64  `json:"property_id"`
}

// DesiredStateResponse is returned to an authenticated agent on every poll.
type DesiredStateResponse struct {
	ControllerID int64 `json:"controller_id"`
	// PollIntervalSeconds is the control plane's suggested base interval.
	// Agents add jitter and may back off further on errors.
	PollIntervalSeconds int                `json:"poll_interval_seconds"`
	Devices             []DeviceAssignment `json:"devices"`
	// Operations are one-time imperative commands. They are delivered
	// at-least-once; agents must deduplicate by Operation.ID.
	Operations []Operation `json:"operations"`
	ServerTime time.Time   `json:"server_time"`
}

// DeviceAssignment is one TV the agent is responsible for.
type DeviceAssignment struct {
	DeviceID   int64  `json:"device_id"`
	Name       string `json:"name"`
	ADBAddress string `json:"adb_address"` // host:port on the management network
	// Generation increments whenever Spec changes. Agents echo the last
	// generation they fully reconciled as ObservedGeneration.
	Generation int64      `json:"generation"`
	Spec       DeviceSpec `json:"spec"`
}

// DeviceSpec is the declarative desired state of a TV. Reconciliation must
// be idempotent: applying a spec to a compliant device makes no changes.
type DeviceSpec struct {
	Launcher     *LauncherSpec `json:"launcher,omitempty"`
	RequiredApps []AppSpec     `json:"required_apps,omitempty"`
	Settings     []Setting     `json:"settings,omitempty"`
	// CleanupPackages lists streaming apps whose data is cleared by
	// checkout cleanup operations. It is never acted on during
	// reconciliation.
	CleanupPackages []string `json:"cleanup_packages,omitempty"`
}

// Launcher activation strategies. See the launcher activation notes in the
// agent repository for device-specific behavior and recovery procedures.
const (
	// ActivationNone installs the launcher but leaves the stock home screen.
	ActivationNone = "none"
	// ActivationHomeActivity sets the default home activity via
	// `cmd package set-home-activity`. Supported where the OS honors it.
	ActivationHomeActivity = "home_activity"
	// ActivationDisableStock disables the stock launcher package(s) for
	// user 0 so the Hospitium launcher becomes the only HOME handler.
	// ADB-dependent workaround; fully reversible with `pm enable`.
	ActivationDisableStock = "disable_stock"
)

type LauncherSpec struct {
	PackageName string `json:"package_name"`
	// VersionCode is the exact version wanted. Agents never go below the
	// version already on the TV: on Android a downgrade means uninstalling,
	// which deletes the launcher's data, pairing included. To undo a bad
	// release, republish the good build with a higher VersionCode.
	VersionCode int64  `json:"version_code"`
	APKURL      string `json:"apk_url"`
	// SHA256 is the hex digest of the APK. Agents must refuse to install an
	// artifact whose digest does not match.
	SHA256         string   `json:"sha256"`
	Activation     string   `json:"activation"`
	HomeActivity   string   `json:"home_activity,omitempty"` // e.g. "tech.spectrumlabs.hospitium.launcher/.MainActivity"
	StockLaunchers []string `json:"stock_launchers,omitempty"`
}

// AppSpec declares an app that must be present. Store apps (Netflix etc.)
// cannot be sideloaded; when APKURL is empty the agent only verifies
// presence and reports non-compliance.
type AppSpec struct {
	PackageName    string `json:"package_name"`
	MinVersionCode int64  `json:"min_version_code,omitempty"`
	APKURL         string `json:"apk_url,omitempty"`
	SHA256         string `json:"sha256,omitempty"`
}

// Setting is an Android `settings` value. Only allowlisted keys are accepted
// by the control plane.
type Setting struct {
	Namespace string `json:"namespace"` // system | secure | global
	Key       string `json:"key"`
	Value     string `json:"value"`
}

// Operation types. The set is closed: the protocol deliberately has no
// "run arbitrary shell command" operation.
const (
	OpClearAppData      = "clear_app_data"     // params: ClearAppDataParams
	OpReboot            = "reboot"             // params: none
	OpDiagnostics       = "diagnostics"        // params: none; fixed, read-only command set
	OpReinstallLauncher = "reinstall_launcher" // params: none; reinstall the spec's build in place, keeping app data (never uninstalls)
)

type Operation struct {
	ID       string          `json:"id"` // UUID; idempotency key on the agent
	DeviceID int64           `json:"device_id"`
	Type     string          `json:"type"`
	Params   json.RawMessage `json:"params,omitempty"`
	Attempt  int             `json:"attempt"`
	// ExpiresAt bounds execution. Agents must not start an operation after
	// it expires (protects a new guest from a stale checkout cleanup).
	ExpiresAt time.Time `json:"expires_at"`
}

type ClearAppDataParams struct {
	Packages []string `json:"packages"`
	// ReservationRef identifies the stay that triggered the cleanup, for
	// audit only.
	ReservationRef string `json:"reservation_ref,omitempty"`
}

// StatusReport is posted by the agent after each reconcile pass.
type StatusReport struct {
	AgentVersion  string         `json:"agent_version"`
	Hostname      string         `json:"hostname"`
	UptimeSeconds int64          `json:"uptime_seconds"`
	MemAllocBytes uint64         `json:"mem_alloc_bytes"`
	Devices       []DeviceStatus `json:"devices"`
}

type DeviceStatus struct {
	DeviceID           int64            `json:"device_id"`
	ObservedGeneration int64            `json:"observed_generation"`
	Reachable          bool             `json:"reachable"`
	Authorized         bool             `json:"authorized"` // ADB key accepted on the TV
	Compliant          bool             `json:"compliant"`
	Model              string           `json:"model,omitempty"`
	Manufacturer       string           `json:"manufacturer,omitempty"`
	OSVersion          string           `json:"os_version,omitempty"`
	SDKLevel           int              `json:"sdk_level,omitempty"`
	Packages           map[string]int64 `json:"packages,omitempty"` // package -> versionCode, managed packages only
	Conditions         []Condition      `json:"conditions,omitempty"`
	LastError          string           `json:"last_error,omitempty"`
	ReconciledAt       time.Time        `json:"reconciled_at"`
	NextAttemptAt      *time.Time       `json:"next_attempt_at,omitempty"`
}

// Condition follows the Kubernetes convention: Type is a noun, Status is
// "True" | "False" | "Unknown".
type Condition struct {
	Type    string `json:"type"` // e.g. Reachable, LauncherInstalled, LauncherActive, AppsPresent, SettingsApplied
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

const (
	OpStatusSucceeded = "succeeded"
	OpStatusFailed    = "failed"
	OpStatusRejected  = "rejected" // agent refused (expired, unknown type, bad params); never retried
)

type OperationResult struct {
	Status       string            `json:"status"`
	Attempt      int               `json:"attempt"`
	Message      string            `json:"message,omitempty"`
	Retryable    bool              `json:"retryable"`
	Verification map[string]string `json:"verification,omitempty"`
	StartedAt    time.Time         `json:"started_at"`
	FinishedAt   time.Time         `json:"finished_at"`
}
