package updater

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"runtime"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"

	"github.com/blang/semver/v4"
)

var (
	// UpdateMoveAfterQuit is exported for testing
	UpdateMoveAfterQuit = true
)

// PrintCallback is a method that can print formatted text, e.g. out.Print
type PrintCallback func(context.Context, string, ...interface{})

// Update will start th interactive update assistant
func Update(ctx context.Context, currentVersion semver.Version, printf PrintCallback) error {
	if printf == nil {
		printf = func(ctx context.Context, fmt string, args ...interface{}) {
			// no-op
		}
	}

	if err := IsUpdateable(ctx); err != nil {
		printf(ctx, "Your gopass binary is externally managed. Can not update: %q", err)
		return err
	}

	dest, err := executable(ctx)
	if err != nil {
		return err
	}

	rel, err := FetchLatestRelease(ctx)
	if err != nil {
		return err
	}

	debug.Log("Current: %s - Latest: %s", currentVersion.String(), rel.Version.String())
	// binary is newer or equal to the latest release -> nothing to do
	if currentVersion.GTE(rel.Version) {
		printf(ctx, "gopass is up to date (%s)", currentVersion.String())
		if gfu := os.Getenv("GOPASS_FORCE_UPDATE"); gfu == "" {
			return nil
		}
	}

	printf(ctx, "latest version is %s", rel.Version.String())
	printf(ctx, "downloading SHA256SUMS")
	_, sha256sums, err := downloadAsset(ctx, rel.Assets, "SHA256SUMS")
	if err != nil {
		return err
	}

	out.Print(ctx, "downloading SHA256SUMS.sig")
	_, sig, err := downloadAsset(ctx, rel.Assets, "SHA256SUMS.sig")
	if err != nil {
		return err
	}

	debug.Log("verifying GPG signature ...")
	ok, err := gpgVerify(sha256sums, sig)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	if !ok {
		return fmt.Errorf("GPG signature verification failed")
	}
	printf(ctx, "GPG signature verification succeeded")

	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}

	suffix := fmt.Sprintf("%s-%s.%s", runtime.GOOS, runtime.GOARCH, ext)
	printf(ctx, "downloading gopass-%s", suffix)
	dlFilename, buf, err := downloadAsset(ctx, rel.Assets, suffix)
	if err != nil {
		return err
	}

	debug.Log("finding hashsum entry for %q", dlFilename)
	wantHash, err := findHashForFile(sha256sums, dlFilename)
	if err != nil {
		return err
	}

	debug.Log("calculating hashsum of downloaded archive ...")
	gotHash := sha256.Sum256(buf)
	if !bytes.Equal(wantHash, gotHash[:]) {
		return fmt.Errorf("SHA256 hash mismatch, want %02x, got %02x", wantHash, gotHash)
	}

	debug.Log("hashsums match!")
	printf(ctx, "downloaded gopass-%s", suffix)

	debug.Log("extracting binary from tarball ...")
	size, err := extractFile(buf, dlFilename, dest)
	if err != nil {
		return err
	}
	debug.Log("extracted %q to %q", dlFilename, dest)

	printf(ctx, "saved %d bytes to %s", size, dest)
	printf(ctx, "successfully updated gopass to version %s", rel.Version.String())
	return nil
}
