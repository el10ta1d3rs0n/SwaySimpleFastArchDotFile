# SwaySimpleFastArchDotFile

Hey! This is my beginner-friendly Arch Linux setup for a SwayFX desktop. It installs my selected packages and applies my configuration files to help get a Wayland desktop up and running without setting everything up by hand.

It’s best suited for a **fresh, minimal Arch Linux installation**. You don’t need to install a desktop environment first, but you do need an internet connection and a regular user account with `sudo` access.

## What this setup does

- Installs the packages listed in `packages.txt` using `yay`
- Applies configuration files from `templates/`
- Embeds and installs the wallpapers automatically
- Uses `/home/$USER` in the wallpaper path instead of a hard-coded username
- Offers to install the LazyVim starter for Neovim
- Adds Starship, Fastfetch, and completion setup to `~/.zshrc`
- Changes your login shell to Zsh
- Installs Ly and journald configuration
- Enables Ly for the next boot
- Prompts before replacing an existing Neovim configuration or rebooting

## Before you start

This setup is intended for a fresh, minimal Arch Linux installation. If you already use another login manager, **disable it before enabling Ly**. Running multiple login managers can cause problems.

Please review the scripts and configuration files before running the installer. It changes files in your home directory and under `/etc`, removes `~/.bashrc`, changes your login shell, and enables Ly. Back up anything you want to keep.

Run the installer as a **regular user**, not as root. It uses `sudo` when it needs to make system-level changes.

## Install from source

You’ll need Go and Git. Clone the repository, build the installer, and run it:

```bash
git clone https://github.com/el10ta1d3rs0n/SwaySimpleFastArchDotFile.git
cd SwaySimpleFastArchDotFile
go build -o Simple-Fast-Arch-Setup .
./Simple-Fast-Arch-Setup
```

Follow the prompts. The installer will ask for your password when it needs `sudo` access. If `yay` isn’t installed, it will install the prerequisites and build `yay`.

## Install a prebuilt binary

If a prebuilt binary is available under **Releases**:

1. Download the binary for your system from the repository’s Releases page.
2. Open a terminal in the download folder.
3. Make it executable and run it:

```bash
chmod +x Simple-Fast-Arch-Setup
./Simple-Fast-Arch-Setup
```

## Wallpapers

The wallpapers are embedded in the installer program and installed automatically. The wallpaper path uses `/home/$USER`, so it works with your home directory without requiring you to change a hard-coded username or manually copy wallpaper files.

## Keybindings

`$mod` is the **Super/Windows key**.

### Custom keybindings

| Shortcut | Action |
|---|---|
| `Super + B` | Open Helium Browser |
| `Super + E` | Open Yazi in Foot |
| `Super + V` | Open clipboard history |
| `Super + Escape` | Open the wlogout menu |
| `Super + Period` | Open bemoji |
| `Super + P` | Select a screenshot area and copy it to the clipboard |
| `Super + C` | Close the focused window |
| `Super + T` | Toggle floating mode |
| `Super + Shift + S` | Toggle sticky mode |

### Essential shortcuts

| Shortcut | Action |
|---|---|
| `Super + Enter` | Open Foot terminal |
| `Super + D` | Open Fuzzel app launcher |
| `Super + Shift + Q` | Close the focused window |
| `Super + Shift + C` | Reload the Sway config |
| `Super + Shift + E` | Show the prompt to exit Sway |
| `Super + F` | Toggle fullscreen |
| `Super + H/J/K/L` or arrow keys | Move focus between windows |
| `Super + Shift + H/J/K/L` or `Super + Shift + Arrow` | Move the focused window |
| `Super + 1–0` | Switch to workspaces 1–10 |
| `Super + Shift + 1–0` | Move the focused window to workspaces 1–10 |
| `Super + R` | Enter resize mode |

In resize mode, use `H/J/K/L` or the arrow keys to resize the focused container. Press **Enter** or **Escape** to exit resize mode.

## Customize it

- Edit `packages.txt` to change which packages get installed.
- Edit files under `templates/` to change the desktop configurations.
- Edit `ly.config` to customize Ly.

Review your changes before running the installer.

## Contribute

Have an idea or want to help make this setup even better? You’re welcome to **fork this repository**, make your changes, and open a pull request. You can also open an issue to suggest an improvement or report a problem.

If you find the project useful, please **give it a star on GitHub** ⭐ It helps others discover it!

## License

See [`LICENSE`](LICENSE) for the license.

