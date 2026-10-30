package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed packages.txt templates ly.config
var embeddedFiles embed.FS

func status(message string) {
	fmt.Printf("\n[+] %s\n\n", message)
}

func run(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fail(message string, err error) {
	fmt.Fprintf(os.Stderr, "[-] %s: %v\n", message, err)
	os.Exit(1)
}

func readPackages() ([]string, error) {
	file, err := embeddedFiles.ReadFile("packages.txt")
	if err != nil {
		return nil, err
	}

	var packages []string
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(file)))

	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name == "" || strings.HasPrefix(name, "#") {
			continue
		}
		if strings.HasPrefix(name, "-") {
			return nil, fmt.Errorf("invalid package name: %q", name)
		}
		if !seen[name] {
			packages = append(packages, name)
			seen[name] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(packages) == 0 {
		return nil, fmt.Errorf("packages.txt contains no package names")
	}

	return packages, nil
}

func installYay() {
	if _, err := exec.LookPath("yay"); err == nil {
		status("yay is already installed")
		return
	}

	if _, err := exec.LookPath("sudo"); err != nil {
		fail("sudo is required to install prerequisites", err)
	}

	status("Installing build prerequisites")
	if err := run("", "sudo", "pacman", "-S", "--needed", "git", "base-devel"); err != nil {
		fail("Could not install prerequisites", err)
	}

	workDir, err := os.MkdirTemp("", "yay-build-*")
	if err != nil {
		fail("Could not create temporary directory", err)
	}
	defer os.RemoveAll(workDir)

	yayDir := filepath.Join(workDir, "yay")

	status("Downloading yay")
	if err := run("", "git", "clone", "--depth=1", "https://aur.archlinux.org/yay.git", yayDir); err != nil {
		fail("Could not download yay", err)
	}

	status("Installing yay")
	if err := run(yayDir, "makepkg", "-si"); err != nil {
		fail("Could not build or install yay", err)
	}
}

func installPackages(packages []string) {
	var missing []string

	for _, name := range packages {
		cmd := exec.Command("yay", "-Q", name)
		if cmd.Run() == nil {
			fmt.Printf("[+] %s is already installed\n", name)
		} else {
			missing = append(missing, name)
		}
	}

	if len(missing) == 0 {
		status("All listed packages are already installed")
		return
	}

	status("Installing missing packages")
	args := append([]string{"-S", "--needed"}, missing...)
	if err := run("", "yay", args...); err != nil {
		fail("Could not install packages", err)
	}
}

func installTemplates() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fail("Could not find the user config directory", err)
	}

	status("Applying configuration templates")

	err = fs.WalkDir(embeddedFiles, "templates", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel("templates", path)
		if err != nil {
			return err
		}
		relativePath = filepath.ToSlash(relativePath)

		// Neovim is installed separately; mimeapps.list is intentionally ignored.
		if relativePath == "nvim" ||
			strings.HasPrefix(relativePath, "nvim/") ||
			relativePath == "mimeapps.list" {
			return nil
		}

		destination := filepath.Join(configDir, filepath.FromSlash(relativePath))

		if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
			return err
		}

		content, err := embeddedFiles.ReadFile(path)
		if err != nil {
			return err
		}

		// Replace an existing file or symlink at this destination.
		if err := os.Remove(destination); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := os.WriteFile(destination, content, 0644); err != nil {
			return err
		}

		fmt.Printf("[+] Applied %s\n", destination)
		return nil
	})

	if err != nil {
		fail("Could not apply configuration templates", err)
	}
}

func installLyConfig() {
	if _, err := exec.LookPath("sudo"); err != nil {
		fail("sudo is required to install Ly configuration", err)
	}

	content, err := embeddedFiles.ReadFile("ly.config")
	if err != nil {
		fail("Could not read embedded Ly configuration", err)
	}

	// Write to a temporary user-owned file, then use sudo to install it.
	tempFile, err := os.CreateTemp("", "ly-config-*")
	if err != nil {
		fail("Could not create temporary Ly configuration file", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.Write(content); err != nil {
		tempFile.Close()
		fail("Could not write temporary Ly configuration file", err)
	}
	if err := tempFile.Close(); err != nil {
		fail("Could not close temporary Ly configuration file", err)
	}

	status("Installing Ly configuration")
	if err := run("", "sudo", "install", "-D", "-m", "0644", tempPath, "/etc/ly/config.ini"); err != nil {
		fail("Could not install Ly configuration", err)
	}
}

func installJournalLimit() {
	const content = `[Journal]
SystemMaxUse=50M
`

	tempFile, err := os.CreateTemp("", "journald-limit-*")
	if err != nil {
		fail("Could not create temporary journald configuration", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.WriteString(content); err != nil {
		tempFile.Close()
		fail("Could not write temporary journald configuration", err)
	}
	if err := tempFile.Close(); err != nil {
		fail("Could not close temporary journald configuration", err)
	}

	status("Setting the system journal size limit to 50 MB")
	if err := run(
		"",
		"sudo", "install", "-D", "-m", "0644",
		tempPath,
		"/etc/systemd/journald.conf.d/50-local-size-limit.conf",
	); err != nil {
		fail("Could not install journald configuration", err)
	}

	if err := run("", "sudo", "systemctl", "restart", "systemd-journald"); err != nil {
		fail("Could not restart systemd-journald", err)
	}
}

func askToOverwriteNvim() bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("An existing Neovim config was found. Overwrite it with LazyVim? Type yes or no: ")

		answer, err := reader.ReadString('\n')
		if err != nil && len(answer) == 0 {
			fmt.Fprintln(os.Stderr, "\n[-] Could not read your response; keeping the existing Neovim config.")
			return false
		}

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "yes", "y":
			return true
		case "no", "n":
			return false
		default:
			fmt.Println("Please enter yes or no.")
		}
	}
}

func installLazyVim() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fail("Could not find the user config directory", err)
	}

	nvimConfig := filepath.Join(configDir, "nvim")

	if _, err := os.Lstat(nvimConfig); err == nil {
		if !askToOverwriteNvim() {
			fmt.Println("[+] Keeping the existing Neovim config")
			return
		}
	} else if !os.IsNotExist(err) {
		fail("Could not check the existing Neovim config", err)
	}

	status("Installing LazyVim")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		fail("Could not create the config directory", err)
	}

	// Replace the existing config only after the user confirms.
	if err := os.RemoveAll(nvimConfig); err != nil {
		fail("Could not remove the existing Neovim config", err)
	}

	if err := run(
		"",
		"git",
		"clone",
		"--depth=1",
		"https://github.com/LazyVim/starter",
		nvimConfig,
	); err != nil {
		fail("Could not download the LazyVim starter", err)
	}

	if err := os.RemoveAll(filepath.Join(nvimConfig, ".git")); err != nil {
		fail("Could not remove the starter's Git metadata", err)
	}

	fmt.Println("[+] LazyVim starter installed; launch nvim to finish setting it up")
}

func installZshConfig() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fail("Could not find the user's home directory", err)
	}

	zshrcPath := filepath.Join(homeDir, ".zshrc")

	const config = `eval "$(starship init zsh)"

HISTFILE=/dev/null
SAVEHIST=0
HISTSIZE=0

# Show Fastfetch once per interactive shell, if installed
if [[ -o interactive ]] && [[ -z ${FASTFETCH_SHOWN:-} ]] && (( $+commands[fastfetch] )); then
  typeset -g FASTFETCH_SHOWN=1
  fastfetch
fi

# Initialize the Zsh completion system
autoload -Uz compinit
compinit

# Load the completion module for menu control
zmodload zsh/complist

# Start interactive menu selection with Tab
zstyle ':completion:*' menu select

# High-contrast color for the selected completion item
zstyle ':completion:*:default' list-colors ${(s.:.)LS_COLORS} "ma=48;5;33;38;5;255;1"

# Tab moves forward; Shift+Tab moves backward in the selection menu
bindkey -M menuselect '^I' menu-complete
bindkey -M menuselect '[Z' reverse-menu-complete
`

	existing, err := os.ReadFile(zshrcPath)
	if err != nil && !os.IsNotExist(err) {
		fail("Could not read ~/.zshrc", err)
	}

	existingText := string(existing)
	hasStarship := strings.Contains(existingText, `eval "$(starship init zsh)"`)
	hasCompletion := strings.Contains(existingText, "zstyle ':completion:*' menu select")

	if hasStarship && hasCompletion {
		fmt.Println("[+] Starship, Fastfetch, and completion setup already exists in ~/.zshrc")
		return
	}

	file, err := os.OpenFile(zshrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fail("Could not open ~/.zshrc", err)
	}

	if len(existing) > 0 && !strings.HasSuffix(existingText, "\n") {
		if _, err := file.WriteString("\n"); err != nil {
			file.Close()
			fail("Could not update ~/.zshrc", err)
		}
	}

	if _, err := file.WriteString("\n" + config); err != nil {
		file.Close()
		fail("Could not update ~/.zshrc", err)
	}

	if err := file.Close(); err != nil {
		fail("Could not close ~/.zshrc", err)
	}

	status("Added Starship, Fastfetch, and completion setup to ~/.zshrc")
}

func setZshAsLoginShell() {
	const zshPath = "/usr/bin/zsh"

	if _, err := os.Stat(zshPath); err != nil {
		fail("Zsh is not installed at "+zshPath, err)
	}

	status("Changing the login shell to Zsh")
	if err := run("", "chsh", "-s", zshPath); err != nil {
		fail("Could not change the login shell to Zsh", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fail("Could not find the user's home directory", err)
	}

	bashrcPath := filepath.Join(homeDir, ".bashrc")
	if err := os.Remove(bashrcPath); err != nil && !os.IsNotExist(err) {
		fail("Could not remove ~/.bashrc", err)
	}

	fmt.Println("[+] Removed ~/.bashrc")
}

func askToReboot() bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Reboot the system now? Type yes or no: ")

		answer, err := reader.ReadString('\n')
		if err != nil && len(answer) == 0 {
			fmt.Fprintln(os.Stderr, "\n[-] Could not read your response; skipping reboot.")
			return false
		}

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "yes", "y":
			return true
		case "no", "n":
			return false
		default:
			fmt.Println("Please enter yes or no.")
		}
	}
}

func enableAndStartLy() {
	status("Enabling Ly for the next boot")

	if err := run("", "sudo", "systemctl", "enable", "ly@tty1.service"); err != nil {
		fail("Could not enable Ly", err)
	}
}

func rebootIfSelected() {
	if !askToReboot() {
		status("Reboot skipped")
		return
	}

	status("Rebooting")
	if err := run("", "sudo", "systemctl", "reboot"); err != nil {
		fail("Could not reboot the system", err)
	}
}

func main() {
	if _, err := os.Stat("/etc/arch-release"); err != nil {
		fmt.Fprintln(os.Stderr, "[-] This program is intended for Arch Linux.")
		os.Exit(1)
	}

	if os.Geteuid() == 0 {
		fmt.Fprintln(os.Stderr, "[-] Run this program as a regular user, not as root.")
		os.Exit(1)
	}

	packages, err := readPackages()
	if err != nil {
		fail("Could not read embedded package list", err)
	}

	installYay()
	installPackages(packages)
	installTemplates()
	installLazyVim()
	installZshConfig()
	setZshAsLoginShell()
	installLyConfig()
	installJournalLimit()
	enableAndStartLy()
	rebootIfSelected()

	status("Setup complete")
}

