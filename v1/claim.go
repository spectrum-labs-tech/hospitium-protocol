// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"crypto/sha256"
	"encoding/base32"
	"strings"
	"time"
)

// Hub claiming (added in v0.4.0): a hub that hasn't enrolled proves who it is with a claim secret
// it generated itself, and an owner claims it by the claim code printed on its label. The code is
// derived from the secret by a one-way hash, so knowing the code (from the label) doesn't let anyone
// act as the hub.

// PathAgentClaim is polled by an unenrolled hub until an owner has claimed it.
const PathAgentClaim = "/v1/agent/claim"

// ClaimSecretBytes is the length of the random secret a hub generates once.
const ClaimSecretBytes = 32

// ClaimRequest is posted by an unenrolled hub. The control plane derives the claim code from
// ClaimSecret and looks for an owner's claim of it.
type ClaimRequest struct {
	ClaimSecret  string `json:"claim_secret"` // base64url, no padding, of ClaimSecretBytes random bytes
	Hostname     string `json:"hostname"`
	AgentVersion string `json:"agent_version"`
	Hardware     string `json:"hardware,omitempty"`
}

// ClaimPending is the answer (HTTP 202) while no owner has claimed the hub yet. The hub polls again
// after RetryAfterSeconds. Once claimed, the answer is an EnrollResponse (HTTP 200), exactly as for
// enrollment with a token.
type ClaimPending struct {
	Status            string    `json:"status"` // "waiting"
	RetryAfterSeconds int       `json:"retry_after_seconds"`
	ServerTime        time.Time `json:"server_time"`
}

// claimAlphabet is Crockford's base32: no I, L, O or U, so a code read off a label can't be
// mistyped between look-alike characters.
var claimAlphabet = base32.NewEncoding("0123456789ABCDEFGHJKMNPQRSTVWXYZ").WithPadding(base32.NoPadding)

// claimCodeChars is the code's length: 10 characters, 50 bits of the secret's hash.
const claimCodeChars = 10

// ClaimCode derives a hub's claim code from its secret, formatted for a label as XXXXX-XXXXX.
func ClaimCode(secret []byte) string {
	h := sha256.Sum256(append([]byte("hospitium-claim-v1:"), secret...))
	c := claimAlphabet.EncodeToString(h[:])[:claimCodeChars]
	return c[:5] + "-" + c[5:]
}

// NormalizeClaimCode turns what a person typed into the canonical code, or "" if it can't be one:
// case and separators are ignored, and the look-alikes Crockford allows (O for 0, I and L for 1)
// are mapped.
func NormalizeClaimCode(s string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(s) {
		switch r {
		case ' ', '-', '_':
			continue
		case 'O':
			r = '0'
		case 'I', 'L':
			r = '1'
		}
		if !strings.ContainsRune("0123456789ABCDEFGHJKMNPQRSTVWXYZ", r) {
			return ""
		}
		b.WriteRune(r)
	}
	c := b.String()
	if len(c) != claimCodeChars {
		return ""
	}
	return c[:5] + "-" + c[5:]
}
