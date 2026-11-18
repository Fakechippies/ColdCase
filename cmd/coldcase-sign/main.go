// Command coldcase-sign signs and verifies release artifacts with Ed25519,
// using the exact same signing approach as ColdCase's forensic audit trail
// (pkg/session.Sign / Verify, "ed25519:<hex>" signature format).
//
// Signing (used by the release workflow):
//
//	COLDCASE_SIGNING_KEY=<hex of 64-byte ed25519 private key> \
//	    coldcase-sign sign dist/checksums.txt
//
// Verifying (used by users after downloading a release):
//
//	coldcase-sign verify -pub <hex public key> dist/checksums.txt dist/checksums.txt.sig
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"coldcase/pkg/session"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "sign":
		signCmd(os.Args[2:])
	case "verify":
		verifyCmd(os.Args[2:])
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  coldcase-sign sign <file>            # signs <file>, writes <file>.sig
                                       # key: $COLDCASE_SIGNING_KEY (hex, 64-byte ed25519 private key)
  coldcase-sign verify -pub <hex> <file> <sig-file>`)
	os.Exit(2)
}

func loadSigningKey() ed25519.PrivateKey {
	raw := strings.TrimSpace(os.Getenv("COLDCASE_SIGNING_KEY"))
	if raw == "" {
		fatal("COLDCASE_SIGNING_KEY is not set")
	}
	blob, err := hex.DecodeString(raw)
	if err != nil {
		fatal("COLDCASE_SIGNING_KEY is not valid hex: " + err.Error())
	}
	if len(blob) != ed25519.PrivateKeySize {
		fatal(fmt.Sprintf("COLDCASE_SIGNING_KEY is %d bytes, want %d (ed25519 private key)",
			len(blob), ed25519.PrivateKeySize))
	}
	return ed25519.PrivateKey(blob)
}

func signCmd(args []string) {
	fs := flag.NewFlagSet("sign", flag.ExitOnError)
	if err := fs.Parse(args); err != nil || fs.NArg() != 1 {
		usage()
	}
	path := fs.Arg(0)

	data, err := os.ReadFile(path)
	if err != nil {
		fatal(err.Error())
	}

	sig := session.Sign(loadSigningKey(), data)
	out := path + ".sig"
	if err := os.WriteFile(out, []byte(sig), 0644); err != nil {
		fatal(err.Error())
	}

	// Print the matching public key so CI can attach it to the release.
	pubHex := hex.EncodeToString(loadSigningKey().Public().(ed25519.PublicKey))
	fmt.Printf("signed %s -> %s\npublic key (hex): %s\n", path, out, pubHex)
}

func verifyCmd(args []string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	pubHex := fs.String("pub", "", "ed25519 public key as hex (required)")
	if err := fs.Parse(args); err != nil || fs.NArg() != 2 {
		usage()
	}

	if *pubHex == "" {
		fatal("missing -pub <hex public key>")
	}
	pubBlob, err := hex.DecodeString(strings.TrimSpace(*pubHex))
	if err != nil {
		fatal("public key is not valid hex: " + err.Error())
	}
	if len(pubBlob) != ed25519.PublicKeySize {
		fatal(fmt.Sprintf("public key is %d bytes, want %d", len(pubBlob), ed25519.PublicKeySize))
	}

	filePath, sigPath := fs.Arg(0), fs.Arg(1)
	data, err := os.ReadFile(filePath)
	if err != nil {
		fatal(err.Error())
	}
	sig, err := os.ReadFile(sigPath)
	if err != nil {
		fatal(err.Error())
	}

	if !session.Verify(ed25519.PublicKey(pubBlob), data, strings.TrimSpace(string(sig))) {
		fmt.Println("INVALID: signature does not match")
		os.Exit(1)
	}
	fmt.Println("OK: signature is valid")
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "coldcase-sign: "+msg)
	os.Exit(1)
}
