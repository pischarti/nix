# Neovim Configuration with Lazy.nvim and NvimTree

This is a modern Neovim configuration using the Lazy.nvim plugin manager and NvimTree file explorer.

## 🚀 Features

- **Plugin Manager**: Lazy.nvim for fast and efficient plugin management
- **File Explorer**: NvimTree for browsing and managing files
- **Fuzzy Finder**: Telescope for finding files, text, and more
- **Colorscheme**: Catppuccin Mocha theme
- **Syntax Highlighting**: TreeSitter for better syntax highlighting
- **Statusline**: Lualine for a beautiful statusline
- **Auto Completion**: Auto-pairs for bracket completion
- **Comments**: Easy commenting with Comment.nvim
- **Help**: Which-key for keybinding hints
- **Git Integration**: GitSigns for inline git diff, blame, and hunks

## 🎯 Key Bindings

### General
- **Leader key**: `<Space>`
- **Save file**: `<Ctrl-s>`
- **Quit all**: `<leader>qq`

### File Explorer (NvimTree)
- **Toggle file explorer**: `<leader>e`
- **Find current file in explorer**: `<leader>ef`

### Telescope (Fuzzy Finder)
- **Find files**: `<leader>ff`
- **Live grep**: `<leader>fg`
- **Find buffers**: `<leader>fb`
- **Help tags**: `<leader>fh`
- **Recent files**: `<leader>fr`

### Window Management
- **Move between windows**: `<Ctrl-h/j/k/l>`
- **Split window horizontally**: `<leader>-`
- **Split window vertically**: `<leader>|`
- **Close window**: `<leader>wd`

### Buffer Navigation
- **Next buffer**: `<Shift-l>` or `]b`
- **Previous buffer**: `<Shift-h>` or `[b`

### Movement
- **Move lines up/down**: `<Alt-k>/<Alt-j>` (works in visual mode too)

### Git (GitSigns)
- **Next git hunk**: `]c`
- **Previous git hunk**: `[c`
- **Stage hunk**: `<leader>gs`
- **Reset hunk**: `<leader>gr`
- **Stage buffer**: `<leader>gS`
- **Reset buffer**: `<leader>gR`
- **Preview hunk**: `<leader>gp`
- **Blame line**: `<leader>gb`
- **Toggle line blame**: `<leader>gB`
- **Diff this**: `<leader>gd`
- **Toggle deleted lines**: `<leader>gt`

## 📁 Directory Structure

```
~/.config/nvim/
├── init.lua                 # Main configuration entry point
├── lua/
│   ├── config/
│   │   ├── options.lua     # Neovim options and settings
│   │   └── keymaps.lua     # Key mappings
│   └── plugins/
│       ├── nvim-tree.lua   # NvimTree configuration
│       ├── gitsigns.lua    # Git integration with inline diff
│       └── essentials.lua  # Other essential plugins
└── README.md               # This file
```

## 🔧 Customization

### Adding New Plugins
1. Create a new file in `lua/plugins/` or add to existing files
2. Follow the Lazy.nvim plugin specification format
3. Restart Neovim or run `:Lazy sync`

### Modifying Settings
- Edit `lua/config/options.lua` for Neovim settings
- Edit `lua/config/keymaps.lua` for key bindings
- Plugin-specific settings are in their respective files in `lua/plugins/`

## 🚦 Getting Started

1. Open Neovim: `nvim`
2. Plugins will automatically install on first launch
3. Press `<leader>e` to open the file explorer
4. Press `<leader>ff` to find files
5. Press `<leader>l` to open Lazy.nvim plugin manager
6. Type `<leader>` and wait to see available keybindings

## 💡 Tips

- Use `:checkhealth` to verify your Neovim setup
- Run `:Lazy` to manage plugins
- Use `:NvimTreeToggle` to toggle the file explorer
- Press `<leader>` and wait to see all available keybindings

### Git Integration Tips
- GitSigns will automatically show git diff signs in the left column
- `+` indicates added lines, `~` for changes, `_` for deletions
- Use `]c` and `[c` to jump between git hunks (changes)
- Press `<leader>gp` to preview changes in a popup window
- Press `<leader>gb` to see git blame for the current line
- GitSigns works automatically when you open files in a git repository

## 💾 Backup and Restore Configuration

### Method 1: Git Repository (Recommended)

The best way to backup and sync your Neovim config across machines is using Git:

**Initial Setup:**
```bash
# In your nvim config directory
cd ~/.config/nvim
git init
git add .
git commit -m "Initial nvim configuration"

# Push to remote repository (GitHub/GitLab)
git remote add origin https://github.com/yourusername/nvim-config.git
git branch -M main
git push -u origin main
```

**On a new machine:**
```bash
# Clone your configuration
cd ~/.config
git clone https://github.com/yourusername/nvim-config.git nvim

# Install Neovim (if not already installed)
# macOS: brew install neovim
# Linux: use your package manager

# Start Neovim to trigger automatic plugin installation
nvim
```

**Keep configs in sync:**
```bash
# After making changes
cd ~/.config/nvim
git add .
git commit -m "Update config"
git push

# On other machines
cd ~/.config/nvim
git pull
```

### Method 2: Manual Backup

Alternatively, create a backup archive:

```bash
# Create backup
cd ~
tar -czf nvim-config-backup.tar.gz .config/nvim/

# Or copy to external drive
cp -r ~/.config/nvim/ /path/to/external/drive/nvim-backup/
```

**Restore from backup:**
```bash
# Extract backup on new machine
cd ~/.config
tar -xzf nvim-config-backup.tar.gz
```

### Important Notes

- **Plugin Installation**: Lazy.nvim will automatically install all plugins listed in `lazy-lock.json` on first startup
- **Dependencies**: Some plugins may require additional tools (LSP servers, ripgrep, fd, etc.)
- **Health Check**: Run `:checkhealth` after setup to verify everything works correctly
- **System Compatibility**: This config works on macOS, Linux, and Windows with Neovim 0.8+

### Troubleshooting

If plugins don't install automatically:
1. Open Neovim
2. Run `:Lazy sync` to manually sync plugins
3. Run `:checkhealth lazy` to diagnose issues

Enjoy your new Neovim setup! 🎉
