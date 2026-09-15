// Package manifest defines the signed, portable description of a model as fixed
// by PROTOCOL.md v1 (schemaVersion 1).
//
// Do NOT change the wire shape here without changing PROTOCOL.md.
package manifest

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// SchemaVersion is the only manifest schema currently defined.
const SchemaVersion = 1

// RedistributableLicenses is the Wave 0 SPDX allowlist from PROTOCOL.md §3.
// Extend deliberately and with counsel; never widen silently.
var RedistributableLicenses = []string{
	"Apache-2.0",
	"MIT",
	"BSD-2-Clause",
	"BSD-3-Clause",
	"CC-BY-4.0",
	"CC-BY-SA-4.0",
	"CC0-1.0",
	"OpenRAIL-permissive",
}

// File is one content-addressed file in a model. Identity is (sha256, size).
type File struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// License couples an SPDX id with the redistribution decision. Both must agree
// with the PROTOCOL.md allowlist at ingest.
type License struct {
	SPDX            string `json:"spdx"`
	Redistributable bool   `json:"redistributable"`
}

// Publisher identifies the signing key and carries the Ed25519 signature over
// the manifest's canonical bytes (see Canonical).
type Publisher struct {
	KeyID     string `json:"keyId"`
	Signature string `json:"signature"` // base64(ed25519 signature)
}

// Manifest is the top-level, signable model descriptor.
type Manifest struct {
	SchemaVersion int       `json:"schemaVersion"`
	ModelID       string    `json:"modelId"`
	Files         []File    `json:"files"`
	License       License   `json:"license"`
	Publisher     Publisher `json:"publisher"`
	Magnet        string    `json:"magnet"`
	Webseeds      []string  `json:"webseeds"`
	CreatedAt     string    `json:"createdAt"`
}

// Parse decodes and validates a manifest. It MUST reject manifests that fail the
// license gate (PROTOCOL.md §3) and manifests with an unknown schemaVersion.
func Parse(b []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return Manifest{}, fmt.Errorf("manifest: decode: %w", err)
	}
	if m.SchemaVersion != SchemaVersion {
		return Manifest{}, fmt.Errorf("manifest: unsupported schemaVersion %d", m.SchemaVersion)
	}
	if !LicenseAllowed(m.License.SPDX) || !m.License.Redistributable {
		return Manifest{}, errors.New("manifest: license not redistributable")
	}
	return m, nil
}

// Canonical returns the deterministic bytes used for signing and verification:
// the manifest with an empty publisher.signature, stable field order, no
// insignificant whitespace (PROTOCOL.md §4).
func (m Manifest) Canonical() []byte {
	c := m
	c.Publisher.Signature = ""
	b, err := json.Marshal(&c)
	if err != nil {
		return nil
	}
	return b
}

// Verify checks the publisher's Ed25519 signature over Canonical() against pub.
func (m Manifest) Verify(pub ed25519.PublicKey) error {
	sig, err := base64.StdEncoding.DecodeString(m.Publisher.Signature)
	if err != nil {
		return fmt.Errorf("manifest: decode signature: %w", err)
	}
	if !ed25519.Verify(pub, m.Canonical(), sig) {
		return errors.New("manifest: invalid signature")
	}
	return nil
}

// LicenseAllowed reports whether an SPDX id is redistributable under the
// Wave 0 allowlist.
func LicenseAllowed(spdx string) bool {
	for _, allowed := range RedistributableLicenses {
		if spdx == allowed {
			return true
		}
	}
	return false
}
