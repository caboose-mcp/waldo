package auth

import (
	"fmt"
)

// HankoClient wraps Hanko authentication for waldo registry.
// Stubs for future implementation with github.com/teamhanko/hanko-sdk-go
type HankoClient struct {
	BaseURL string
	// TODO: Add hanko-sdk-go client
	// client *hanko.Client
}

// NewHankoClient creates a new Hanko authentication client.
func NewHankoClient(baseURL string) *HankoClient {
	return &HankoClient{
		BaseURL: baseURL,
	}
}

// Login initiates passwordless authentication via Hanko.
// Opens browser to Hanko login page, returns JWT token on success.
// TODO: Implement with hanko-sdk-go
func (h *HankoClient) Login() (string, error) {
	fmt.Println("TODO: Implement Hanko passwordless login")
	fmt.Println("  Expected: Open browser, user authenticates with passkey/biometric")
	fmt.Println("  Returns: JWT token for registry API calls")
	return "", fmt.Errorf("not implemented")
}

// ValidateToken checks if a JWT is valid and not expired.
// TODO: Implement JWT validation
func (h *HankoClient) ValidateToken(token string) error {
	fmt.Println("TODO: Implement JWT token validation")
	fmt.Println("  Expected: Verify signature, check expiration")
	return fmt.Errorf("not implemented")
}

// GetUser retrieves authenticated user info from token.
// TODO: Implement user info extraction
func (h *HankoClient) GetUser(token string) (map[string]interface{}, error) {
	fmt.Println("TODO: Implement GetUser")
	fmt.Println("  Expected: Extract user ID, email, name from JWT claims")
	return nil, fmt.Errorf("not implemented")
}

// Logout revokes a user session.
// TODO: Implement session revocation
func (h *HankoClient) Logout(token string) error {
	fmt.Println("TODO: Implement Logout")
	fmt.Println("  Expected: Revoke token on Hanko server")
	return fmt.Errorf("not implemented")
}

// RegistrationURL returns the Hanko signup URL.
func (h *HankoClient) RegistrationURL() string {
	return h.BaseURL + "/register"
}

// LoginURL returns the Hanko login URL.
func (h *HankoClient) LoginURL() string {
	return h.BaseURL + "/login"
}
