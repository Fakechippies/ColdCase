# Releasing ColdCase

Releases are fully automated: tag a commit, CI builds, signs and publishes.

## Cutting a release

1. Make sure `main` is green and `CHANGELOG.md` has an `[Unreleased]` section
   summarizing the changes since the last tag.
2. Version the changelog and commit:

   ```bash
   # move [Unreleased] items under a [x.y.z] - YYYY-MM-DD heading
   git add CHANGELOG.md
   git commit -m "chore(release): v0.1.0"
   git tag v0.1.0
   git push origin main v0.1.0
   ```

3. The `Release` workflow then:
   - builds binaries for linux/darwin/windows (amd64 + arm64) via GoReleaser,
   - creates the GitHub Release with a generated changelog,
   - signs `checksums.txt` with Ed25519 and uploads `checksums.txt.sig`,
   - publishes the release public key in the run summary.

## Signing keys

- The **private** release key lives only in the GitHub Actions secret
  `COLDCASE_SIGNING_KEY` (hex of a 64-byte Ed25519 private key).
- The matching **public** key is committed to the repository at
  `keys/release_public.key` (one line, hex).
- Signatures use the same format as ColdCase's forensic audit trail:
  `ed25519:<hex>` over the exact bytes of `checksums.txt`.

To rotate keys: generate a new pair, update the secret, commit the new public
key, and cut a release. Old releases remain verifiable with their own
published key (each release run summary states the key that signed it).

## Verifying a release

Download `checksums.txt` and `checksums.txt.sig` from the release page, plus
the public key (`keys/release_public.key` in the matching commit).

```bash
# 1. Verify the checksum manifest signature
coldcase-sign verify -pub "$(cat keys/release_public.key)" \
    checksums.txt checksums.txt.sig

# 2. Verify your downloaded archive against the signed manifest
shasum -a 256 --check checksums.txt     # macOS/BSD
sha256sum --check checksums.txt         # Linux
```

If step 1 fails, the manifest was modified after signing — do not use the
release. If step 2 fails, the archive is corrupt or was tampered with.

## Container images

Release tags also publish the forensics container image to
`ghcr.io/<owner>/<repo>` tagged `vX.Y.Z`, `X.Y` and `latest`.
