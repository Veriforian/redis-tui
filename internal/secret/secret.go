// Package secret provides a chain of secret stores for storing and retrieving secrets.
// Secrets are encrypted and stored in an order of priorities:
//   - Keyring (macOS, Linux, Windows)
//   - Pass (macOS, Linux)
//   - File (macOS, Linux, Windows)
package secret
