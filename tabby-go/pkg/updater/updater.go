// Package updater provides automatic update checking for Tabby Go.
//
// It checks GitHub releases for newer versions and reports
// available updates without automatically installing them.
package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	repoOwner   = "robertpelloni"
	repoName    = "tabby"
	releaseURL  = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
	userAgent   = "TabbyGo-UpdateCheck/1.0"
)

// ReleaseInfo contains information about a GitHub release.
type ReleaseInfo struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a downloadable asset from a release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UpdateStatus represents the result of an update check.
type UpdateStatus struct {
	CurrentVersion string      `json:"currentVersion"`
	LatestVersion  string      `json:"latestVersion"`
	UpdateAvailable bool       `json:"updateAvailable"`
	ReleaseInfo    *ReleaseInfo `json:"releaseInfo,omitempty"`
	Error          string      `json:"error,omitempty"`
	CheckedAt      time.Time   `json:"checkedAt"`
}

// Checker handles update checking.
type Checker struct {
	currentVersion string
	client         *http.Client
	lastStatus     *UpdateStatus
}

// NewChecker creates a new update checker.
func NewChecker(version string) *Checker {
	return &Checker{
		currentVersion: version,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// CheckForUpdates checks GitHub for a newer release.
func (c *Checker) CheckForUpdates(ctx context.Context) (*UpdateStatus, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", releaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		status := &UpdateStatus{
			CurrentVersion: c.currentVersion,
			Error:          err.Error(),
			CheckedAt:      time.Now(),
		}
		c.lastStatus = status
		return status, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
		status := &UpdateStatus{
			CurrentVersion: c.currentVersion,
			Error:          err.Error(),
			CheckedAt:      time.Now(),
		}
		c.lastStatus = status
		return status, err
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	latest := normalizeVersion(release.TagName)
	current := normalizeVersion(c.currentVersion)

	status := &UpdateStatus{
		CurrentVersion: c.currentVersion,
		LatestVersion:  latest,
		UpdateAvailable: latest > current,
		ReleaseInfo:    &release,
		CheckedAt:      time.Now(),
	}

	c.lastStatus = status
	return status, nil
}

// GetLastStatus returns the result of the most recent check.
func (c *Checker) GetLastStatus() *UpdateStatus {
	return c.lastStatus
}

// GetCurrentVersion returns the current application version.
func (c *Checker) GetCurrentVersion() string {
	return c.currentVersion
}

// normalizeVersion strips v prefix and normalizes for comparison.
func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	return v
}
