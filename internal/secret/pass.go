package secret

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

type PassStore struct{}

func NewPassStore() *PassStore {
	return &PassStore{}
}

func (s *PassStore) Name() string {
	return "Pass_GPG"
}

// buildPath standardizes the storage path inside the pass store.
func (s *PassStore) buildPath(service, user string) string {
	return fmt.Sprintf("redis-tui:%s:%s", service, user)
}

// isAvailable checks if the pass binary is installed and in the path.
func (s *PassStore) isAvailable() error {
	if _, err := exec.LookPath("pass"); err != nil {
		return ErrUnavailable
	}
	return nil
}

func (s *PassStore) Get(service, user string) ([]byte, error) {
	if err := s.isAvailable(); err != nil {
		return nil, err
	}

	path := s.buildPath(service, user)
	cmd := exec.Command("pass", "show", path)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to execute pass show: %w", err)
	}

	// pass appends a newline, strip it off.
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}

	return out, nil
}

func (s *PassStore) Set(service, user string, pwd []byte) error {
	if err := s.isAvailable(); err != nil {
		return err
	}

	path := s.buildPath(service, user)
	cmd := exec.Command("pass", "insert", "-f", path)
	cmd.Stdin = bytes.NewReader(pwd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute pass insert: %w", err)
	}

	return nil
}

func (s *PassStore) Delete(service, user string) error {
	if err := s.isAvailable(); err != nil {
		return err
	}

	path := s.buildPath(service, user)
	cmd := exec.Command("pass", "rm", "-f", path)
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to execute pass rm: %w", err)
	}

	return nil
}
