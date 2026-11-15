// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// A v0.2 status (no facts, no capabilities) must still decode, and a status
// without them must encode exactly as before, so old and new agents and
// control planes interoperate.
func TestCapabilityFieldsAreAdditive(t *testing.T) {
	t.Parallel()
	old := `{"device_id":7,"observed_generation":2,"reachable":true,"authorized":true,"compliant":true,"reconciled_at":"2026-11-15T10:00:00Z"}`
	var st DeviceStatus
	if err := json.Unmarshal([]byte(old), &st); err != nil {
		t.Fatal(err)
	}
	if st.Facts != nil || st.Capabilities != nil {
		t.Fatalf("old status should decode with no facts or capabilities: %+v", st)
	}
	out, err := json.Marshal(st)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"facts", "capabilities"} {
		if strings.Contains(string(out), `"`+k+`"`) {
			t.Fatalf("empty %s should be omitted: %s", k, out)
		}
	}
	res, _ := json.Marshal(OperationResult{Status: OpStatusSucceeded})
	if strings.Contains(string(res), "capabilities") {
		t.Fatalf("empty capabilities should be omitted: %s", res)
	}
}

func TestCapabilityObservationRoundTrip(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	in := DeviceStatus{
		DeviceID: 7,
		Facts:    &DeviceFacts{Driver: DriverAndroidADB, Model: "Chromecast", BuildFingerprint: "google/sabrina/sabrina:12/STTE.1/1:user/release-keys"},
		Capabilities: []CapabilityObservation{
			{Capability: CapAppDataCleanup, Subject: "com.example.app", State: CapSupported, Evidence: EvidenceVerified, ObservedAt: at},
			{Capability: CapCustomLauncher, State: CapUnsupported, Evidence: EvidenceVerified, Detail: "HOME resolved to stock", ObservedAt: at},
		},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out DeviceStatus
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.Facts == nil || out.Facts.BuildFingerprint != in.Facts.BuildFingerprint || len(out.Capabilities) != 2 ||
		out.Capabilities[0] != in.Capabilities[0] || out.Capabilities[1] != in.Capabilities[1] {
		t.Fatalf("round trip lost data:\n in %+v\nout %+v", in, out)
	}
	if !strings.Contains(string(b), `"subject":"com.example.app"`) || strings.Count(string(b), `"subject"`) != 1 {
		t.Fatalf("subject should appear only where set: %s", b)
	}
}
