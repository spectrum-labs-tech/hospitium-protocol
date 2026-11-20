// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Where a Background's image comes from, which decides how a launcher fetches it.
const (
	// BackgroundOriginControlPlane: URLs are PathLauncherMedia paths on the control plane,
	// fetched with the device token and cacheable forever by address.
	BackgroundOriginControlPlane = "control_plane"
	// BackgroundOriginUnsplash: URLs are absolute https URLs on UnsplashImageHost. Unsplash's API
	// terms require loading images from the URLs it returns (hotlinking) and crediting the
	// photographer wherever the image is shown. A launcher fetches them with no Authorization
	// header and no cookies, and shows Attribution whenever the image is on screen.
	BackgroundOriginUnsplash = "unsplash"
)

// UnsplashImageHost is the only host a launcher loads an Unsplash background from.
const UnsplashImageHost = "images.unsplash.com"

// How a launcher keeps text legible over a background.
const (
	// BackgroundOverlayGradient darkens most behind the text (the left of the screen and a band
	// along the bottom), fading toward the rest of the image.
	BackgroundOverlayGradient = "gradient"
	// BackgroundOverlayDim darkens the whole image evenly.
	BackgroundOverlayDim = "dim"
)

// Bounds on Background.Scrim, the overlay's opacity.
const (
	BackgroundScrimMin = 0.30
	BackgroundScrimMax = 0.80
)

// Background (added in v0.8.0) is the image behind a launcher's home screen.
type Background struct {
	URL     string `json:"url"`                // 1920x1080
	URL2160 string `json:"url_2160,omitempty"` // 3840x2160, when the source is large enough
	Origin  string `json:"origin"`             // a BackgroundOrigin* constant
	Overlay string `json:"overlay"`            // a BackgroundOverlay* constant
	// Scrim is the overlay's opacity, BackgroundScrimMin to BackgroundScrimMax: stronger for
	// brighter, busier images.
	Scrim float64 `json:"scrim"`
	// Placeholder is a BlurHash (https://blurha.sh) a launcher may draw while the image loads.
	Placeholder string `json:"placeholder,omitempty"`
	// Attribution, when present, must be shown whenever this image is, and never with another.
	Attribution *BackgroundAttribution `json:"attribution,omitempty"`
}

// BackgroundAttribution is the credit an image's license or source requires on screen.
type BackgroundAttribution struct {
	Text string `json:"text"` // e.g. "Photo by Ana Ruiz on Unsplash"
}

var mediaPath = regexp.MustCompile(`^` + regexp.QuoteMeta(strings.TrimSuffix(PathLauncherMedia, "{sha256}")) + `[0-9a-f]{64}$`)

// ValidMediaPath reports whether p is a PathLauncherMedia path with a well-formed address.
func ValidMediaPath(p string) bool { return mediaPath.MatchString(p) }

// Valid reports why a launcher must not use b, or nil. Control planes check what they send with
// it; launchers apply the same rules before fetching anything.
func (b Background) Valid() error {
	switch b.Overlay {
	case BackgroundOverlayGradient, BackgroundOverlayDim:
	default:
		return fmt.Errorf("background overlay %q is not one of %q, %q", b.Overlay, BackgroundOverlayGradient, BackgroundOverlayDim)
	}
	if b.Scrim < BackgroundScrimMin || b.Scrim > BackgroundScrimMax {
		return fmt.Errorf("background scrim %.2f is outside %.2f-%.2f", b.Scrim, BackgroundScrimMin, BackgroundScrimMax)
	}
	if len(b.Placeholder) > 100 {
		return errors.New("background placeholder is longer than 100 characters")
	}
	if b.Attribution != nil && strings.TrimSpace(b.Attribution.Text) == "" {
		return errors.New("background attribution has no text")
	}
	urls := []string{b.URL}
	if b.URL2160 != "" {
		urls = append(urls, b.URL2160)
	}
	switch b.Origin {
	case BackgroundOriginControlPlane:
		for _, u := range urls {
			if !ValidMediaPath(u) {
				return fmt.Errorf("control-plane background %q is not a media path", u)
			}
		}
	case BackgroundOriginUnsplash:
		if b.Attribution == nil {
			return errors.New("an Unsplash background needs its attribution")
		}
		for _, u := range urls {
			if err := validUnsplashURL(u); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("background origin %q is unknown", b.Origin)
	}
	return nil
}

func validUnsplashURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host != UnsplashImageHost || u.User != nil || u.Fragment != "" || u.Path == "" || u.Path == "/" {
		return fmt.Errorf("background %q is not an https URL on %s", raw, UnsplashImageHost)
	}
	return nil
}
