package testutil

// MockSecureStoreClient implements the complete service.ConfigService interface for testing.
type MockSecureStoreClient struct {
	// Configurable return values
	SaveError  error
	LoadError  error
	CloseError error
}

// NewMockSecureStoreClient creates a new fully-mocked Config client.
func NewMockSecureStoreClient() *MockSecureStoreClient {
	return &MockSecureStoreClient{}
}

func (m *MockSecureStoreClient) Save(_ string, _ string) error { return m.SaveError }
func (m *MockSecureStoreClient) Load(_ string) (string, error) { return "", m.LoadError }
func (m *MockSecureStoreClient) Close() error                  { return m.CloseError }
