package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/looplj/axonhub/internal/build"
	"github.com/looplj/axonhub/internal/ent"
)

const (
	axonHubLiteGitHubRepo = "AkumaRealLabs/axonhub-lite"
	axonHubLiteGitHubURL  = "https://github.com/" + axonHubLiteGitHubRepo
)

// Version retrieves the system version from system settings.
// Returns empty string if not set.
func (s *SystemService) Version(ctx context.Context) (string, error) {
	value, err := s.getSystemValue(ctx, SystemKeyVersion)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}

		return "", fmt.Errorf("failed to get system version: %w", err)
	}

	return value, nil
}

// SetVersion sets the system version.
func (s *SystemService) SetVersion(ctx context.Context, version string) error {
	return s.setSystemValue(ctx, SystemKeyVersion, version)
}

// VersionCheckResult contains the result of a version check.
type VersionCheckResult struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
	ReleaseURL     string `json:"release_url"`
}

// CheckForUpdate checks if there is a newer AxonHub Lite version available on GitHub.
func (s *SystemService) CheckForUpdate(ctx context.Context) (*VersionCheckResult, error) {
	currentVersion := build.Version

	latestVersion, err := s.fetchLatestGitHubVersion(ctx, currentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest version: %w", err)
	}

	hasUpdate := s.isNewerVersion(currentVersion, latestVersion)
	releaseURL := fmt.Sprintf("%s/releases/tag/%s", axonHubLiteGitHubURL, latestVersion)

	return &VersionCheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		HasUpdate:      hasUpdate,
		ReleaseURL:     releaseURL,
	}, nil
}

// fetchLatestGitHubVersion fetches the latest stable version from GitHub.
// It skips beta and rc versions.
func (s *SystemService) fetchLatestGitHubVersion(ctx context.Context, currentVersion string) (string, error) {
	return FetchLatestGitHubVersion(ctx, currentVersion)
}

// isNewerVersion compares two semantic versions and returns true if latest is newer than current.
func (s *SystemService) isNewerVersion(current, latest string) bool {
	return IsNewerVersion(current, latest)
}

// GitHubRelease represents a GitHub release.
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
	PublishedAt time.Time `json:"published_at"`
}

// GitHubTag represents a GitHub tag.
type GitHubTag struct {
	Name string `json:"name"`
}

// releaseCooldownDuration is the time to wait after a release is published before considering it available.
// This accounts for build and upload time.
const releaseCooldownDuration = 30 * time.Minute

// FetchLatestGitHubVersion fetches the latest stable version tag from GitHub for AxonHub Lite.
// It prefers public releases, then falls back to tags because lite builds are published as draft releases.
func FetchLatestGitHubVersion(ctx context.Context, currentVersion string) (string, error) {
	latestVersion, err := FetchLatestGitHubRelease(ctx, currentVersion)
	if err == nil {
		return latestVersion, nil
	}

	latestVersion, tagErr := FetchLatestGitHubTag(ctx, currentVersion)
	if tagErr == nil {
		return latestVersion, nil
	}

	if currentVersion != "" {
		return currentVersion, nil
	}

	return "", fmt.Errorf("failed to fetch latest release: %w; failed to fetch latest tag: %w", err, tagErr)
}

// FetchLatestGitHubRelease fetches the latest stable release tag from GitHub for AxonHub Lite.
// It skips beta, rc, and prerelease versions, and waits for a cooldown period after release.
func FetchLatestGitHubRelease(ctx context.Context, currentVersion string) (string, error) {
	baseURL := fmt.Sprintf("https://api.github.com/repos/%s/releases", axonHubLiteGitHubRepo)

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("per_page", "10")
	q.Set("page", "1")
	u.RawQuery = q.Encode()
	apiURL := u.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "AxonHub-Version-Checker")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch releases: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("failed to decode releases: %w", err)
	}

	now := time.Now().UTC()

	// Find the latest stable release (not prerelease, not draft, not beta/rc, and past cooldown)
	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}

		// Only consider AxonHub Lite tags starting with "v".
		if !isUpdateCandidateForCurrentVersion(currentVersion, release.TagName) {
			continue
		}

		if isPreReleaseTag(release.TagName) {
			continue
		}

		// Check if the release has passed the cooldown period
		if now.Sub(release.PublishedAt) < releaseCooldownDuration {
			continue
		}

		return release.TagName, nil
	}

	return "", fmt.Errorf("no stable release found")
}

// FetchLatestGitHubTag fetches the latest stable version tag from GitHub.
func FetchLatestGitHubTag(ctx context.Context, currentVersion string) (string, error) {
	baseURL := fmt.Sprintf("https://api.github.com/repos/%s/tags", axonHubLiteGitHubRepo)

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("per_page", "30")
	q.Set("page", "1")
	u.RawQuery = q.Encode()
	apiURL := u.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "AxonHub-Version-Checker")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch tags: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var tags []GitHubTag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", fmt.Errorf("failed to decode tags: %w", err)
	}

	var latestVersion string
	for _, tag := range tags {
		if !isUpdateCandidateForCurrentVersion(currentVersion, tag.Name) {
			continue
		}

		if latestVersion == "" || IsNewerVersion(latestVersion, tag.Name) {
			latestVersion = tag.Name
		}
	}

	if latestVersion == "" {
		return "", fmt.Errorf("no stable tag found")
	}

	return latestVersion, nil
}

// isAxonHubTag returns true if the tag is an axonhub version tag (vX.Y.Z format).
// Tags with a service prefix (e.g., "axonclaw/v1.0.0") are not axonhub tags.
func isAxonHubTag(tag string) bool {
	if !strings.HasPrefix(tag, "v") {
		return false
	}

	_, err := semver.NewVersion(tag)
	return err == nil
}

func isUpdateCandidateForCurrentVersion(_ string, candidate string) bool {
	if !isAxonHubTag(candidate) {
		return false
	}

	if isPreReleaseTag(candidate) || isLiteReleaseTag(candidate) {
		return false
	}

	return true
}

func isLiteReleaseTag(tag string) bool {
	return strings.Contains(strings.ToLower(tag), "-lite")
}

// isPreReleaseTag checks if a version tag contains beta, rc, alpha, or similar prerelease indicators.
func isPreReleaseTag(tag string) bool {
	lowerTag := strings.ToLower(tag)
	preReleasePatterns := []string{"-beta", "-rc", "-alpha", "-dev", "-preview", "-snapshot"}

	for _, pattern := range preReleasePatterns {
		if strings.Contains(lowerTag, pattern) {
			return true
		}
	}

	return false
}

// IsNewerVersion compares two semantic versions and returns true if latest is newer than current.
// Versions are expected to be in format "vX.Y.Z" or "X.Y.Z".
func IsNewerVersion(current, latest string) bool {
	vCurrent, err := semver.NewVersion(current)
	if err != nil {
		// Handle error, maybe log it and return false
		return false
	}

	vLatest, err := semver.NewVersion(latest)
	if err != nil {
		// Handle error, maybe log it and return false
		return false
	}

	return vLatest.GreaterThan(vCurrent)
}
