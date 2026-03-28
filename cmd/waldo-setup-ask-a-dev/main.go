package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/caboose-mcp/waldo/internal/ask-a-dev"
	"github.com/charmbracelet/lipgloss"
)

func main() {

	// Styling
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		MarginBottom(1)

	stepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Bold(true)

	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("2"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("1"))

	// Clear screen
	fmt.Print("\033[2J\033[H")

	// Header
	fmt.Println(headerStyle.Render("🎭 ask-a-dev Setup"))
	fmt.Println(stepStyle.Render("Configure your transparent Q&A service"))
	fmt.Println()

	cfg := askadev.NewSetupConfig()

	// Step 1: Discord Bot Token
	fmt.Println(stepStyle.Render("[1/5] Discord Bot Token"))
	for {
		token, err := askadev.PromptSecure("  Enter your Discord bot token (hidden): ")
		if err != nil {
			fmt.Println(errorStyle.Render("  ✗ Error reading input"))
			continue
		}

		if strings.TrimSpace(token) == "" {
			fmt.Println(errorStyle.Render("  ✗ Token cannot be empty"))
			continue
		}

		if err := askadev.ValidateDiscordToken(token); err != nil {
			fmt.Printf(errorStyle.Render("  ✗ %s\n"), err.Error())
			continue
		}

		cfg.DiscordBotToken = token
		fmt.Println(successStyle.Render("  ✓ Token validated"))
		break
	}
	fmt.Println()

	// Step 2: Claude API Key
	fmt.Println(stepStyle.Render("[2/5] Anthropic Claude API Key"))
	for {
		key, err := askadev.PromptSecure("  Enter your Claude API key (hidden): ")
		if err != nil {
			fmt.Println(errorStyle.Render("  ✗ Error reading input"))
			continue
		}

		if strings.TrimSpace(key) == "" {
			fmt.Println(errorStyle.Render("  ✗ Key cannot be empty"))
			continue
		}

		if err := askadev.ValidateClaudeKey(key); err != nil {
			fmt.Printf(errorStyle.Render("  ✗ %s\n"), err.Error())
			continue
		}

		cfg.ClaudeAPIKey = key
		fmt.Println(successStyle.Render("  ✓ Key validated"))
		break
	}
	fmt.Println()

	// Step 3: GitHub PAT
	fmt.Println(stepStyle.Render("[3/5] GitHub Fine-Grained PAT"))
	fmt.Println("  (Need help? See: https://github.com/settings/tokens)")
	for {
		pat, err := askadev.PromptSecure("  Enter your GitHub PAT (hidden): ")
		if err != nil {
			fmt.Println(errorStyle.Render("  ✗ Error reading input"))
			continue
		}

		if strings.TrimSpace(pat) == "" {
			fmt.Println(errorStyle.Render("  ✗ PAT cannot be empty"))
			continue
		}

		if err := askadev.ValidateGitHubPAT(pat); err != nil {
			fmt.Printf(errorStyle.Render("  ✗ %s\n"), err.Error())
			continue
		}

		cfg.GitHubPAT = pat
		fmt.Println(successStyle.Render("  ✓ PAT validated"))
		break
	}
	fmt.Println()

	// Step 4: Schedule Configuration
	fmt.Println(stepStyle.Render("[4/5] Configure Schedule (UTC)"))

	for {
		qTime, err := askadev.PromptRegular("  Question time (HH:MM)", cfg.QuestionTime)
		if err != nil {
			fmt.Println(errorStyle.Render("  ✗ Error reading input"))
			continue
		}

		if err := askadev.ValidateTime(qTime); err != nil {
			fmt.Printf(errorStyle.Render("  ✗ %s\n"), err.Error())
			continue
		}

		cfg.QuestionTime = qTime
		break
	}

	for {
		sTime, err := askadev.PromptRegular("  Summary time (HH:MM)", cfg.SummaryTime)
		if err != nil {
			fmt.Println(errorStyle.Render("  ✗ Error reading input"))
			continue
		}

		if err := askadev.ValidateTime(sTime); err != nil {
			fmt.Printf(errorStyle.Render("  ✗ %s\n"), err.Error())
			continue
		}

		cfg.SummaryTime = sTime
		break
	}

	for {
		tz, err := askadev.PromptRegular("  Timezone (IANA format)", cfg.Timezone)
		if err != nil {
			fmt.Println(errorStyle.Render("  ✗ Error reading input"))
			continue
		}

		if err := askadev.ValidateTimezone(tz); err != nil {
			fmt.Printf(errorStyle.Render("  ✗ %s\n"), err.Error())
			continue
		}

		cfg.Timezone = tz
		break
	}

	fmt.Println(successStyle.Render("  ✓ Schedule configured"))
	fmt.Println()

	// Step 5: n8n Configuration (optional)
	fmt.Println(stepStyle.Render("[5/5] n8n Cloud Configuration (optional)"))

	n8nInstance, err := askadev.PromptRegular("  n8n instance URL", "")
	if strings.TrimSpace(n8nInstance) != "" {
		cfg.N8NInstanceURL = n8nInstance

		n8nKey, err := askadev.PromptSecure("  n8n API key (hidden): ")
		if err == nil && strings.TrimSpace(n8nKey) != "" {
			cfg.N8NAPIKey = n8nKey
			fmt.Println(successStyle.Render("  ✓ n8n configured"))
		}
	} else {
		fmt.Println("  (Skipped — can configure later)")
	}
	fmt.Println()

	// Preview
	cfg.PrintPreview()
	fmt.Println()

	// Confirmation
	confirm, err := askadev.PromptRegular("  Save configuration securely?", "y")
	if err != nil || (strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes") {
		fmt.Println(errorStyle.Render("✗ Setup cancelled"))
		os.Exit(1)
	}

	// Save
	if err := cfg.SaveSecurely(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("✗ Failed to save: %v", err)))
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(successStyle.Render("✓ Configuration saved securely!"))
	fmt.Println()

	// Next steps
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("📖 Next Steps:"))
	fmt.Println("  1. Create n8n workflows (see SETUP.md for templates)")
	fmt.Println("  2. Configure n8n credentials with tokens from above")
	fmt.Println("  3. Activate workflows")
	fmt.Println("  4. Test with: waldo ask-a-dev test")
	fmt.Println()
	fmt.Println("  Docs: https://github.com/caboose-mcp/waldo/tree/main/ask-a-dev")
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Stored at: ~/.config/waldo/ask-a-dev/credentials.json"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Permissions: 0600 (readable only by you)"))
}
