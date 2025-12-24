# Agent Skills Specification Summary

This is a condensed reference of the Agent Skills specification.

## Directory Structure

```
skill-name/
├── SKILL.md          # Required
├── scripts/          # Optional - executable code
├── references/       # Optional - additional docs
└── assets/           # Optional - static resources
```

## SKILL.md Format

### Required Frontmatter

```yaml
---
name: skill-name
description: What this skill does and when to use it.
---
```

### Optional Frontmatter Fields

```yaml
---
name: skill-name
description: Description here.
license: Apache-2.0
compatibility: Requires git, docker
metadata:
  author: example-org
  version: "1.0"
allowed-tools: Bash(git:*) Read
---
```

## Field Constraints

| Field | Required | Max Length | Notes |
|-------|----------|------------|-------|
| name | Yes | 64 chars | lowercase, hyphens only |
| description | Yes | 1024 chars | non-empty |
| license | No | - | license name or file ref |
| compatibility | No | 500 chars | environment requirements |
| metadata | No | - | key-value pairs |
| allowed-tools | No | - | space-delimited list |

## Name Validation Rules

- Lowercase letters, numbers, hyphens only
- No leading/trailing hyphens
- No consecutive hyphens
- Must match directory name

## Progressive Disclosure

1. **Metadata** (~100 tokens): Loaded at startup
2. **Instructions** (<5000 tokens): Loaded on activation
3. **Resources**: Loaded on demand

## Best Practices

- Keep SKILL.md under 500 lines
- Use clear, keyword-rich descriptions
- Split detailed content into references/
- Keep file references one level deep
