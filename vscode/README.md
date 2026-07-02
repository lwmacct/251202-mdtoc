# mc-mdtoc

<!--TOC-->

## Table of Contents

- [Features](#features) `:22+8`
- [Requirements](#requirements) `:30+10`
- [Usage](#usage) `:40+15`
  - [Format on Save](#format-on-save) `:42+6`
  - [Manual Commands](#manual-commands) `:48+7`
- [Configuration](#configuration) `:55+31`
  - [Recommended Settings](#recommended-settings) `:67+19`
- [Example](#example) `:86+30`
- [Links](#links) `:116+5`
- [License](#license) `:121+3`

<!--TOC-->

Markdown TOC (Table of Contents) generator for Visual Studio Code.

Automatically generate and update Table of Contents for your Markdown files, similar to Prettier's format-on-save experience.

## Features

- **Auto-update on save** - Automatically update TOC when saving Markdown files
- **Manual commands** - Update or delete TOC via command palette
- **Section mode** - Generate independent TOC after each H1 heading (default)
- **Global mode** - Generate a single TOC for the entire document
- **Configurable** - Customize heading levels, list style, and more

## Requirements

This extension requires the `mc-mdtoc` CLI to be installed:

```bash
go install github.com/lwmacct/251202-mdtoc/cmd/mc-mdtoc@latest
```

The extension will prompt you to install it if not found.

## Usage

### Format on Save

1. Add `<!--TOC-->` markers in your Markdown file where you want the TOC
2. Enable format on save in settings
3. Save the file - TOC will be automatically generated/updated

### Manual Commands

- `Ctrl+Shift+T` (Mac: `Cmd+Shift+T`) - Update TOC
- Command Palette: "mc-mdtoc: Update TOC"
- Command Palette: "mc-mdtoc: Delete TOC"
- Right-click context menu in Markdown files

## Configuration

| Setting                | Default | Description                        |
| ---------------------- | ------- | ---------------------------------- |
| `mcMdtoc.enable`       | `true`  | Enable/disable the extension       |
| `mcMdtoc.formatOnSave` | `false` | Auto-update TOC on save            |
| `mcMdtoc.cliPath`      | `""`    | Custom path to mc-mdtoc executable |
| `mcMdtoc.globalMode`   | `false` | Use global mode (single TOC)       |
| `mcMdtoc.minLevel`     | `1`     | Minimum heading level (1-6)        |
| `mcMdtoc.maxLevel`     | `3`     | Maximum heading level (1-6)        |
| `mcMdtoc.ordered`      | `false` | Use ordered list (1. 2. 3.)        |

### Recommended Settings

```json
{
  "[markdown]": {
    "editor.formatOnSave": true,
    "editor.defaultFormatter": "lwmacct.mc-mdtoc"
  }
}
```

Or enable auto-update via the extension's own setting:

```json
{
  "mcMdtoc.formatOnSave": true
}
```

## Example

Before:

```markdown
# My Document

## Introduction

Content here...

## Getting Started

More content...
```

After saving:

```markdown
# My Document

## Introduction

Content here...

## Getting Started

More content...
```

## Links

- [CLI Documentation](https://github.com/lwmacct/251202-mdtoc)
- [Report Issues](https://github.com/lwmacct/251202-mdtoc/issues)

## License

MIT
