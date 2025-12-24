package test

import (
	"bergo/skills"
	"os"
	"path/filepath"
	"testing"
)

func TestParseSkillContent(t *testing.T) {
	content := `---
name: test-skill
description: A test skill for unit testing.
---

# Test Skill

This is the body content.
`
	skill, err := skills.ParseSkillContent(content, "test-skill")
	if err != nil {
		t.Fatalf("Failed to parse skill: %v", err)
	}

	if skill.Name != "test-skill" {
		t.Errorf("Expected name 'test-skill', got '%s'", skill.Name)
	}

	if skill.Description != "A test skill for unit testing." {
		t.Errorf("Unexpected description: %s", skill.Description)
	}

	if skill.Body == "" {
		t.Error("Body should not be empty")
	}
}

func TestSkillValidation(t *testing.T) {
	tests := []struct {
		name    string
		skill   skills.Skill
		wantErr bool
	}{
		{
			name: "valid skill",
			skill: skills.Skill{
				Name:        "valid-skill",
				Description: "A valid skill",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			skill: skills.Skill{
				Name:        "",
				Description: "A skill",
			},
			wantErr: true,
		},
		{
			name: "uppercase name",
			skill: skills.Skill{
				Name:        "Invalid-Skill",
				Description: "A skill",
			},
			wantErr: true,
		},
		{
			name: "name starts with hyphen",
			skill: skills.Skill{
				Name:        "-invalid",
				Description: "A skill",
			},
			wantErr: true,
		},
		{
			name: "consecutive hyphens",
			skill: skills.Skill{
				Name:        "invalid--skill",
				Description: "A skill",
			},
			wantErr: true,
		},
		{
			name: "empty description",
			skill: skills.Skill{
				Name:        "valid-name",
				Description: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.skill.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadSkills(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "skills-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建 .bergoskills 目录
	skillsDir := filepath.Join(tmpDir, ".bergoskills", "my-skill")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建 SKILL.md
	skillContent := `---
name: my-skill
description: My test skill
---

Instructions here.
`
	skillFile := filepath.Join(skillsDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试加载
	manager := &skills.Manager{}
	manager.Clear()
	
	// 使用新的 manager 实例
	newManager := skills.GetManager()
	newManager.Clear()
	
	if err := newManager.LoadSkills(tmpDir); err != nil {
		t.Fatalf("LoadSkills failed: %v", err)
	}

	if newManager.Count() != 1 {
		t.Errorf("Expected 1 skill, got %d", newManager.Count())
	}

	skill := newManager.GetSkill("my-skill")
	if skill == nil {
		t.Error("Expected to find 'my-skill'")
	}
}
