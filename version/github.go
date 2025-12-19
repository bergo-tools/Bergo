package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GitHubRelease GitHub发布信息
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
}

// GetLatestGitHubRelease 获取GitHub最新版本
func GetLatestGitHubRelease(owner, repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Bergo-Update-Checker")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return &release, nil
}

// CheckForUpdates 检查更新
func CheckForUpdates(currentVersion, owner, repo string) (bool, *GitHubRelease, error) {
	// 如果当前版本是开发版本或未知，不检查更新
	if currentVersion == "" || currentVersion == "v0.0.0" || currentVersion == "unknown" {
		return false, nil, nil
	}

	release, err := GetLatestGitHubRelease(owner, repo)
	if err != nil {
		return false, nil, fmt.Errorf("获取最新版本失败: %w", err)
	}

	// 移除tag_name中的v前缀进行比较
	latestVersion := release.TagName
	isNewer, err := IsNewerVersion(currentVersion, latestVersion)
	fmt.Println("latestVersion:", latestVersion, isNewer)
	if err != nil {
		return false, nil, fmt.Errorf("版本比较失败: %w", err)
	}

	return isNewer, release, nil
}
