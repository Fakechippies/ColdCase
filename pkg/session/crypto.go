package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

// GenerateKeys creates a new Ed25519 keypair
func GenerateKeys() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	return pub, priv, err
}

// Sign data using Ed25519
func Sign(priv ed25519.PrivateKey, data []byte) string {
	sig := ed25519.Sign(priv, data)
	return "ed25519:" + hex.EncodeToString(sig)
}

// Verify signature using Ed25519
func Verify(pub ed25519.PublicKey, data []byte, sigStr string) bool {
	if len(sigStr) < 9 || sigStr[:8] != "ed25519:" {
		return false
	}
	sig, err := hex.DecodeString(sigStr[8:])
	if err != nil {
		return false
	}
	return ed25519.Verify(pub, data, sig)
}

// DeriveKey from passphrase using PBKDF2
func DeriveKey(passphrase string, salt []byte) []byte {
	if salt == nil {
		salt = []byte("coldcase-salt") // In production, use unique salts
	}
	return pbkdf2.Key([]byte(passphrase), salt, 100000, 32, sha256.New)
}

// Encrypt data using AES-256-GCM
func Encrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// Decrypt data using AES-256-GCM
func Decrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// HashFile computes SHA256 of a file
func HashFile(path string) (sha256Hash, md5Hash string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	h256 := sha256.New()
	if _, err := io.Copy(h256, f); err != nil {
		return "", "", err
	}

	return hex.EncodeToString(h256.Sum(nil)), "", nil
}

// CommandSigningData returns the canonical byte string that is signed for a
// command entry. Both signing (Logger.LogCommand) and verification
// (Manager.VerifySession, web UI) use this helper so the signed format can
// never drift between writer and reader.
func CommandSigningData(cmd CommandEntry) []byte {
	return []byte(fmt.Sprintf("%d|%s|%s|%s",
		cmd.Index,
		cmd.Timestamp.UTC().Format(time.RFC3339),
		cmd.FullCommand,
		cmd.WorkingDirectory,
	))
}

// SessionSigningData returns the canonical byte string signed when a session
// is sealed.
func SessionSigningData(s *Session) []byte {
	return []byte(fmt.Sprintf("%s|%s|%d",
		s.ID,
		s.Created.UTC().Format(time.RFC3339),
		len(s.Commands),
	))
}

// KeysDir returns the directory where ColdCase stores its Ed25519 keypair.
func KeysDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".coldcase", "keys")
}

// SaveKeys persists an Ed25519 keypair to the default key directory.
func SaveKeys(priv ed25519.PrivateKey, pub ed25519.PublicKey) error {
	dir := KeysDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "private.key"), priv, 0600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "public.key"), pub, 0644)
}

// KeysExist reports whether a usable signing keypair is present on disk.
func KeysExist() bool {
	if _, err := LoadPrivateKey(); err != nil {
		return false
	}
	if _, err := LoadPublicKey(); err != nil {
		return false
	}
	return true
}

// LoadPrivateKey from the default location
func LoadPrivateKey() (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(filepath.Join(KeysDir(), "private.key"))
	if err != nil {
		return nil, err
	}
	if len(data) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: %d", len(data))
	}
	return ed25519.PrivateKey(data), nil
}

// LoadPublicKey from the default location
func LoadPublicKey() (ed25519.PublicKey, error) {
	data, err := os.ReadFile(filepath.Join(KeysDir(), "public.key"))
	if err != nil {
		return nil, err
	}
	if len(data) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: %d", len(data))
	}
	return ed25519.PublicKey(data), nil
}

// ParsePublicKey decodes a hex-encoded Ed25519 public key as stored on a
// session.
func ParsePublicKey(s string) (ed25519.PublicKey, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(b) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: %d", len(b))
	}
	return ed25519.PublicKey(b), nil
}

// VerifySession checks all command signatures.
func (m *Manager) VerifySession(s *Session, pub ed25519.PublicKey) error {
	for i, cmd := range s.Commands {
		if !Verify(pub, CommandSigningData(cmd), cmd.Signature) {
			return fmt.Errorf("signature verification failed for command #%d", i+1)
		}
	}
	return nil
}
