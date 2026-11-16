// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"strings"
	"testing"
)

func TestValidUsageCounter(t *testing.T) {
	t.Parallel()
	for _, ok := range []string{"home", "wifi", "house_rules", "amenities", "announcement:12", "app:com.netflix.ninja", "app:com.google.android.youtube.tv"} {
		if !ValidUsageCounter(ok) {
			t.Errorf("%q should be valid", ok)
		}
	}
	for _, bad := range []string{"", "Home", "search", "announcement:", "announcement:0", "announcement:012", "announcement:x",
		"app:", "app:netflix", "app:com..netflix", "app:com.netflix ninja", "app:" + strings.Repeat("a.", 200) + "b", "guest:alice"} {
		if ValidUsageCounter(bad) {
			t.Errorf("%q should be rejected", bad)
		}
	}
}
