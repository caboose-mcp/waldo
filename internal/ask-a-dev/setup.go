package askadev

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/term"
)

// SetupConfig holds all configuration for ask-a-dev service
type SetupConfig struct {
	DiscordBotToken    string `json:"discord_bot_token"`
	ClaudeAPIKey       string `json:"anthropic_api_key"`
	GitHubPAT          string `json:"github_pat"`
	N8NInstanceURL     string `json:"n8n_instance_url"`
	N8NAPIKey          string `json:"n8n_api_key"`
	QuestionChannelID  string `json:"question_channel_id,omitempty"`
	AdminChannelID     string `json:"admin_channel_id,omitempty"`
	QuestionTime       string `json:"question_time"`
	SummaryTime        string `json:"summary_time"`
	Timezone           string `json:"timezone"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

// NewSetupConfig creates an empty config
func NewSetupConfig() *SetupConfig {
	return &SetupConfig{
		QuestionTime: "09:00",
		SummaryTime:  "22:00",
		Timezone:     "UTC",
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
}

// PromptSecure reads input without echoing (for passwords/tokens)
func PromptSecure(prompt string) (string, error) {
	fmt.Print(prompt)

	// Read from stdin without echo
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback for non-terminal (testing)
		var input string
		_, err := fmt.Scanln(&input)
		return input, err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	input, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	fmt.Println() // newline after password input
	return string(input), nil
}

// PromptRegular reads input with echo
func PromptRegular(prompt string, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	var input string
	_, err := fmt.Scanln(&input)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	if input == "" && defaultValue != "" {
		return defaultValue, nil
	}

	return input, nil
}

// ValidateDiscordToken checks Discord bot token format
func ValidateDiscordToken(token string) error {
	// Discord bot tokens are base64-like, typically 60-68 chars
	re := regexp.MustCompile(`^[A-Za-z0-9_-]{60,68}$`)
	if !re.MatchString(token) {
		return fmt.Errorf("invalid Discord bot token format (expected 60-68 alphanumeric characters)")
	}
	return nil
}

// ValidateClaudeKey checks Anthropic API key format
func ValidateClaudeKey(key string) error {
	// Anthropic keys start with sk-ant-
	if !strings.HasPrefix(key, "sk-ant-") {
		return fmt.Errorf("invalid Claude API key format (must start with 'sk-ant-')")
	}
	if len(key) < 20 {
		return fmt.Errorf("Claude API key too short")
	}
	return nil
}

// ValidateGitHubPAT checks GitHub PAT format
func ValidateGitHubPAT(pat string) error {
	// Fine-grained: ghp_*
	// Legacy classic: ghp_* (also 36+ chars)
	// New format: github_pat_*
	validFineGrained := regexp.MustCompile(`^ghp_[A-Za-z0-9_]{36,}$`)
	validNewFormat := regexp.MustCompile(`^github_pat_[A-Za-z0-9_]{82}$`)

	if !validFineGrained.MatchString(pat) && !validNewFormat.MatchString(pat) {
		return fmt.Errorf("invalid GitHub PAT format (must start with 'ghp_' or 'github_pat_')")
	}
	return nil
}

// ValidateTime checks time format (HH:MM)
func ValidateTime(timeStr string) error {
	re := regexp.MustCompile(`^\d{2}:\d{2}$`)
	if !re.MatchString(timeStr) {
		return fmt.Errorf("invalid time format (expected HH:MM)")
	}

	parts := strings.Split(timeStr, ":")
	hour := parts[0]
	minute := parts[1]

	if hour > "23" || minute > "59" {
		return fmt.Errorf("invalid time (hours must be 00-23, minutes must be 00-59)")
	}
	return nil
}

// ValidateTimezone checks if timezone is reasonable
func ValidateTimezone(tz string) error {
	if tz == "" {
		return fmt.Errorf("timezone cannot be empty")
	}
	// Basic check: no special chars, looks like IANA name or UTC
	if len(tz) > 50 {
		return fmt.Errorf("timezone name too long")
	}
	// Reject obviously bad values
	if strings.Contains(tz, ";") || strings.Contains(tz, "|") || strings.Contains(tz, "&&") {
		return fmt.Errorf("timezone contains invalid characters")
	}
	return nil
}

// SaveSecurely writes config to ~/.config/waldo/ask-a-dev/credentials.json with 0600 perms
func (cfg *SetupConfig) SaveSecurely() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "waldo", "ask-a-dev")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	credFile := filepath.Join(configDir, "credentials.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(credFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	// Verify permissions were set correctly
	fi, err := os.Stat(credFile)
	if err != nil {
		return fmt.Errorf("failed to stat credentials file: %w", err)
	}

	if perm := fi.Mode().Perm(); perm != 0600 {
		return fmt.Errorf("credentials file has wrong permissions: %o (expected 0600)", perm)
	}

	return nil
}

// LoadSecurely reads config from ~/.config/waldo/ask-a-dev/credentials.json
func LoadSecurely() (*SetupConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	credFile := filepath.Join(home, ".config", "waldo", "ask-a-dev", "credentials.json")

	// Check file permissions
	fi, err := os.Stat(credFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Not set up yet
		}
		return nil, fmt.Errorf("failed to stat credentials file: %w", err)
	}

	if perm := fi.Mode().Perm(); perm != 0600 {
		return nil, fmt.Errorf("credentials file has wrong permissions: %o (expected 0600)", perm)
	}

	data, err := os.ReadFile(credFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var cfg SetupConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &cfg, nil
}

// GetCredentialFile returns the path to credentials.json
func GetCredentialFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "waldo", "ask-a-dev", "credentials.json"), nil
}

// PrintPreview displays config without printing sensitive values
func (cfg *SetupConfig) PrintPreview() {
	maskToken := func(token string) string {
		if len(token) < 8 {
			return "••••••••"
		}
		return token[:4] + "•" + token[len(token)-3:]
	}

	fmt.Println("\n────────────────────────────────────────")
	fmt.Println("  Configuration Preview")
	fmt.Println("────────────────────────────────────────")
	fmt.Printf("  Discord Bot Token:     %s\n", maskToken(cfg.DiscordBotToken))
	fmt.Printf("  Claude API Key:        %s\n", maskToken(cfg.ClaudeAPIKey))
	fmt.Printf("  GitHub PAT:            %s\n", maskToken(cfg.GitHubPAT))
	if cfg.N8NInstanceURL != "" {
		fmt.Printf("  n8n Instance:          %s\n", cfg.N8NInstanceURL)
	}
	if cfg.N8NAPIKey != "" {
		fmt.Printf("  n8n API Key:           %s\n", maskToken(cfg.N8NAPIKey))
	}
	if cfg.QuestionChannelID != "" {
		fmt.Printf("  Question Channel ID:   %s\n", cfg.QuestionChannelID)
	}
	if cfg.AdminChannelID != "" {
		fmt.Printf("  Admin Channel ID:      %s\n", cfg.AdminChannelID)
	}
	fmt.Printf("  Question Time:         %s %s\n", cfg.QuestionTime, cfg.Timezone)
	fmt.Printf("  Summary Time:          %s %s\n", cfg.SummaryTime, cfg.Timezone)
	fmt.Println("────────────────────────────────────────")
}

// ZeroMemory overwrites sensitive strings with zeros (best-effort)
func ZeroMemory(s *string) {
	if s == nil {
		return
	}
	// Convert to byte array and overwrite
	b := []byte(*s)
	for i := range b {
		b[i] = 0
	}
	*s = ""
}
