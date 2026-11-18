package session

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSignVerifyRoundtrip(t *testing.T) {
	pub, priv, err := GenerateKeys()
	if err != nil {
		t.Fatalf("GenerateKeys: %v", err)
	}

	data := []byte("coldcase command entry #1")
	sig := Sign(priv, data)

	if !Verify(pub, data, sig) {
		t.Fatal("Verify rejected a valid signature")
	}
	if !strings.HasPrefix(sig, "ed25519:") {
		t.Fatalf("signature missing ed25519 prefix: %q", sig[:min(12, len(sig))])
	}
}

func TestVerifyRejectsTamperedData(t *testing.T) {
	pub, priv, _ := GenerateKeys()
	data := []byte("integrity matters")
	sig := Sign(priv, data)

	if Verify(pub, []byte("integrity matters!"), sig) {
		t.Fatal("Verify accepted tampered data")
	}
}

func TestVerifyRejectsBadKeyAndFormat(t *testing.T) {
	pub, priv, _ := GenerateKeys()
	otherPub, otherPriv, _ := GenerateKeys()
	_ = otherPriv

	if Verify(pub, []byte("x"), Sign(priv, []byte("x"))[:8]+"zz") {
		t.Fatal("Verify accepted malformed hex payload")
	}
	if Verify(pub, []byte("x"), "sha256:deadbeef") {
		t.Fatal("Verify accepted foreign algorithm prefix")
	}
	if Verify(pub, []byte("x"), "") {
		t.Fatal("Verify accepted empty signature")
	}
	// Right data, wrong public key.
	if Verify(otherPub, []byte("x"), Sign(priv, []byte("x"))) {
		t.Fatal("Verify accepted signature from a different keypair")
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := DeriveKey("correct horse battery staple", nil)
	plaintext := []byte("sealed session payload")

	ciphertext, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext equals plaintext")
	}

	decrypted, err := Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("Decrypt got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptRejectsWrongKeyAndShortCiphertext(t *testing.T) {
	key := DeriveKey("key one", nil)
	ciphertext, err := Encrypt(key, []byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	if _, err := Decrypt(DeriveKey("key two", nil), ciphertext); err == nil {
		t.Fatal("Decrypt succeeded with the wrong key")
	}
	if _, err := Decrypt(key, []byte("short")); err == nil {
		t.Fatal("Decrypt succeeded on truncated ciphertext")
	}
}

func TestDeriveKeyIsDeterministicAndSalted(t *testing.T) {
	a := DeriveKey("passphrase", nil)
	b := DeriveKey("passphrase", nil)
	if !bytes.Equal(a, b) {
		t.Fatal("DeriveKey is not deterministic for the same input")
	}
	if bytes.Equal(a, DeriveKey("passphrase", []byte("other-salt"))) {
		t.Fatal("DeriveKey ignored the salt")
	}
	if len(a) != 32 {
		t.Fatalf("derived key length = %d, want 32", len(a))
	}
}

func TestCommandSigningDataIsCanonical(t *testing.T) {
	entry := CommandEntry{
		Index:            3,
		FullCommand:      "coldcase hash evidence.bin",
		WorkingDirectory: "/cases/case-001",
	}
	entry.Timestamp = entry.Timestamp.UTC()

	got := string(CommandSigningData(entry))
	want := "3|" + entry.Timestamp.Format("2006-01-02T15:04:05Z07:00") + "|coldcase hash evidence.bin|/cases/case-001"
	if got != want {
		t.Fatalf("CommandSigningData = %q, want %q", got, want)
	}
}

func TestHashFileSHA256(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "evidence.txt")
	content := []byte("known plaintext\n")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}

	sha, _, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}
	if len(sha) != 64 {
		t.Fatalf("sha256 length = %d, want 64", len(sha))
	}

	// Independent implementation must agree.
	sum := sha256.Sum256(content)
	if got, want := sha, hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("HashFile = %q, want %q", got, want)
	}

	sha2, _, err := HashFile(path)
	if err != nil || sha2 != sha {
		t.Fatalf("HashFile not deterministic: %q vs %q", sha, sha2)
	}
	if _, _, err := HashFile(filepath.Join(dir, "missing.bin")); err == nil {
		t.Fatal("HashFile succeeded on a missing file")
	}
}

func TestSignVerifyWithRawKeys(t *testing.T) {
	// Mirrors SaveKeys/LoadKeys: raw 64-byte private / 32-byte public.
	pub, priv, _ := GenerateKeys()
	if len(priv) != ed25519.PrivateKeySize || len(pub) != ed25519.PublicKeySize {
		t.Fatalf("unexpected key sizes: priv=%d pub=%d", len(priv), len(pub))
	}
	if !Verify(pub, []byte("raw key check"), Sign(priv, []byte("raw key check"))) {
		t.Fatal("roundtrip failed with raw persisted key material")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
