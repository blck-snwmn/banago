package templates

const ClaudeMD = `# banago Project - Claude Code Guide

**Model**: ` + "`gemini-3-pro-image-preview`" + ` - Prompts must be optimized for this model.

## Quick Start: Image Generation Flow

Follow these steps to generate images:

### Step 1: Check Current State
` + "```bash" + `
banago status
` + "```" + `
This shows whether you are in a subproject and its configuration.

### Step 2: Prepare Subproject

**If no subproject exists or you need a new one:**
` + "```bash" + `
# List existing subprojects
banago subproject list

# Create a new subproject
banago subproject create <name> --description "Description"

# Navigate to the subproject
cd subprojects/<name>
` + "```" + `

**If a subproject already exists:**
` + "```bash" + `
cd subprojects/<name>
` + "```" + `

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
   - Use English (better results)
   - Be specific about costume, pose, expression, background
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
        └── history/      # Generation history
` + "```" + `

## Prompt Guidelines

1. **Language**: English recommended (better results with Gemini)
2. **Structure**: Role → Constraints → Instructions → Goal
3. **Detail**: Specific costume, pose, expression, background
4. **Restrictions**: Explicitly prohibit text generation

## Important Rules

- **Do NOT edit files in** ` + "`history/`" + `: These are archives. Create new prompts instead.
- **Do NOT modify** ` + "`inputs/`" + ` images: Keep them consistent for reproducibility.
- To improve results: Reference history, update ` + "`context.md`" + `, create new prompt.
`

const GeminiMD = `# banago Project - Gemini CLI Guide

Model: ` + "`gemini-3-pro-image-preview`" + `

## Quick Start: Image Generation Flow

### Step 1: Check Current State
` + "```bash" + `
banago status
` + "```" + `

### Step 2: Prepare Subproject

**If no subproject exists:**
` + "```bash" + `
banago subproject list                    # List existing
banago subproject create <name>           # Create new
cd subprojects/<name>                     # Navigate
` + "```" + `

**If subproject exists:**
` + "```bash" + `
cd subprojects/<name>
` + "```" + `

### Step 3: Setup Subproject

**Important**: Do NOT create or edit files with made-up information. Always ask the user first.

1. **Character file** (if needed):
   - Ask the user which character to use
   - Known characters: Research, draft, **confirm with user** before creating
   - Original characters: Ask user for details
   - Do NOT assume or guess - always confirm first

2. **Context file**:
   - **Ask the user** for scene details (costume, pose, background)
   - Edit ` + "`context.md`" + ` based on user's response
   - Do NOT assume or guess - always confirm first

3. **config.yaml** - Add or modify fields (do NOT replace entire file):
` + "```yaml" + `
character_file: <name>.md
input_images:
  - image1.png
` + "```" + `

4. **Reference images**:
   - **Ask the user** what reference images to use
   - Place provided images in ` + "`inputs/`" + `, add to ` + "`config.yaml`" + `

### Step 4: Create Prompt

1. Read ` + "`context.md`" + ` and ` + "`characters/<name>.md`" + `
2. Review reference images in ` + "`inputs/`" + `
3. Draft a prompt optimized for ` + "`gemini-3-pro-image-preview`" + `:
   - Use English (better results)
   - Be specific about costume, pose, expression, background
   - Explicitly prohibit text generation in the image
4. **Show the prompt to the user and get confirmation before generating**

### Step 5: Generate
` + "```bash" + `
banago generate --prompt "Your prompt here"
` + "```" + `

### Step 6: Iterate
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
| ` + "`banago status`" + ` | Show current state |
| ` + "`banago subproject list`" + ` | List subprojects |
| ` + "`banago subproject create <name>`" + ` | Create subproject |
| ` + "`banago history`" + ` | Show generation history |
| ` + "`banago generate`" + ` | Generate images |

## History Contents

Each ` + "`history/<uuid>/`" + ` contains:
- ` + "`prompt.txt`" + `: Prompt used
- ` + "`context.md`" + `: Context at generation time
- ` + "`character.md`" + `: Character info (if configured)
- ` + "`output_*.png`" + `: Generated images
- ` + "`meta.yaml`" + `: Metadata

## Important Rules

- **Do NOT edit** ` + "`history/`" + ` files: They are archives.
- **Do NOT modify** ` + "`inputs/`" + ` images.
- To improve: Reference history, update ` + "`context.md`" + `, create new prompt.
`

const AgentsMD = `# banago Project - AI Agent Common Guide

## Quick Start: Image Generation Flow

### Step 1: Check Current State
` + "```bash" + `
banago status
` + "```" + `
Shows whether you are in a subproject and its configuration.

### Step 2: Prepare Subproject

**If no subproject exists or you need a new one:**
` + "```bash" + `
banago subproject list                              # List existing
banago subproject create <name> --description "..." # Create new
cd subprojects/<name>                               # Navigate
` + "```" + `

**If a subproject already exists:**
` + "```bash" + `
cd subprojects/<name>
` + "```" + `

### Step 3: Setup Subproject

**Important**: Do NOT create or edit files with made-up information. Always ask the user first.

1. **Character file** (if needed):
   - Ask the user which character to use
   - Known characters: Research, draft, **confirm with user** before creating
   - Original characters: Ask user for details
   - Do NOT assume or guess - always confirm first

2. **Context file**:
   - **Ask the user** for scene details (costume, pose, background)
   - Edit ` + "`context.md`" + ` based on user's response
   - Do NOT assume or guess - always confirm first

3. **config.yaml** - Add or modify fields (do NOT replace entire file):
` + "```yaml" + `
character_file: <name>.md
input_images:
  - image1.png
` + "```" + `

4. **Reference images**:
   - **Ask the user** what reference images to use
   - Place provided images in ` + "`inputs/`" + `, add to ` + "`config.yaml`" + `

### Step 4: Create Prompt

1. Read ` + "`context.md`" + ` and ` + "`characters/<name>.md`" + `
2. Review reference images in ` + "`inputs/`" + `
3. Draft a prompt optimized for ` + "`gemini-3-pro-image-preview`" + `:
   - Use English (better results)
   - Be specific about costume, pose, expression, background
   - Explicitly prohibit text generation in the image
4. **Show the prompt to the user and get confirmation before generating**

### Step 5: Generate
` + "```bash" + `
banago generate --prompt "Your prompt here"
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

## Project Structure

` + "```" + `
<project-root>/
├── banago.yaml           # Project config
├── CLAUDE.md / GEMINI.md / AGENTS.md
├── characters/           # Shared character definitions
│   └── <name>.md
└── subprojects/
    └── <name>/
        ├── config.yaml   # character_file, input_images
        ├── context.md    # Scene/costume details
        ├── inputs/       # Reference images
        └── history/      # Generation history (UUID v7)
            └── <uuid>/
                ├── prompt.txt
                ├── context.md
                ├── character.md
                ├── meta.yaml
                └── output_*.png
` + "```" + `

## Command Reference

| Command | Description |
|---------|-------------|
| ` + "`banago status`" + ` | Show current state and subproject info |
| ` + "`banago subproject list`" + ` | List all subprojects |
| ` + "`banago subproject create <name>`" + ` | Create a new subproject |
| ` + "`banago history`" + ` | Show generation history |
| ` + "`banago generate --prompt-file <path>`" + ` | Generate with prompt file |
| ` + "`banago generate --prompt \"...\"`" + ` | Generate with inline prompt |

## Important Rules

### History Files (Required Reading)

**Do NOT edit files in ` + "`history/`" + `.**

- Files in ` + "`history/<uuid>/`" + ` are archives of generation state
- Editing breaks reproducibility
- To improve: **reference** history and create **new** prompt file
- To change context: edit ` + "`context.md`" + ` in subproject root (not in history)

### Other Rules

- Do NOT modify ` + "`inputs/`" + ` images (for consistency)
- History sorted by UUID v7 (chronological)
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
