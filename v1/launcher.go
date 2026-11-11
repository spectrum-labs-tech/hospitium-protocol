// SPDX-License-Identifier: Apache-2.0

package v1

import "time"

// Launcher endpoints. The launcher talks to the control plane directly and
// never depends on the management agent being present.
const (
	PathLauncherPairingStart  = "/v1/launcher/pairings"
	PathLauncherPairingStatus = "/v1/launcher/pairings/{code}"
	PathLauncherConfig        = "/v1/launcher/config"
)

// PairingStartResponse: the launcher displays Code on screen; an administrator
// enters it in Hospitium to bind this TV to a device record. The launcher polls
// PathLauncherPairingStatus with PollToken until it receives a DeviceToken.
type PairingStartResponse struct {
	Code      string    `json:"code"`
	PollToken string    `json:"poll_token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PairingStatusResponse struct {
	Status      string `json:"status"`                 // pending | claimed | expired
	DeviceToken string `json:"device_token,omitempty"` // returned exactly once
	DeviceID    int64  `json:"device_id,omitempty"`
}

// LauncherConfig is everything the TV needs to render. It is safe to cache.
// Guest fields carry their own expiry which the launcher must enforce locally,
// including while offline.
type LauncherConfig struct {
	ConfigVersion string         `json:"config_version"` // opaque; changes when content changes
	Property      PropertyInfo   `json:"property"`
	Unit          *UnitInfo      `json:"unit,omitempty"`
	Guest         *GuestInfo     `json:"guest,omitempty"`
	Announcements []Announcement `json:"announcements,omitempty"`
	Apps          []LauncherApp  `json:"apps,omitempty"`
	ServerTime    time.Time      `json:"server_time"`
	// RefreshAfterSeconds is the suggested polling interval.
	RefreshAfterSeconds int `json:"refresh_after_seconds"`
}

type PropertyInfo struct {
	Name       string            `json:"name"`
	Timezone   string            `json:"timezone"`
	Branding   map[string]string `json:"branding,omitempty"` // logo_url, background_url, primary_color, ...
	Wifi       *WifiInfo         `json:"wifi,omitempty"`
	Amenities  []string          `json:"amenities,omitempty"`
	HouseRules []string          `json:"house_rules,omitempty"`
}

type WifiInfo struct {
	SSID     string `json:"ssid"`
	Password string `json:"password,omitempty"`
}

type UnitInfo struct {
	Name string `json:"name"`
}

// GuestInfo is deliberately minimal: a display name, never the full
// reservation. ExpiresAt is the latest moment the launcher may display it.
type GuestInfo struct {
	DisplayName string    `json:"display_name"`
	CheckIn     time.Time `json:"check_in"`
	CheckOut    time.Time `json:"check_out"`
	ExpiresAt   time.Time `json:"expires_at"`
	// CheckoutReminderAt is when the launcher should start showing a
	// checkout reminder.
	CheckoutReminderAt time.Time `json:"checkout_reminder_at"`
}

type Announcement struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// LauncherApp is a streaming app to feature on the home screen if installed.
type LauncherApp struct {
	PackageName string `json:"package_name"`
	Label       string `json:"label"`
}
