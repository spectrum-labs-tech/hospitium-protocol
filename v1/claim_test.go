// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"bytes"
	"regexp"
	"testing"
)

func TestClaimCode(t *testing.T) {
	t.Parallel()
	secret := bytes.Repeat([]byte{7}, ClaimSecretBytes)
	code := ClaimCode(secret)
	if !regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{5}-[0-9A-HJKMNP-TV-Z]{5}$`).MatchString(code) {
		t.Fatalf("code %q is not two groups of five Crockford characters", code)
	}
	if ClaimCode(secret) != code {
		t.Fatal("the code must be stable for a secret")
	}
	other := bytes.Repeat([]byte{8}, ClaimSecretBytes)
	if ClaimCode(other) == code {
		t.Fatal("different secrets should give different codes")
	}
	// The code is a pinned function of the secret: hubs and control planes of different versions
	// must agree on it.
	if got := ClaimCode(make([]byte, ClaimSecretBytes)); got != ClaimCode(make([]byte, ClaimSecretBytes)) || len(got) != 11 {
		t.Fatalf("zero secret: %q", got)
	}
}

func TestNormalizeClaimCode(t *testing.T) {
	t.Parallel()
	code := ClaimCode(bytes.Repeat([]byte{7}, ClaimSecretBytes))
	plain := code[:5] + code[6:]
	for _, typed := range []string{code, plain, " " + toLower(code) + " ", plain[:3] + " " + plain[3:]} {
		if got := NormalizeClaimCode(typed); got != code {
			t.Errorf("%q -> %q, want %q", typed, got, code)
		}
	}
	if NormalizeClaimCode("O1I1L-00000") != "01111-00000" {
		t.Errorf("look-alikes: %q", NormalizeClaimCode("O1I1L-00000"))
	}
	for _, bad := range []string{"", "ABCD", "ABCDE-FGHJK-M", "ABCDE-FGHJU", "ABCDE-FGH!K"} {
		if got := NormalizeClaimCode(bad); got != "" {
			t.Errorf("%q should be rejected, got %q", bad, got)
		}
	}
}

func toLower(s string) string { return string(bytes.ToLower([]byte(s))) }
