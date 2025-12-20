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
   - Run ` + "`ls characters/`" + ` to check existing character files
   - If the character file exists, use it
   - If not, search the web for character details (e.g., ` + "`<character name> appearance features costume`" + `)
   - Draft a character file based on search results, then **confirm with user** before creating
   - For original characters: Ask the user to provide appearance and personality details
   - Do NOT assume or guess - always confirm with user before creating the file

2. **Context file**:
   - **Ask the user**: "What scene do you want to generate? Please describe costume, pose, expression, background, and any specific requirements."
   - **Read the existing file first**, then edit based on user's response
   - Do NOT assume or guess - always confirm with user before editing

3. **config.yaml**:
   - **Read the existing file first** to understand current settings
   - **IMPORTANT: Do NOT replace the entire file** - only add or modify specific fields
   - **Keep all existing fields** (` + "`version`" + `, ` + "`name`" + `, ` + "`created_at`" + `, ` + "`context_file`" + `, etc.)
   - Fields to add/modify:
` + "```yaml" + `
# character_file: filename only, NO path (e.g., "usada_pekora.md", NOT "../../characters/usada_pekora.md")
character_file: <name>.md
# input_images: filenames in inputs/ directory
input_images:
  - image1.png
  - image2.jpg
` + "```" + `
   - Example of a complete config.yaml (for reference only):
` + "```yaml" + `
version: "1.0"
name: example
created_at: "2025-01-01T00:00:00Z"
context_file: context.md
character_file: usada_pekora.md    # filename only
input_images:
  - ref1.png
  - ref2.jpg
` + "```" + `

4. **Reference images**:
   - **Ask the user**: "What reference images do you want to use? Please provide image files or paths."
   - Place provided images in ` + "`inputs/`" + ` and add filenames to ` + "`config.yaml`" + ` ` + "`input_images`" + `

### Step 4: Create Prompt File

**Always save prompts to a file.** This prevents context loss during long conversations.

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
4. **Save the prompt to a file** (e.g., ` + "`prompt.txt`" + ` in the subproject directory)
5. **Show the prompt to the user and get confirmation before generating**

### Step 5: Generate Image

**Always use ` + "`--prompt-file`" + `** to generate from the saved prompt file:
` + "```bash" + `
banago generate --prompt-file prompt.txt
` + "```" + `

### Step 6: Iterate and Improve
` + "```bash" + `
banago history
` + "```" + `
1. Review past prompts in ` + "`history/<uuid>/prompt.txt`" + ` (read-only snapshots)
2. **Read the current prompt file first** before making changes
3. Edit the prompt file based on results
4. **Show improved prompt to user and get confirmation**
5. Generate again with ` + "`banago generate --prompt-file prompt.txt`" + `

### Step 7: Edit Generated Images (Optional)

Use ` + "`banago edit`" + ` for small fixes (e.g., wrong button color, minor adjustments):

` + "```bash" + `
# Edit the latest generated image
banago edit --latest -p "Change the button color to red"

# Edit a specific history entry
banago edit --id <uuid> -p "Fix the background lighting"

# Chain edits (edit an edited image)
banago edit --latest --edit-latest -p "Further adjust the shadows"
` + "```" + `

**Important**: Edit is for small fixes, not regeneration. For major changes, use ` + "`generate`" + ` with an improved prompt.

---

## Command Reference

| Command | Description |
|---------|-------------|
| ` + "`banago status`" + ` | Show current state and subproject info |
| ` + "`banago subproject list`" + ` | List all subprojects |
| ` + "`banago subproject create <name>`" + ` | Create a new subproject |
| ` + "`banago history`" + ` | Show generation history |
| ` + "`banago generate --prompt-file <path>`" + ` | Generate with prompt file (recommended) |
| ` + "`banago regenerate --latest`" + ` | Regenerate with latest history |
| ` + "`banago regenerate --id <uuid>`" + ` | Regenerate with specific history |
| ` + "`banago edit --latest -p \"...\"`" + ` | Edit latest generated image |
| ` + "`banago edit --latest --edit-latest -p \"...\"`" + ` | Edit latest edit result |
| ` + "`banago edit --id <uuid> -p \"...\"`" + ` | Edit specific history entry |

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
        ├── prompt.txt    # Current prompt (editable)
        ├── inputs/       # Reference images
        └── history/      # Generation history (UUID v7)
            └── <uuid>/
                ├── prompt.txt    # Prompt snapshot (read-only)
                ├── context.md    # Context at generation time
                ├── character.md  # Character info (if configured)
                ├── output_*.png  # Generated images
                ├── meta.yaml     # Metadata
                └── edits/        # Edit history
                    └── <edit-uuid>/
                        ├── edit-prompt.txt  # Edit prompt
                        ├── edit-meta.yaml   # Edit metadata
                        └── output_*.png     # Edited images
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

