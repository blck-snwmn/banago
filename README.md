# banago

Image generation CLI using Gemini API (`gemini-3-pro-image-preview`)

## Installation

```bash
go install github.com/blck-snwmn/banago@latest
```

## Configuration

```bash
export GEMINI_API_KEY="your-api-key"
```

Or use the `--api-key` flag.

## Usage

### Initialize a project

```bash
banago init
```

### Create a subproject

```bash
banago subproject create my-project
cd subprojects/my-project
```

### Generate images

```bash
# Run inside a subproject
banago generate --prompt "description of the image"

# Use a prompt file
banago generate --prompt-file prompt.txt

# Specify additional images
banago generate --prompt "..." --image ref.png
```

### Regenerate

```bash
# Regenerate from the latest history
banago regenerate --latest

# Regenerate from a specific history
banago regenerate --id <uuid>
```

### Check status

```bash
banago status
```

### View history

```bash
banago history
banago history --limit 5
```
