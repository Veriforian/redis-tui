// Package secstore provides a secure storage for sensitive data.
// It uses the keyring package as the initial attempt to store sensitive credentials.
// It falls back to a user provided master key file if keyring is not available.
// The final fallback uses a local key file.
package secstore
