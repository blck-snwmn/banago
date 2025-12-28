package templates

const baseGuide = `**Model**: ` + "`gemini-3-pro-image-preview`" + ` - Prompts must be optimized for this model.

## Important: Your Image Reading Capability

**You CAN read and understand image files.** Do not hesitate to view images directly.

- You are powered by a **multimodal model** that understands both text and images
- Reading a file (whether text or image) is just file I/O - the CLI handles it the same way
- You SHOULD proactively read image files to:
  - Understand reference images in ` + "`inputs/`" + `
  - Review generated images in ` + "`history/<uuid>/output_*.png`" + `
  - Check edited images in ` + "`history/<uuid>/edits/<edit-uuid>/output_*.png`" + `
  - Analyze character appearance from existing images
- **Do NOT assume you cannot read images** - you absolutely can and should

When reviewing generation results or preparing prompts, actively read the relevant image files to provide better feedback and create more accurate prompts.

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
2. **Read the reference images** in ` + "`inputs/`" + ` directly
   - You CAN read image files - do not skip this step
   - Understand the character's appearance, style, and details from the images
   - This helps you create more accurate and consistent prompts
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

### Step 6: Review and Decide Next Action

After generation, you MUST review the results and decide on the appropriate action.

#### 6.1: Review Generated Images
` + "```bash" + `
banago history
` + "```" + `
1. **Read the generated images** in ` + "`history/<uuid>/output_*.png`" + `
   - You CAN and SHOULD view images directly - do not skip this step
2. **Read the prompt** used for generation in ` + "`history/<uuid>/prompt.txt`" + `
3. **Compare**: Does the generated image match the prompt's intent?

#### 6.2: Decide Action Based on Results

| Result | Action | Command |
|--------|--------|---------|
| ✅ Perfect | Done! Inform the user | - |
| ❌ Completely different | Revise prompt → New generation | ` + "`banago generate`" + ` |
| ⚠️ Partially wrong | Edit specific parts | ` + "`banago edit`" + ` |

**Choose "New generation" (` + "`generate`" + `) when:**
- Overall composition is wrong (e.g., wrong pose, wrong scene)
- Character appearance is fundamentally incorrect
- Style or atmosphere doesn't match at all
- Multiple major issues exist

**Choose "Edit" (` + "`edit`" + `) when:**
- Small details are wrong (e.g., button color, accessory)
- Minor adjustments needed (e.g., lighting, expression tweak)
- Overall image is good but one specific element needs fixing

Note: ` + "`regenerate`" + ` command reuses the same prompt. Use ` + "`generate`" + ` when you need to revise the prompt.

#### 6.3a: If New Generation (Major Issues)

1. Analyze what went wrong in the current image
2. **Read the current prompt file** (` + "`prompt.txt`" + ` in subproject root)
3. Revise the prompt to address the issues
4. **Show the revised prompt to user and get confirmation**
5. Generate again:
` + "```bash" + `
banago generate --prompt-file prompt.txt
` + "```" + `
6. Return to Step 6.1 to review the new result

#### 6.3b: If Editing (Minor Issues)

1. Identify the specific part that needs fixing
2. Create a focused edit prompt describing the change
3. Run the edit command:
` + "```bash" + `
# Edit the latest generated image
banago edit --latest -p "Change the button color to red"

# Edit a specific history entry
banago edit --id <uuid> -p "Fix the background lighting"

# Chain edits (edit a previously edited image)
banago edit --latest --edit-latest -p "Further adjust the shadows"
` + "```" + `
4. **Read the edited images** in ` + "`history/<uuid>/edits/<edit-uuid>/output_*.png`" + `
5. If still not right, either:
   - Chain another edit (for minor adjustments)
   - Go back to regeneration (if edits aren't working)

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
| ` + "`banago serve`" + ` | Browse generated images in browser |

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
