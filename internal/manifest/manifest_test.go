package manifest_test

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/modeltorrent-foundation/mt/internal/manifest"
)

// goodManifest builds a redistributable, signed manifest for round-trip tests.
func goodManifest(t *testing.T) (manifest.Manifest, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	m := manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		ModelID:       "Qwen/Qwen3-8B",
		Files: []manifest.File{
			{Path: "model.safetensors", Size: 65536, SHA256: "aa"},
		},
		License:   manifest.License{SPDX: "Apache-2.0", Redistributable: true},
		Publisher: manifest.Publisher{KeyID: "ed25519:test"},
		Magnet:    "magnet:?xt=urn:btmh:0001",
		CreatedAt: "2026-09-15T00:00:00Z",
	}
	sig := ed25519.Sign(priv, m.Canonical())
	m.Publisher.Signature = base64.StdEncoding.EncodeToString(sig)
	return m, pub
}

func TestLicenseAllowed(t *testing.T) {
	if !manifest.LicenseAllowed("Apache-2.0") {
		t.Errorf("Apache-2.0 should be redistributable")
	}
	if !manifest.LicenseAllowed("MIT") {
		t.Errorf("MIT should be redistributable")
	}
	if manifest.LicenseAllowed("LicenseRef-Proprietary") {
		t.Errorf("proprietary license must not be redistributable")
	}
	if manifest.LicenseAllowed("Llama-Community") {
		t.Errorf("Llama Community is not on the v1 allowlist")
	}
}

func TestParseRejectsNonRedistributable(t *testing.T) {
	bad := manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		ModelID:       "meta-llama/leaked",
		License:       manifest.License{SPDX: "LicenseRef-Proprietary", Redistributable: false},
	}
	b, _ := json.Marshal(bad)
	if _, err := manifest.Parse(b); err == nil {
		t.Fatalf("Parse must reject a non-redistributable manifest")
	}
}

func TestParseRoundTripAndVerify(t *testing.T) {
	m, pub := goodManifest(t)
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got, err := manifest.Parse(b)
	if err != nil {
		t.Fatalf("Parse good manifest: %v", err)
	}
	if got.ModelID != m.ModelID {
		t.Errorf("modelId = %q want %q", got.ModelID, m.ModelID)
	}
	if err := got.Verify(pub); err != nil {
		t.Errorf("Verify good signature: %v", err)
	}
}

func TestCanonicalExcludesSignature(t *testing.T) {
	m, _ := goodManifest(t)
	withSig := m.Canonical()
	m.Publisher.Signature = "tampered-different-value"
	withOther := m.Canonical()
	if string(withSig) != string(withOther) {
		t.Errorf("Canonical() must ignore publisher.signature so verification is stable")
	}
	if len(withSig) == 0 {
		t.Errorf("Canonical() must produce deterministic non-empty bytes")
	}
}
