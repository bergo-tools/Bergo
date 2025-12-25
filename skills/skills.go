package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Skill 表示一个 Agent Skill
type Skill struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	License       string            `yaml:"license,omitempty"`
	Compatibility string            `yaml:"compatibility,omitempty"`
	Metadata      map[string]string `yaml:"metadata,omitempty"`
	AllowedTools  string            `yaml:"allowed-tools,omitempty"`

	// 非 YAML 字段
	Body string `yaml:"-"` // Markdown 内容
	Path string `yaml:"-"` // skill 目录路径
}

// SkillsDir 是 skills 目录名
const SkillsDir = ".bergoskills"

// GetSkillsPath 获取用户主目录下的 skills 路径
func GetSkillsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, SkillsDir), nil
}

// nameRegex 用于验证 skill name
var nameRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// Validate 验证 skill 是否符合规范
func (s *Skill) Validate() error {
	// 验证 name
	if s.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(s.Name) > 64 {
		return fmt.Errorf("name must be at most 64 characters")
	}
	if !nameRegex.MatchString(s.Name) {
		return fmt.Errorf("name must contain only lowercase letters, numbers, and hyphens, cannot start/end with hyphen or have consecutive hyphens")
	}

	// 验证 description
	if s.Description == "" {
		return fmt.Errorf("description is required")
	}

	return nil
}

// ParseSkillFile 解析 SKILL.md 文件
func ParseSkillFile(path string) (*Skill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill file: %w", err)
	}

	return ParseSkillContent(string(content), filepath.Dir(path))
}

// ParseSkillContent 解析 SKILL.md 内容
func ParseSkillContent(content string, skillPath string) (*Skill, error) {
	// 检查是否以 --- 开头
	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		return nil, fmt.Errorf("SKILL.md must start with YAML frontmatter (---)")
	}

	// 分离 frontmatter 和 body
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid SKILL.md format: missing closing ---")
	}

	frontmatter := strings.TrimSpace(parts[1])
	body := strings.TrimSpace(parts[2])

	// 解析 YAML frontmatter
	skill := &Skill{}
	if err := yaml.Unmarshal([]byte(frontmatter), skill); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	skill.Body = body
	skill.Path = skillPath

	// 验证 skill
	if err := skill.Validate(); err != nil {
		return nil, fmt.Errorf("skill validation failed: %w", err)
	}

	// 验证 name 与目录名匹配
	dirName := filepath.Base(skillPath)
	if skill.Name != dirName {
		return nil, fmt.Errorf("skill name '%s' must match directory name '%s'", skill.Name, dirName)
	}

	return skill, nil
}
