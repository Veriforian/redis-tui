package ui

import (
	"strings"
)

func (m Model) viewMasterPassword() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Secure Vault Locked"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("OS Keyring unavailable. Falling back to encrypted file store."))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Enter your master password to unlock or create the vault."))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Password:"))
	b.WriteString("\n")
	b.WriteString(m.MasterPasswordInput.View())
	b.WriteString("\n\n")

	if m.ConnectionError != "" { // Reusing this field for the auth error display
		b.WriteString(errorStyle.Render(m.ConnectionError))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("enter:unlock  esc:quit"))

	return m.renderModal(b.String())
}
