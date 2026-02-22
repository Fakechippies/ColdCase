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

// LoadPrivateKey from the default location
func LoadPrivateKey() (ed25519.PrivateKey, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".coldcase", "keys", "private.key")
	data, err := os.ReadFile(path)
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
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".coldcase", "keys", "public.key")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: %d", len(data))
	}
	return ed25519.PublicKey(data), nil
}

// VerifySession checks all command signatures
func (m *Manager) VerifySession(s *Session, pub ed25519.PublicKey) error {
	for i, cmd := range s.Commands {
		// Re-create the data that was signed
		data := fmt.Sprintf("%d|%s|%s|%s", cmd.Index, cmd.Timestamp.UTC().Format(time.RFC3339), cmd.FullCommand, cmd.WorkingDirectory)
		if !Verify(pub, []byte(data), cmd.Signature) {
			return fmt.Errorf("signature verification failed for command #%d", i+1)
		}
	}
	return nil
}
