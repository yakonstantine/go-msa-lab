---
name: agent-config
description: Review, improve, and set up GitHub Copilot agent configuration files. Use when asked to review, audit, improve, or create agent definitions, skills, or custom instructions.
---

# GitHub Copilot Agent Configuration

## Files to Audit

Read all of these before producing any output:

- `.github/copilot-instructions.md` - repository-level custom instructions (applies to every request)
- `.github/agents/*.agent.md` - agent definitions
- `.github/skills/*/SKILL.md` - skill definitions

## Rules to Check

### Format

- Every `*.agent.md` file must start with a YAML frontmatter block containing `name` and `description`.
- Every `SKILL.md` file must start with a YAML frontmatter block containing `name` and `description`.
- No markdown code fences wrapping the frontmatter or file content.
- Agent file names: lowercase, hyphens, `.agent.md` suffix.
- Skill directory names and `name` frontmatter field: lowercase, hyphens.

### Description quality

The `description` field is the primary signal the agent uses to decide when to invoke a file.
It must be specific and action-oriented. Flag vague descriptions.

| Weak | Strong |
|---|---|
| `"Senior developer"` | `"Review Go code changes for correctness and production risks"` |
| `"Tech lead"` | `"Critique system design trade-offs and surface architectural risks"` |

### Token efficiency

- `copilot-instructions.md` should contain only rules that apply to nearly every request. Move task-specific guidance into agents or skills.
- Agent files should define role, behavioral constraints, and standing context only. Do not embed full checklists that belong in a skill.
- Skill files should add task-specific workflow only. They must not restate anything already in the agent or custom instructions.

### Duplication

For each piece of content, check which layer owns it:

| Content type | Right place |
|---|---|
| Coding standards applying to all tasks | `copilot-instructions.md` |
| Role definition, behavioral defaults, scope boundaries | Agent file |
| Task-specific workflow, context to load, output format | Skill file |

Flag any content found in more than one layer.

### Scope boundaries

Agents in this repository have non-overlapping responsibilities:
- `senior-dev` owns: implementation, code review, idiomatic Go, testing strategy.
- `tech-lead` owns: architecture, system design, trade-off analysis.

Flag any agent definition that blurs these boundaries or claims both scopes.

## Output Format

**Must fix** - format violations, missing required fields, or broken invocation signals.

**Should fix** - duplication, scope bleed, or token-inefficient content that degrades response quality.

**Suggestion** - optional improvements to description clarity or structure.

**Summary** - one sentence on overall configuration health.

For each finding include: file, specific location, what is wrong, and a concrete suggested edit.
If a category has no findings, omit it.