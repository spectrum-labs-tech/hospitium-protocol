// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"encoding/json"
	"strings"
	"testing"
)

const testAddr = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestBackgroundValid(t *testing.T) {
	t.Parallel()
	media := "/v1/launcher/media/" + testAddr
	unsplash := "https://images.unsplash.com/photo-1506905925346-21bda4d32df4?ixid=abc&w=1920&h=1080&fit=crop"
	credit := &BackgroundAttribution{Text: "Photo by Ana Ruiz on Unsplash"}
	ok := func(b Background) Background { // fills the fields a case doesn't care about
		if b.Overlay == "" {
			b.Overlay = BackgroundOverlayGradient
		}
		if b.Scrim == 0 {
			b.Scrim = 0.55
		}
		return b
	}
	cases := []struct {
		name    string
		bg      Background
		wantErr string
	}{
		{"control plane, both sizes", ok(Background{Origin: BackgroundOriginControlPlane, URL: media, URL2160: media}), ""},
		{"control plane, dim overlay at the bounds", Background{Origin: BackgroundOriginControlPlane, URL: media, Overlay: BackgroundOverlayDim, Scrim: 0.30}, ""},
		{"unsplash with credit", ok(Background{Origin: BackgroundOriginUnsplash, URL: unsplash, URL2160: unsplash, Attribution: credit}), ""},
		{"control plane, absolute URL", ok(Background{Origin: BackgroundOriginControlPlane, URL: "https://example.com" + media}), "not a media path"},
		{"control plane, short address", ok(Background{Origin: BackgroundOriginControlPlane, URL: "/v1/launcher/media/abc"}), "not a media path"},
		{"control plane, uppercase address", ok(Background{Origin: BackgroundOriginControlPlane, URL: "/v1/launcher/media/" + strings.ToUpper(testAddr)}), "not a media path"},
		{"control plane, bad 4K URL", ok(Background{Origin: BackgroundOriginControlPlane, URL: media, URL2160: "/etc/passwd"}), "not a media path"},
		{"unsplash without credit", ok(Background{Origin: BackgroundOriginUnsplash, URL: unsplash}), "needs its attribution"},
		{"unsplash, blank credit", ok(Background{Origin: BackgroundOriginUnsplash, URL: unsplash, Attribution: &BackgroundAttribution{Text: " "}}), "no text"},
		{"unsplash over http", ok(Background{Origin: BackgroundOriginUnsplash, URL: strings.Replace(unsplash, "https", "http", 1), Attribution: credit}), "not an https URL"},
		{"unsplash, other host", ok(Background{Origin: BackgroundOriginUnsplash, URL: "https://images.unsplash.com.evil.example/x", Attribution: credit}), "not an https URL"},
		{"unsplash, userinfo", ok(Background{Origin: BackgroundOriginUnsplash, URL: "https://user@images.unsplash.com/x", Attribution: credit}), "not an https URL"},
		{"unsplash, port", ok(Background{Origin: BackgroundOriginUnsplash, URL: "https://images.unsplash.com:8443/x", Attribution: credit}), "not an https URL"},
		{"unsplash, no path", ok(Background{Origin: BackgroundOriginUnsplash, URL: "https://images.unsplash.com/", Attribution: credit}), "not an https URL"},
		{"unsplash, media path", ok(Background{Origin: BackgroundOriginUnsplash, URL: media, Attribution: credit}), "not an https URL"},
		{"unknown origin", ok(Background{Origin: "pexels", URL: media}), "unknown"},
		{"unknown overlay", Background{Origin: BackgroundOriginControlPlane, URL: media, Overlay: "blur", Scrim: 0.5}, "overlay"},
		{"scrim too low", Background{Origin: BackgroundOriginControlPlane, URL: media, Overlay: BackgroundOverlayDim, Scrim: 0.29}, "scrim"},
		{"scrim too high", Background{Origin: BackgroundOriginControlPlane, URL: media, Overlay: BackgroundOverlayDim, Scrim: 0.81}, "scrim"},
		{"placeholder too long", ok(Background{Origin: BackgroundOriginControlPlane, URL: media, Placeholder: strings.Repeat("L", 101)}), "placeholder"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := c.bg.Valid()
			switch {
			case c.wantErr == "" && err != nil:
				t.Fatalf("want valid, got %v", err)
			case c.wantErr != "" && (err == nil || !strings.Contains(err.Error(), c.wantErr)):
				t.Fatalf("want an error containing %q, got %v", c.wantErr, err)
			}
		})
	}
}

func TestValidMediaPath(t *testing.T) {
	t.Parallel()
	if !ValidMediaPath("/v1/launcher/media/" + testAddr) {
		t.Error("a well-formed media path should be valid")
	}
	for _, bad := range []string{"", "/v1/launcher/media/", "/v1/launcher/media/" + testAddr + "/x", "/v1/agent/artifacts/" + testAddr} {
		if ValidMediaPath(bad) {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

// Older launchers decode a config with the new fields and simply ignore them; a config without
// them decodes to nil and empty, so launchers fall back as before.
func TestLauncherConfigBackgroundJSON(t *testing.T) {
	t.Parallel()
	in := LauncherConfig{
		Background: &Background{Origin: BackgroundOriginUnsplash, URL: "https://images.unsplash.com/p?ixid=1", Overlay: BackgroundOverlayDim,
			Scrim: 0.6, Placeholder: "LEHV6nWB2yk8pyo0adR*.7kCMdnj", Attribution: &BackgroundAttribution{Text: "Photo by A on Unsplash"}},
		FallbackBackgrounds: []string{"/v1/launcher/media/" + testAddr},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"background":{`, `"origin":"unsplash"`, `"attribution":{"text":`, `"fallback_backgrounds":[`} {
		if !strings.Contains(string(b), key) {
			t.Errorf("JSON lacks %s: %s", key, b)
		}
	}
	var out LauncherConfig
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.Background == nil || *out.Background.Attribution != *in.Background.Attribution || out.FallbackBackgrounds[0] != in.FallbackBackgrounds[0] {
		t.Fatalf("round trip lost fields: %+v", out)
	}

	var old LauncherConfig
	if err := json.Unmarshal([]byte(`{"config_version":"x","property":{"name":"P","timezone":"UTC"}}`), &old); err != nil {
		t.Fatal(err)
	}
	if old.Background != nil || old.FallbackBackgrounds != nil {
		t.Fatalf("a config without backgrounds should decode to none: %+v", old)
	}
	if strings.Contains(string(mustJSON(t, old)), "background") {
		t.Fatal("empty background fields should be omitted")
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
