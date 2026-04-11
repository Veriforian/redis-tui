package secret

import (
	"errors"
	"fmt"

	"github.com/davidbudnick/redis-tui/internal/service"
)

var (
	ErrNotFound    = errors.New("secret not found in store")
	ErrUnavailable = errors.New("secret store provider is unavailable")
)

// vault models our internal credential layout
type vault struct {
	Credentials map[string]map[string][]byte `json:"credentials"`
}

type ChainStore struct {
	providers []service.StoreService
}

// NewChainStore initializes a new ChainStore with the given prioritized secret stores.
func NewChainStore(providers ...service.StoreService) *ChainStore {
	return &ChainStore{providers: providers}
}

func (c *ChainStore) Name() string {
	return "Chain"
}

// AddProvider dynamically appends a new store to the fallback chain
func (c *ChainStore) AddProvider(p service.StoreService) {
	c.providers = append(c.providers, p)
}

// IsAvailable performs a dummy write/delete to verify if at least one provider is operational.
func (c *ChainStore) IsAvailable() bool {
	testSvs := "redis-tui-sys"
	testUsr := "availability-check"

	if err := c.Set(testSvs, testUsr, []byte("test")); err != nil {
		return false
	}
	_ = c.Delete(testSvs, testUsr)
	return true
}

// Get iterates through the providers. It returns the first successful found secret.
func (c *ChainStore) Get(service, user string) ([]byte, error) {
	for _, p := range c.providers {
		secret, err := p.Get(service, user)
		if err == nil {
			return secret, nil
		}

		if errors.Is(err, ErrNotFound) || errors.Is(err, ErrUnavailable) {
			continue
		}

		// Bubble up crit errors
		return nil, fmt.Errorf("%s provider failed: %w", p.Name(), err)
	}

	return nil, ErrNotFound
}

// Set iterates through the providers. It sets a secret on all available providers.
func (c *ChainStore) Set(service, user string, pwd []byte) error {
	var success bool
	for _, p := range c.providers {
		err := p.Set(service, user, pwd)
		if err == nil {
			success = true
			continue
		}

		if errors.Is(err, ErrUnavailable) {
			continue
		}

		// Bubble up crit errors
		return fmt.Errorf("%s provider failed to set: %w", p.Name(), err)
	}

	if !success {
		return fmt.Errorf("all providers failed: %w", ErrUnavailable)
	}

	return nil
}

// Delete iterates through the providers. It deletes a secret on all available providers.
func (c *ChainStore) Delete(service, user string) error {
	var success bool
	for _, p := range c.providers {
		err := p.Delete(service, user)
		if err == nil || errors.Is(err, ErrNotFound) {
			success = true
			continue
		}

		if errors.Is(err, ErrUnavailable) {
			continue
		}

		// Bubble up crit errors
		return fmt.Errorf("%s provider failed to delete: %w", p.Name(), err)
	}
	if !success {
		return fmt.Errorf("all providers failed: %w", ErrUnavailable)
	}
	return nil
}
