package cmd

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidbudnick/redis-tui/internal/secret"
	"github.com/davidbudnick/redis-tui/internal/types"
)

// CheckSecretStore verifies if the OS keychain is available.
func (c *Commands) CheckSecretStore() tea.Cmd {
	return func() tea.Msg {
		if c.store == nil {
			return types.SecretStoreUnavailableMsg{}
		}

		if cs, ok := c.store.(*secret.ChainStore); ok {
			if cs.IsAvailable() {
				return types.SecretStoreAvailableMsg{}
			}
		}

		return types.RequireMasterPasswordMsg{}
	}
}

// InitializeFileStore attempts to unlock or create the file-based secret store with the provided master password.
func (c *Commands) InitializeFileStore(configDir string, masterPassword string) tea.Cmd {
	return func() tea.Msg {
		if c.store == nil {
			return types.SecretStoreUnavailableMsg{}
		}
		if cs, ok := c.store.(*secret.ChainStore); ok {
			vaultPath := filepath.Join(configDir, "vault.enc")

			fs, err := secret.NewFileStore(vaultPath, []byte(masterPassword))
			if err != nil {
				return types.MasterPasswordErrorMsg{Err: err}
			}
			cs.AddProvider(fs)
		}
		return types.SecretStoreAvailableMsg{}
	}
}
