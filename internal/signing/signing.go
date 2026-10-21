// Package signing lets skill publishers sign a skill folder and lets consumers
// verify it against a list of publishers they trust.
//
// A signature covers the skill's content hash (every file except SKILL.sig
// itself), so any edit to a signed skill invalidates it. Signatures are
// Ed25519; keys are exchanged as base64 text, so a public key can be pasted in
// a README or committed to a team repository.
package signing

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aminmesbahi/skell/internal/hasher"
)

// FileName is the signature file kept inside a signed skill folder.
const FileName = "SKILL.sig"

const (
	algorithm = "ed25519"
	context   = "skell-skill-v1\n"
)

// Signature is the content of SKILL.sig.
type Signature struct {
	Version     int    `json:"version"`
	Algorithm   string `json:"algorithm"`
	Publisher   string `json:"publisher"`
	KeyID       string `json:"key_id"`
	PublicKey   string `json:"public_key"`
	ContentHash string `json:"content_hash"`
	SignedAt    string `json:"signed_at"`
	Signature   string `json:"signature"`
}

// Status describes the outcome of verifying one skill.
type Status string

// Verification outcomes.
const (
	// StatusTrusted: valid signature from a key in the trust store.
	StatusTrusted Status = "trusted"
	// StatusUntrusted: valid signature, but the key is not trusted here.
	StatusUntrusted Status = "untrusted-key"
	// StatusTampered: the files changed after signing.
	StatusTampered Status = "tampered"
	// StatusInvalid: the signature file is damaged or does not verify.
	StatusInvalid Status = "invalid"
	// StatusUnsigned: the skill carries no signature.
	StatusUnsigned Status = "unsigned"
)

// Result is what Verify reports.
type Result struct {
	Status    Status `json:"status"`
	Publisher string `json:"publisher,omitempty"`
	KeyID     string `json:"key_id,omitempty"`
	SignedAt  string `json:"signed_at,omitempty"`
	Detail    string `json:"detail,omitempty"`
}

// OK reports whether the skill passes a "must be signed by someone I trust" check.
func (r Result) OK() bool { return r.Status == StatusTrusted }

// KeyPair is a freshly generated signing identity.
type KeyPair struct {
	Publisher string
	Public    ed25519.PublicKey
	Private   ed25519.PrivateKey
}

// KeyID is a short stable fingerprint of a public key.
func KeyID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:8])
}

// GenerateKey creates a new Ed25519 key pair for a publisher.
func GenerateKey(publisher string) (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &KeyPair{Publisher: publisher, Public: pub, Private: priv}, nil
}

// EncodePublic returns the base64 text form of a public key.
func EncodePublic(pub ed25519.PublicKey) string { return base64.StdEncoding.EncodeToString(pub) }

// DecodePublic parses a public key given as base64 text, or as the contents
// of a .pub file ("skell-public-key <base64> [publisher]").
func DecodePublic(text string) (ed25519.PublicKey, error) {
	fields := strings.Fields(text)
	if len(fields) > 0 && fields[0] == "skell-public-key" && len(fields) > 1 {
		text = fields[1]
	} else if len(fields) > 0 {
		text = fields[0]
	}
	raw, err := base64.StdEncoding.DecodeString(text)
	if err != nil || len(raw) != ed25519.PublicKeySize {
		return nil, errors.New("not a valid Skell public key")
	}
	return ed25519.PublicKey(raw), nil
}

// WriteKeyFiles saves <dir>/<publisher>.key (private, 0600) and <dir>/<publisher>.pub.
func WriteKeyFiles(dir string, kp *KeyPair) (keyPath, pubPath string, err error) {
	if err = os.MkdirAll(dir, 0o700); err != nil {
		return "", "", err
	}
	keyPath = filepath.Join(dir, kp.Publisher+".key")
	pubPath = filepath.Join(dir, kp.Publisher+".pub")
	for _, p := range []string{keyPath, pubPath} {
		if _, statErr := os.Stat(p); statErr == nil {
			return "", "", fmt.Errorf("%s already exists — refusing to overwrite a key", p)
		}
	}
	priv := "skell-private-key " + base64.StdEncoding.EncodeToString(kp.Private.Seed()) + "\n"
	if err = os.WriteFile(keyPath, []byte(priv), 0o600); err != nil {
		return "", "", err
	}
	pub := fmt.Sprintf("skell-public-key %s %s\n", EncodePublic(kp.Public), kp.Publisher)
	if err = os.WriteFile(pubPath, []byte(pub), 0o644); err != nil { //nolint:gosec // a public key is meant to be shared
		return "", "", err
	}
	return keyPath, pubPath, nil
}

// ReadPrivateKey loads a key written by WriteKeyFiles.
func ReadPrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return nil, err
	}
	fields := strings.Fields(string(data))
	if len(fields) != 2 || fields[0] != "skell-private-key" {
		return nil, fmt.Errorf("%s is not a Skell private key", path)
	}
	seed, err := base64.StdEncoding.DecodeString(fields[1])
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("%s is not a valid Skell private key", path)
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

func message(contentHash string) []byte { return []byte(context + contentHash) }

// Sign signs the skill in dir and writes SKILL.sig into it.
func Sign(dir, publisher string, key ed25519.PrivateKey) (*Signature, error) {
	if publisher == "" {
		return nil, errors.New("a publisher name is required")
	}
	if _, err := os.Stat(filepath.Join(dir, "SKILL.md")); err != nil {
		return nil, fmt.Errorf("%s has no SKILL.md", dir)
	}
	hash, err := hasher.HashDirExcluding(dir, FileName)
	if err != nil {
		return nil, err
	}
	pub := key.Public().(ed25519.PublicKey)
	sig := &Signature{
		Version:     1,
		Algorithm:   algorithm,
		Publisher:   publisher,
		KeyID:       KeyID(pub),
		PublicKey:   EncodePublic(pub),
		ContentHash: hash,
		SignedAt:    time.Now().UTC().Format(time.RFC3339),
		Signature:   base64.StdEncoding.EncodeToString(ed25519.Sign(key, message(hash))),
	}
	data, err := json.MarshalIndent(sig, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, FileName), append(data, '\n'), 0o644); err != nil { //nolint:gosec
		return nil, err
	}
	return sig, nil
}

// Read loads SKILL.sig from a skill folder (os.ErrNotExist when unsigned).
func Read(dir string) (*Signature, error) {
	data, err := os.ReadFile(filepath.Join(dir, FileName)) //nolint:gosec
	if err != nil {
		return nil, err
	}
	var sig Signature
	if err := json.Unmarshal(data, &sig); err != nil {
		return nil, fmt.Errorf("%s: %w", FileName, err)
	}
	return &sig, nil
}

// Verify checks the skill in dir against the trust store (which may be nil).
func Verify(dir string, trust *TrustStore) Result {
	sig, err := Read(dir)
	if errors.Is(err, os.ErrNotExist) {
		return Result{Status: StatusUnsigned}
	}
	if err != nil {
		return Result{Status: StatusInvalid, Detail: err.Error()}
	}
	res := Result{Publisher: sig.Publisher, KeyID: sig.KeyID, SignedAt: sig.SignedAt}

	raw, err := base64.StdEncoding.DecodeString(sig.PublicKey)
	if sig.Algorithm != algorithm || err != nil || len(raw) != ed25519.PublicKeySize {
		res.Status, res.Detail = StatusInvalid, "unsupported or damaged signature"
		return res
	}
	pub := ed25519.PublicKey(raw)
	if KeyID(pub) != sig.KeyID {
		res.Status, res.Detail = StatusInvalid, "key id does not match the embedded public key"
		return res
	}
	sigBytes, err := base64.StdEncoding.DecodeString(sig.Signature)
	if err != nil || !ed25519.Verify(pub, message(sig.ContentHash), sigBytes) {
		res.Status, res.Detail = StatusInvalid, "the signature does not verify"
		return res
	}
	actual, err := hasher.HashDirExcluding(dir, FileName)
	if err != nil {
		res.Status, res.Detail = StatusInvalid, err.Error()
		return res
	}
	if actual != sig.ContentHash {
		res.Status, res.Detail = StatusTampered, "files changed after the skill was signed"
		return res
	}
	if trust != nil && trust.Trusts(sig.KeyID) {
		res.Status = StatusTrusted
		if name := trust.PublisherFor(sig.KeyID); name != "" {
			res.Publisher = name
		}
		return res
	}
	res.Status, res.Detail = StatusUntrusted, "signed by a key you have not trusted — see 'skell trust add'"
	return res
}

// TrustedKey is one publisher in the trust store.
type TrustedKey struct {
	Publisher string `json:"publisher"`
	KeyID     string `json:"key_id"`
	PublicKey string `json:"public_key"`
	AddedAt   string `json:"added_at"`
}

// TrustStore is the set of publishers whose signatures are accepted. It lives
// in <skell home>/trusted-keys.json, so an organisation can distribute it with
// the rest of its Skell configuration.
type TrustStore struct {
	path string
	Keys []TrustedKey `json:"keys"`
}

// StorePath is where the trust store lives under a Skell home folder.
func StorePath(home string) string { return filepath.Join(home, "trusted-keys.json") }

// LoadStore reads the trust store at path; a missing file is an empty store.
func LoadStore(path string) (*TrustStore, error) {
	ts := &TrustStore{path: path}
	data, err := os.ReadFile(path) //nolint:gosec
	if errors.Is(err, os.ErrNotExist) {
		return ts, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, ts); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return ts, nil
}

// Trusts reports whether keyID is in the store.
func (t *TrustStore) Trusts(keyID string) bool {
	for _, k := range t.Keys {
		if k.KeyID == keyID {
			return true
		}
	}
	return false
}

// PublisherFor returns the name a key was trusted under.
func (t *TrustStore) PublisherFor(keyID string) string {
	for _, k := range t.Keys {
		if k.KeyID == keyID {
			return k.Publisher
		}
	}
	return ""
}

// Add trusts a publisher's public key. Re-adding the same key renames it.
func (t *TrustStore) Add(publisher string, pub ed25519.PublicKey) (TrustedKey, error) {
	if strings.TrimSpace(publisher) == "" {
		return TrustedKey{}, errors.New("a publisher name is required")
	}
	k := TrustedKey{
		Publisher: publisher, KeyID: KeyID(pub), PublicKey: EncodePublic(pub),
		AddedAt: time.Now().UTC().Format(time.RFC3339),
	}
	for i := range t.Keys {
		if t.Keys[i].KeyID == k.KeyID {
			k.AddedAt = t.Keys[i].AddedAt
			t.Keys[i] = k
			return k, t.save()
		}
	}
	t.Keys = append(t.Keys, k)
	sort.Slice(t.Keys, func(i, j int) bool { return t.Keys[i].Publisher < t.Keys[j].Publisher })
	return k, t.save()
}

// Remove drops every key matching a publisher name or key id.
func (t *TrustStore) Remove(nameOrID string) (int, error) {
	var kept []TrustedKey
	removed := 0
	for _, k := range t.Keys {
		if k.KeyID == nameOrID || strings.EqualFold(k.Publisher, nameOrID) {
			removed++
			continue
		}
		kept = append(kept, k)
	}
	t.Keys = kept
	if removed == 0 {
		return 0, fmt.Errorf("no trusted publisher %q", nameOrID)
	}
	return removed, t.save()
}

func (t *TrustStore) save() error {
	if t.path == "" {
		return errors.New("trust store has no path")
	}
	if err := os.MkdirAll(filepath.Dir(t.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(t.path, append(data, '\n'), 0o600)
}
