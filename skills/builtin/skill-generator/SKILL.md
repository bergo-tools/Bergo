---
name: skill-generator
description: Generates Agent Skills folders that conform to the Agent Skills specification. Use when the user wants to create a new skill, scaffold a skill directory, or needs help with skill structure and format.
metadata:
  author: bergo
  version: "1.0"
---

# Skill Generator

This skill helps you create new Agent Skills that conform to the Agent Skills specification.

## When to Use

- User wants to create a new skill
- User needs to scaffold a skill directory structure
- User asks about skill format or structure

## Generation Process

### Step 1: Gather Requirements

Ask the user for:
1. **Skill name**: Must be lowercase, use hyphens, 1-64 characters (e.g., `pdf-processing`)
2. **Description**: What the skill does and when to use it (1-1024 characters)
3. **Purpose**: Detailed explanation of the skill's functionality

### Step 2: Create Directory Structure

Create the skill directory under `.bergoskills/`:

```
skill-name/
├── SKILL.md          # Required - main skill file
├── references/       # Optional - additional documentation
├── scripts/          # Optional - executable scripts
└── assets/           # Optional - static resources
```

### Step 3: Generate SKILL.md

The `SKILL.md` must contain:

1. **YAML Frontmatter** (required):
```yaml
---
name: skill-name
description: A clear description of what this skill does and when to use it.
metadata:
  author: bergo
  version: "1.0"
---
```

2. **Markdown Body**: Instructions for the agent to follow

## Naming Rules

The `name` field must:
- Be 1-64 characters
- Use only lowercase letters, numbers, and hyphens
- Not start or end with a hyphen
- Not contain consecutive hyphens (`--`)
- Match the parent directory name

✅ Valid: `pdf-processing`, `data-analysis`, `code-review`
❌ Invalid: `PDF-Processing`, `-pdf`, `pdf--processing`

## Best Practices

1. **Keep SKILL.md under 500 lines** - move detailed content to `references/`
2. **Write clear descriptions** - include keywords that help identify relevant tasks
3. **Structure for progressive disclosure**:
   - Metadata (~100 tokens): loaded at startup
   - Instructions (<5000 tokens): loaded when activated
   - Resources: loaded only when needed

## Example Output

For a skill named `api-testing`:

```yaml
---
name: api-testing
description: Tests REST APIs by sending requests and validating responses. Use when the user wants to test endpoints, verify API behavior, or debug API issues.
metadata:
  author: bergo
  version: "1.0"
---

# API Testing Skill

## Instructions
...
```

## Optional Fields Reference

| Field | Description |
|-------|-------------|
| `license` | License name or reference to LICENSE file |
| `compatibility` | Environment requirements (max 500 chars) |
| `metadata` | Key-value pairs for additional info |
| `allowed-tools` | Space-delimited list of pre-approved tools |

See [references/SPEC-SUMMARY.md](references/SPEC-SUMMARY.md) for the complete specification summary.
