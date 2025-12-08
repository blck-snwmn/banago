# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

banago is a CLI tool for image generation using the Gemini API (`gemini-3-pro-image-preview`). It manages projects with subprojects, character definitions, and generation history.

## Development Commands

```bash
# Build
go build -o banago .

# Run all tests
go test ./...

# Run a single test
go test -run TestNormalizeExt ./cmd

# Run tests in a specific package
go test ./cmd

# Lint (using golangci-lint)
golangci-lint run

# Lint specific packages
golangci-lint run ./cmd/...

# Install locally
go install .
```

## After Code Changes

Before committing, verify all checks pass:

1. **Lint**: `golangci-lint run` - must pass with no errors
2. **Test**: `go test ./...` - all tests must pass
3. **IDE diagnostics**: Check for any errors/warnings in the IDE

## banago CLI Commands

### `banago init`
Initialize a new project in the current directory.

Generated files:
- `banago.yaml` - Project configuration
- `CLAUDE.md` - Claude Code guide for image generation workflow
- `GEMINI.md` - Gemini CLI guide
- `AGENTS.md` - Common AI agent guide
- `characters/` - Directory for shared character definition files
- `subprojects/` - Directory for subprojects

Note: AI guide templates are defined in `internal/templates/ai_guides.go`.

Flags:
- `--name` - Project name (default: directory name)
- `--force` - Overwrite existing project

### `banago subproject create <name>`
Create a new subproject under `subprojects/<name>/`.

Generated files:
- `config.yaml` - Subproject configuration (character_file, input_images, aspect_ratio)
- `context.md` - Scene/costume context information
- `inputs/` - Directory for reference images
- `history/` - Directory for generation history

Flags:
- `--description` - Subproject description

### `banago subproject list`
List all subprojects in the project.

### `banago status`
Show current project/subproject status including context file, character file, input images, and history summary.

### `banago generate`
Generate images using Gemini API. Must specify prompt via `--prompt` or `--prompt-file`.

Flags:
- `-p, --prompt` - Inline prompt text
- `-F, --prompt-file` - Path to prompt file
- `-i, --image` - Additional image files (repeatable)
- `--aspect` - Aspect ratio (e.g., `1:1`, `16:9`)
- `--size` - Image size (`1K`, `2K`, `4K`)
- `-o, --output-dir` - Output directory (outside subproject, default: `dist`)
- `--prefix` - Filename prefix (outside subproject, default: `generated`)

### `banago regenerate`
Regenerate images from a history entry. Uses the same prompt and input images.

Flags:
- `--latest` - Use the latest history entry
- `--id` - Use a specific history entry UUID

### `banago history`
Show generation history of the current subproject.

Flags:
- `--limit` - Number of entries to show (default: 10)

## Architecture

### CLI Layer (`cmd/`)
Cobra-based CLI. See "banago CLI Commands" section for command details.

### Internal Packages

- `internal/config/` - YAML config handling for project (`banago.yaml`) and subproject (`config.yaml`)
- `internal/project/` - Project/subproject operations (finding root, initialization, listing)
- `internal/history/` - Generation history management with UUID v7 IDs
- `internal/generator/` - Gemini API client wrapper
- `internal/templates/` - AI guide templates (CLAUDE.md, GEMINI.md, AGENTS.md)

### Key Data Flow

1. Commands search upward for `banago.yaml` to find project root
2. Subproject context determined by checking if cwd is under `subprojects/<name>/`
3. Generation reads config, collects input images, calls Gemini API
4. Results saved to `history/<uuid>/` with metadata, prompt, and output images

### Project Structure (Runtime)

```
<project>/
├── banago.yaml        # Project config
├── CLAUDE.md          # Claude Code guide
├── GEMINI.md          # Gemini CLI guide
├── AGENTS.md          # Common AI agent guide
├── characters/        # Shared character definitions (.md)
└── subprojects/
    └── <name>/
        ├── config.yaml   # character_file, input_images, aspect_ratio
        ├── context.md    # Scene context
        ├── inputs/       # Reference images
        └── history/      # UUID v7 directories
```

## API Key

Set `GEMINI_API_KEY` environment variable or use `--api-key` flag.
