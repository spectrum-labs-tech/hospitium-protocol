// SPDX-License-Identifier: Apache-2.0

package v1

import "time"

// Launcher endpoints. The launcher talks to the control plane directly and
// never depends on the management agent being present.
const (
	PathLauncherPairingStart  = "/v1/launcher/pairings"
	PathLauncherPairingStatus = "/v1/launcher/pairings/{code}"
	PathLauncherConfig        = "/v1/launcher/config"
	// PathLauncherMedia serves an image by its content address (lowercase hex SHA-256) to device
	// tokens only. Responses are immutable: a launcher may cache by address forever, and should
	// verify the bytes hash to the address.
	PathLauncherMedia = "/v1/launcher/media/{sha256}"
)

// Branding keys in PropertyInfo.Branding. Values are strings; launchers ignore keys they don't
// know. Background URLs are paths relative to the control plane (PathLauncherMedia), fetched with
// the device token; a launcher must not send its token to any other host.
const (
	BrandingPrimaryColor = "primary_color" // "#RRGGBB"
	BrandingLogoURL      = "logo_url"
	// BrandingBackgroundURL is a 1920x1080 JPEG for the home screen.
	BrandingBackgroundURL = "background_url"
	// BrandingBackgroundURL2160 is the same image at 3840x2160, when one exists.
	BrandingBackgroundURL2160 = "background_url_2160"
	// BrandingBackgroundScrim is the opacity ("0.30" to "0.80") of the gradient a launcher draws
	// behind text so it stays legible over the image.
	BrandingBackgroundScrim = "background_scrim"
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
	Branding   map[string]string `json:"branding,omitempty"` // keys: the Branding* constants
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
