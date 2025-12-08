package templates

const baseGuide = `**Model**: ` + "`gemini-3-pro-image-preview`" + ` - Prompts must be optimized for this model.

## Quick Start: Image Generation Flow

Follow these steps to generate images:

### Step 1: Check Current State
` + "```bash" + `
banago status
` + "```" + `
This shows whether you are in a subproject and its configuration.

### Step 2: Prepare Subproject

1. **Run** ` + "`banago subproject list`" + ` to check existing subprojects
2. **If no suitable subproject exists, run** ` + "`banago subproject create <name>`" + ` to create one
3. **Run** ` + "`cd subprojects/<name>`" + ` to navigate into the subproject

**Do NOT skip this step.** The subproject must exist and you must be inside it before proceeding.

### Step 3: Setup Subproject

**Important**: Do NOT create or edit files with made-up information. Always ask the user first.

1. **Character file** (if needed):
   - Ask the user which character to use
   - For known characters (VTubers, anime, etc.): Research and draft details, then **confirm with user** before creating
   - For original characters: Ask the user to provide appearance and personality details
   - Do NOT assume or guess - always confirm with user before creating the file

2. **Context file**:
   - **Ask the user**: "What scene do you want to generate? Please describe costume, pose, expression, background, and any specific requirements."
   - Edit ` + "`context.md`" + ` based on user's response
   - Do NOT assume or guess - always confirm with user before editing

3. **config.yaml** - Add or modify these fields (do NOT replace the entire file):
` + "```yaml" + `
character_file: <name>.md
input_images:
  - image1.png
  - image2.jpg
` + "```" + `

4. **Reference images**:
   - **Ask the user**: "What reference images do you want to use? Please provide image files or paths."
   - Place provided images in ` + "`inputs/`" + ` and add filenames to ` + "`config.yaml`" + ` ` + "`input_images`" + `

### Step 4: Create Prompt

1. Read ` + "`context.md`" + ` and ` + "`characters/<name>.md`" + `
2. Review reference images in ` + "`inputs/`" + `
3. Draft a prompt optimized for ` + "`gemini-3-pro-image-preview`" + `:
   - **Use natural language sentences, NOT tag-based format** (e.g., NOT "1girl, blue hair, standing")
   - **Longer, detailed prompts are better** - don't be brief
   - **Markdown formatting is acceptable** (headers, lists, etc.)
   - Reference images will be sent with the prompt, so you don't need to describe the character's appearance in detail
   - Focus on: scene, pose, expression, costume changes, background, lighting
   - Write in English (better results)
   - Explicitly prohibit text generation in the image
4. **Show the prompt to the user and get confirmation before generating**

### Step 5: Generate Image
` + "```bash" + `
banago generate --prompt "Your prompt here"
# or
banago generate --prompt-file <prompt-file.txt>
` + "```" + `

### Step 6: Iterate and Improve
` + "```bash" + `
banago history
` + "```" + `
1. Review past prompts in ` + "`history/<uuid>/prompt.txt`" + `
2. Improve the prompt based on results
3. **Show improved prompt to user and get confirmation**
4. Generate again with ` + "`banago generate --prompt \"...\"`" + `

---

## Command Reference

| Command | Description |
|---------|-------------|
| ` + "`banago status`" + ` | Show current state and subproject info |
| ` + "`banago subproject list`" + ` | List all subprojects |
| ` + "`banago subproject create <name>`" + ` | Create a new subproject |
| ` + "`banago history`" + ` | Show generation history |
| ` + "`banago generate --prompt \"...\"`" + ` | Generate with inline prompt |
| ` + "`banago generate --prompt-file <path>`" + ` | Generate with prompt file |
| ` + "`banago regenerate --latest`" + ` | Regenerate with latest history |
| ` + "`banago regenerate --id <uuid>`" + ` | Regenerate with specific history |

## Project Structure

` + "```" + `
<project-root>/
├── banago.yaml           # Project config
├── characters/           # Shared character definitions
│   └── <name>.md
└── subprojects/
    └── <name>/
        ├── config.yaml   # Subproject config (character_file, input_images)
        ├── context.md    # Scene/costume details
        ├── inputs/       # Reference images
        └── history/      # Generation history (UUID v7)
            └── <uuid>/
                ├── prompt.txt    # Prompt used
                ├── context.md    # Context at generation time
                ├── character.md  # Character info (if configured)
                ├── output_*.png  # Generated images
                └── meta.yaml     # Metadata
` + "```" + `

## Important Rules

### History Files (Required Reading)

**Do NOT edit files in ` + "`history/`" + `.**

- Files in ` + "`history/<uuid>/`" + ` are archives of generation state
- Editing breaks reproducibility
- To improve: **reference** history and create **new** prompt
- To change context: edit ` + "`context.md`" + ` in subproject root (not in history)

### Other Rules

- Do NOT modify ` + "`inputs/`" + ` images (for consistency)
- History sorted by UUID v7 (chronological)
`

// ClaudeMD is the guide for Claude Code
const ClaudeMD = `# banago Project - Claude Code Guide

` + baseGuide

// GeminiMD is the guide for Gemini CLI
const GeminiMD = `# banago Project - Gemini CLI Guide

` + baseGuide

// AgentsMD is the common guide for AI agents
const AgentsMD = `# banago Project - AI Agent Guide

` + baseGuide

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
