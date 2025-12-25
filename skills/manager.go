package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Manager 管理所有加载的 skills
type Manager struct {
	skills map[string]*Skill
	mu     sync.RWMutex
}

var (
	globalManager *Manager
	once          sync.Once
)

// GetManager 获取全局 skills 管理器
func GetManager() *Manager {
	once.Do(func() {
		globalManager = &Manager{
			skills: make(map[string]*Skill),
		}
	})
	return globalManager
}

func (m *Manager) GetSkillsPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("failed to get user home directory: %w", err))
	}
	return filepath.Join(homeDir, ".bergoskills")
}

// LoadBuiltinSkills 释放内置skills到用户主目录
func (m *Manager) LoadBuiltinSkills() error {
	// 释放内置skills到用户主目录的.bergoskills下
	if err := ExtractBuiltinSkills(); err != nil {
		return fmt.Errorf("failed to extract builtin skills: %w", err)
	}
	return nil
}

// LoadSkills 从用户主目录加载所有 skills
func (m *Manager) LoadSkills() error {
	skillsPath, err := GetSkillsPath()
	if err != nil {
		return err
	}
	return m.loadSkillsFromDir(skillsPath)
}

// LoadSkillsFromPath 从指定路径加载所有 skills（主要用于测试）
func (m *Manager) LoadSkillsFromPath(skillsPath string) error {
	return m.loadSkillsFromDir(skillsPath)
}

// loadSkillsFromDir 从指定的skills目录加载所有skills
func (m *Manager) loadSkillsFromDir(skillsPath string) error {
	// 检查目录是否存在
	info, err := os.Stat(skillsPath)
	if os.IsNotExist(err) {
		return nil // 目录不存在，不是错误
	}
	if err != nil {
		return fmt.Errorf("failed to stat skills directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", skillsPath)
	}

	// 遍历 skills 目录
	entries, err := os.ReadDir(skillsPath)
	if err != nil {
		return fmt.Errorf("failed to read skills directory: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillDir := filepath.Join(skillsPath, entry.Name())
		skillFile := filepath.Join(skillDir, "SKILL.md")

		// 检查 SKILL.md 是否存在
		if _, err := os.Stat(skillFile); os.IsNotExist(err) {
			continue
		}

		skill, err := ParseSkillFile(skillFile)
		if err != nil {
			// 记录错误但继续加载其他 skills
			fmt.Fprintf(os.Stderr, "Warning: failed to load skill %s: %v\n", entry.Name(), err)
			continue
		}

		m.skills[skill.Name] = skill
	}

	return nil
}

// GetSkill 获取指定名称的 skill
func (m *Manager) GetSkill(name string) *Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.skills[name]
}

// GetAllSkills 获取所有 skills
func (m *Manager) GetAllSkills() []*Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	skills := make([]*Skill, 0, len(m.skills))
	for _, skill := range m.skills {
		skills = append(skills, skill)
	}
	return skills
}

// GetSkillsSummary 获取所有 skills 的摘要信息（用于注入 system prompt）
func (m *Manager) GetSkillsSummary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Available Skills\n\n")

	for _, skill := range m.skills {
		sb.WriteString(fmt.Sprintf("### %s\n", skill.Name))
		sb.WriteString(fmt.Sprintf("%s\n\n", skill.Description))
	}

	return sb.String()
}

// GetSkillDetail 获取 skill 的详细内容（包括 body）
func (m *Manager) GetSkillDetail(name string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	skill, ok := m.skills[name]
	if !ok {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Skill: %s\n\n", skill.Name))
	sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", skill.Description))

	if skill.Compatibility != "" {
		sb.WriteString(fmt.Sprintf("**Compatibility:** %s\n\n", skill.Compatibility))
	}

	if skill.Body != "" {
		sb.WriteString("## Instructions\n\n")
		sb.WriteString(skill.Body)
	}

	return sb.String()
}

// Clear 清空所有 skills
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.skills = make(map[string]*Skill)
}

// Count 返回 skills 数量
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.skills)
}
