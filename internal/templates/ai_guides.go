package templates

const ClaudeMD = `# banago Project - Claude Code Guide

## Overview

This project uses the banago CLI for image generation workflows.
Your role is to read context information and input images, then create high-quality image generation prompts.

## Workflow

### 1. Check Context
Read the ` + "`context.md`" + ` in the current subproject.
It contains character settings, costume details, and generation goals.

Character base information is in the ` + "`characters/`" + ` directory, referenced by ` + "`character_file`" + ` in ` + "`config.yaml`" + `.

### 2. Check Input Images
Review the reference images in the ` + "`inputs/`" + ` directory.
The order is specified by ` + "`input_images`" + ` in ` + "`config.yaml`" + `.

### 3. Review History (Optional)
You can reference past generation results in ` + "`history/`" + `.
Each history entry contains:
- ` + "`prompt.txt`" + `: The prompt used
- ` + "`context.md`" + `: Context information at generation time
- ` + "`character.md`" + `: Character information at generation time (if configured)
- ` + "`output_*.png`" + `: Generated images
- ` + "`meta.yaml`" + `: Metadata

### 4. Generate
Run generation with:

` + "```bash" + `
banago generate --prompt-file <prompt-file>
` + "```" + `

## Available Commands

| Command | Description |
|---------|-------------|
| ` + "`banago status`" + ` | Show current subproject status |
| ` + "`banago history`" + ` | Show generation history |
| ` + "`banago generate --prompt \"...\"`" + ` | Generate with inline prompt |
| ` + "`banago generate --prompt-file <path>`" + ` | Generate with prompt from file |

## Prompt Guidelines

1. **Language**: English recommended (Gemini image generation works better with English)
2. **Structure**: Role setting → Constraints → Specific instructions → Goal
3. **Detail**: Be specific about costume details, poses, expressions, backgrounds
4. **Restrictions**: Explicitly prohibit text generation

## Important Notes

- **Do NOT edit history files**: Files in ` + "`history/`" + ` (prompt.txt, context.md, character.md, meta.yaml) must **never be edited**. They are archives recording the state at generation time.
- To improve prompts, create a new prompt file and run ` + "`banago generate`" + `.
- To change context, edit ` + "`context.md`" + ` in the subproject root.
`

const GeminiMD = `# banago Project - Gemini CLI Guide

## Overview

This project uses the banago CLI for image generation workflows.
Your role is to read context information and input images, then create high-quality image generation prompts.

## Model

This project uses ` + "`gemini-3-pro-image-preview`" + `.

## Workflow

### 1. Check Status
Run ` + "`banago status`" + ` to check the current subproject status.

### 2. Check Context
- ` + "`characters/<name>.md`" + `: Character base information
- ` + "`context.md`" + `: Subproject-specific information (costumes, scenes, etc.)
- ` + "`inputs/`" + `: Reference images

### 3. Generate
` + "```bash" + `
banago generate --prompt-file <path>
` + "```" + `

### 4. Improvement Cycle
` + "```bash" + `
banago history           # Check history
# Reference prompt.txt from history and create a new prompt
banago generate --prompt-file <new-prompt>
` + "```" + `

## Available Commands

| Command | Description |
|---------|-------------|
| ` + "`banago status`" + ` | Show current subproject status |
| ` + "`banago history`" + ` | Show generation history |
| ` + "`banago generate`" + ` | Generate images |

## History Contents

Each history entry (` + "`history/<uuid>/`" + `) contains:
- ` + "`prompt.txt`" + `: The prompt used
- ` + "`context.md`" + `: Context information at generation time
- ` + "`character.md`" + `: Character information at generation time (if configured)
- ` + "`output_*.png`" + `: Generated images
- ` + "`meta.yaml`" + `: Metadata

## Important Notes

- **Do NOT edit history files**: Files in ` + "`history/`" + ` must **never be edited**. They are archives recording the state at generation time.
- To improve prompts, reference history and create a new prompt file.
- To change context, edit ` + "`context.md`" + ` in the subproject root.
`

const AgentsMD = `# banago Project - AI Agent Common Guide

## Project Structure

` + "```" + `
<project-root>/
├── banago.yaml          # Project config
├── CLAUDE.md            # Claude Code guide
├── GEMINI.md            # Gemini CLI guide
├── AGENTS.md            # This file
├── characters/          # Shared character definitions
│   └── <name>.md
└── subprojects/
    └── <name>/
        ├── config.yaml   # Subproject config
        ├── context.md    # Additional info (costumes, scenes, etc.)
        ├── inputs/       # Input images
        └── history/      # Generation history (UUID v7 directories)
            └── <uuid>/
                ├── prompt.txt    # Prompt used
                ├── context.md    # Context at generation time
                ├── character.md  # Character info at generation time
                ├── meta.yaml     # Metadata
                └── output_*.png  # Generated images
` + "```" + `

## Detailed Workflow

### New Generation Flow
1. Run ` + "`banago status`" + ` to check current state
2. Read ` + "`context.md`" + ` and understand character info
3. Review reference images in ` + "`inputs/`" + `
4. Create a prompt (save to file recommended)
5. Run ` + "`banago generate --prompt-file <path>`" + `

### Improvement Flow
1. Run ` + "`banago history`" + ` to check past generations
2. Reference ` + "`prompt.txt`" + ` from the entry you want to improve
3. Update ` + "`context.md`" + ` if needed
4. Run ` + "`banago generate --prompt-file <path>`" + ` again

## Important Notes

### History File Handling (Required Reading)

**Do NOT edit files in ` + "`history/`" + `.**

- All files in ` + "`history/<uuid>/`" + ` (prompt.txt, context.md, character.md, meta.yaml) are archives recording the state at generation time
- Editing these files will make it impossible to reproduce past generations
- To improve prompts, **reference** history and create a new prompt file
- To change context, edit ` + "`context.md`" + ` in the subproject root

### Other Notes

- Do NOT modify images in ` + "`inputs/`" + ` (for history consistency)
- History is sorted by UUID v7, which is chronological
`

const DefaultContextMD = `# Context Information

Add subproject-specific information here.

## Costume/Appearance Details

(Describe costume and style details here)

## Scene Setting

(Describe background and situation here)

## Generation Notes

(Add specific points to note here)
`

const DefaultCharacterMD = `# Character Information

Add character base information here.

## Basic Profile

- Name:
- Gender:
- Age:

## Appearance

- Hair color/style:
- Eye color:
- Body type:
- Distinctive features:

## Personality/Setting

(Describe character personality and background here)
`
