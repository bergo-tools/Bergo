package version

import (
	"bergo/locales"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// VersionInfo 版本信息结构体
type VersionInfo struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Build      string
	Raw        string
}

// ParseVersion 解析版本字符串
func ParseVersion(versionStr string) (*VersionInfo, error) {
	if versionStr == "" {
		versionStr = "v0.0.0"
	}

	// 移除开头的v
	versionStr = strings.TrimPrefix(versionStr, "v")

	// 正则表达式匹配语义化版本号
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9.-]+))?(?:\+([a-zA-Z0-9.-]+))?`)
	matches := re.FindStringSubmatch(versionStr)

	if matches == nil {
		fmt.Println("null/exi")
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	preRelease := matches[4]
	build := matches[5]

	return &VersionInfo{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: preRelease,
		Build:      build,
		Raw:        versionStr,
	}, nil
}

// CompareVersions 比较两个版本
// 返回: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) (int, error) {
	ver1, err := ParseVersion(v1)
	if err != nil {
		return 0, err
	}

	ver2, err := ParseVersion(v2)
	if err != nil {
		return 0, err
	}

	// 比较主版本号
	if ver1.Major != ver2.Major {
		if ver1.Major < ver2.Major {
			return -1, nil
		}
		return 1, nil
	}

	// 比较次版本号
	if ver1.Minor != ver2.Minor {
		if ver1.Minor < ver2.Minor {
			return -1, nil
		}
		return 1, nil
	}

	// 比较修订号
	if ver1.Patch != ver2.Patch {
		if ver1.Patch < ver2.Patch {
			return -1, nil
		}
		return 1, nil
	}

	// 比较预发布版本
	if ver1.PreRelease != ver2.PreRelease {
		// 空预发布版本 > 有预发布版本
		if ver1.PreRelease == "" && ver2.PreRelease != "" {
			return 1, nil
		}
		if ver1.PreRelease != "" && ver2.PreRelease == "" {
			return -1, nil
		}
		// 比较预发布字符串
		if ver1.PreRelease < ver2.PreRelease {
			return -1, nil
		}
		if ver1.PreRelease > ver2.PreRelease {
			return 1, nil
		}
	}

	return 0, nil
}

// IsNewerVersion 检查新版本是否比当前版本新
func IsNewerVersion(current, latest string) (bool, error) {
	comparison, err := CompareVersions(current, latest)
	if err != nil {
		return false, err
	}
	return comparison < 0, nil
}

// FormatVersion 格式化版本信息显示
func FormatVersion(version, buildTime, commitHash string) string {

	return locales.Sprintf("Verison: %s ", version)
}

// 版本信息，通过编译时注入
var (
	Version    string // 版本号，格式如 v0.0.0alpha
	BuildTime  string // 构建时间
	CommitHash string // 提交哈希
)
