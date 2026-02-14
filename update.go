package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var githubAPIBase = "https://api.github.com"

const githubRepo = "davidbudnick/redis-tui"

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func runUpdate(currentVersion string) error {
	if currentVersion == "dev" || !isSemver(currentVersion) {
		return fmt.Errorf("cannot self-update a development build (version=%q); use the install script instead:\n  curl -fsSL https://raw.githubusercontent.com/davidbudnick/redis-tui/main/install.sh | bash", currentVersion)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("could not resolve executable path: %w", err)
	}

	if isHomebrew(execPath) {
		return fmt.Errorf("this binary was installed via Homebrew; please update with:\n  brew upgrade redis-tui")
	}

	if err := checkWriteAccess(execPath); err != nil {
		return fmt.Errorf("no write permission for %s; try:\n  sudo redis-tui --update", execPath)
	}

	latest, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch latest version: %w", err)
	}

	if strings.TrimPrefix(latest, "v") == strings.TrimPrefix(currentVersion, "v") {
		fmt.Printf("Already up to date (v%s).\n", strings.TrimPrefix(currentVersion, "v"))
		return nil
	}

	ver := strings.TrimPrefix(latest, "v")
	archive := archiveName(ver, runtime.GOOS, runtime.GOARCH)
	baseURL := fmt.Sprintf("https://github.com/%s/releases/download/%s", githubRepo, latest)
	archiveURL := baseURL + "/" + archive
	checksumURL := baseURL + "/checksums.txt"

	tmpDir, err := os.MkdirTemp("", "redis-tui-update-*")
	if err != nil {
		return fmt.Errorf("could not create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, archive)
	checksumPath := filepath.Join(tmpDir, "checksums.txt")

	fmt.Printf("Downloading redis-tui v%s...\n", ver)

	if err := downloadFile(archiveURL, archivePath); err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}
	if err := downloadFile(checksumURL, checksumPath); err != nil {
		return fmt.Errorf("failed to download checksums: %w", err)
	}

	if err := verifyChecksum(archivePath, checksumPath, archive); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	newBinaryPath := filepath.Join(tmpDir, "redis-tui")
	if err := extractBinary(archivePath, newBinaryPath); err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	if err := replaceBinary(execPath, newBinaryPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("Successfully updated to v%s.\n", ver)
	return nil
}

func fetchLatestVersion() (string, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", githubAPIBase, githubRepo)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("empty tag_name in response")
	}

	return release.TagName, nil
}

func archiveName(ver, goos, goarch string) string {
	osName := strings.ToUpper(goos[:1]) + goos[1:]
	arch := goarch
	if goarch == "amd64" {
		arch = "x86_64"
	}

	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}

	return fmt.Sprintf("redis-tui_%s_%s_%s.%s", ver, osName, arch, ext)
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	return nil
}

func verifyChecksum(archivePath, checksumPath, archiveFilename string) error {
	data, err := os.ReadFile(checksumPath)
	if err != nil {
		return fmt.Errorf("could not read checksums file: %w", err)
	}

	var expectedHash string
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == archiveFilename {
			expectedHash = parts[0]
			break
		}
	}

	if expectedHash == "" {
		return fmt.Errorf("no checksum found for %s", archiveFilename)
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("could not open archive: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("could not hash archive: %w", err)
	}

	actualHash := hex.EncodeToString(h.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

func extractBinary(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("could not open archive: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("could not open gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("could not read tar entry: %w", err)
		}

		if filepath.Base(hdr.Name) == "redis-tui" && hdr.Typeflag == tar.TypeReg {
			out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return fmt.Errorf("could not create binary: %w", err)
			}
			defer out.Close()

			if _, err := io.Copy(out, tr); err != nil {
				return fmt.Errorf("could not write binary: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("binary not found in archive")
}

func replaceBinary(currentPath, newPath string) error {
	oldPath := currentPath + ".old"

	if err := os.Rename(currentPath, oldPath); err != nil {
		return fmt.Errorf("could not back up current binary: %w", err)
	}

	if err := os.Rename(newPath, currentPath); err != nil {
		// Rollback: restore the old binary
		_ = os.Rename(oldPath, currentPath)
		return fmt.Errorf("could not install new binary: %w", err)
	}

	_ = os.Remove(oldPath)
	return nil
}

func isSemver(s string) bool {
	s = strings.TrimPrefix(s, "v")
	matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+`, s)
	return matched
}

func isHomebrew(path string) bool {
	return strings.Contains(path, "/Cellar/") || strings.Contains(path, "/homebrew/")
}

func checkWriteAccess(path string) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".redis-tui-write-check-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	tmp.Close()
	return os.Remove(name)
}
