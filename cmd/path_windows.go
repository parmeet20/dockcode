//go:build windows

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// EnsureGoBinInPath checks if the default Go binary path ($HOME\go\bin) exists
// in the user's PATH environment variable. If missing, it permanently writes it
// to the Windows Registry (HKCU\Environment) — which works for both CMD and
// PowerShell — and broadcasts a WM_SETTINGCHANGE so that any open Explorer
// windows and new CMD/PowerShell processes pick up the change immediately
// without a full logout/reboot.
func EnsureGoBinInPath() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	goBin := filepath.Join(home, "go", "bin")

	// ── 1. Check current process PATH first ───────────────────────────────────
	currentPath := os.Getenv("PATH")
	if isInPath(currentPath, goBin) {
		return nil // Already available to this process
	}

	// ── 2. Read the registry User PATH ────────────────────────────────────────
	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		"Environment",
		registry.QUERY_VALUE|registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("path setup: open registry: %w", err)
	}
	defer k.Close()

	regVal, valType, err := k.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("path setup: read registry Path: %w", err)
	}

	// Already in registry?
	if isInPath(regVal, goBin) {
		// Registry is correct but current process didn't have it — update process env
		_ = os.Setenv("PATH", currentPath+string(os.PathListSeparator)+goBin)
		return nil
	}

	// ── 3. Append to registry value ───────────────────────────────────────────
	newVal := regVal
	if newVal != "" && !strings.HasSuffix(newVal, ";") {
		newVal += ";"
	}
	newVal += goBin

	// Preserve REG_EXPAND_SZ vs REG_SZ so that %USERPROFILE% paths still expand
	if valType == registry.EXPAND_SZ || strings.Contains(newVal, "%") {
		err = k.SetExpandStringValue("Path", newVal)
	} else {
		err = k.SetStringValue("Path", newVal)
	}
	if err != nil {
		return fmt.Errorf("path setup: write registry: %w", err)
	}

	// ── 4. Update current process PATH so it takes effect immediately ─────────
	_ = os.Setenv("PATH", currentPath+string(os.PathListSeparator)+goBin)

	// ── 5. Broadcast WM_SETTINGCHANGE so open CMD/Explorer windows see it ─────
	broadcastEnvChange()

	return nil
}

// isInPath checks if target is already one of the semicolon-separated paths.
func isInPath(pathEnv, target string) bool {
	for _, p := range filepath.SplitList(pathEnv) {
		if strings.EqualFold(filepath.Clean(p), filepath.Clean(target)) {
			return true
		}
	}
	return false
}

// broadcastEnvChange sends WM_SETTINGCHANGE with lParam="Environment" so that
// Windows Explorer, open CMD windows, and new PowerShell/CMD sessions reload
// their environment without requiring a logout.
func broadcastEnvChange() {
	const (
		hwndBroadcast   = uintptr(0xFFFF)
		wmSettingChange = uintptr(0x001A)
		smtoAbortIfHung = uintptr(0x0002)
		timeoutMs       = 5000
	)
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessageTimeout := user32.NewProc("SendMessageTimeoutW")
	envStr, _ := syscall.UTF16PtrFromString("Environment")
	var result uintptr
	_, _, _ = sendMessageTimeout.Call(
		hwndBroadcast,
		wmSettingChange,
		0,
		uintptr(unsafe.Pointer(envStr)),
		smtoAbortIfHung,
		timeoutMs,
		uintptr(unsafe.Pointer(&result)),
	)
}
