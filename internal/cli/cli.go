package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/luckywirasakti/jurnal/internal/config"
	"github.com/luckywirasakti/jurnal/internal/git"
	"github.com/luckywirasakti/jurnal/internal/planner"
)

// PrintCredits displays developer attribution.
func PrintCredits(version string) {
	fmt.Println("==================================================")
	fmt.Printf("Jurnal CLI - Version %s (Go Native)\n\n", version)
	fmt.Println("Author:   Lucky Wirasakti")
	fmt.Println("Email:    lucky.wirasakti@gmail.com")
	fmt.Println("GitHub:   https://github.com/luckywirasakti")
	fmt.Println("Website:  https://luckywirasakti.web.id/")
	fmt.Println("==================================================")
}

// PrintUsage prints the CLI manual.
func PrintUsage() {
	fmt.Println("AI-powered git staging and commit workflow assistant")
	fmt.Println("\nUsage:")
	fmt.Println("  jurnal <command>")
	fmt.Println("\nCommands:")
	fmt.Println("  setup    Configure Jurnal API and Git settings")
	fmt.Println("  stage    Plan and commit changes step-by-step")
	fmt.Println("  -v       Show version credits")
}

// HandleSetup configures credentials and settings interactively.
func HandleSetup() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n--- Configuring Jurnal Local Settings ---")

	fmt.Print("OpenAI-compatible Base URL (e.g. https://api.openai.com/v1): ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)

	fmt.Print("API Key: ")
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)

	fmt.Print("Default Model Name (default: gpt-4o-mini): ")
	model, _ := reader.ReadString('\n')
	model = strings.TrimSpace(model)
	if model == "" {
		model = "gpt-4o-mini"
	}

	fmt.Print("Git Global User Name: ")
	gitName, _ := reader.ReadString('\n')
	gitName = strings.TrimSpace(gitName)

	fmt.Print("Git Global User Email: ")
	gitEmail, _ := reader.ReadString('\n')
	gitEmail = strings.TrimSpace(gitEmail)

	err := config.Save(url, key, model)
	if err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✔ Credentials saved to ~/.jurnal.json")

	if gitName != "" && gitEmail != "" {
		if git.ConfigureUser(gitName, gitEmail) {
			fmt.Println("✔ Git global configuration set successfully")
		} else {
			fmt.Println("✖ Failed to set Git global configuration")
		}
	} else {
		fmt.Println("⚠ Git configurations skipped (name/email empty)")
	}
}

// HandleStage coordinates branch generation and logical batch committing.
func HandleStage() {
	if !git.IsRepository() {
		fmt.Println("✖ Not a git repository")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil || cfg.OpenAIBaseURL == "" {
		fmt.Println("✖ Jurnal has not been configured yet.")
		fmt.Println("Run 'jurnal setup' first to configure credentials.")
		os.Exit(1)
	}

	fmt.Print("Analyzing local changes... ")
	status := git.GetStatus()
	diff := git.GetDiff()

	if strings.Contains(status, "nothing to commit") && strings.TrimSpace(diff) == "" {
		fmt.Println("\n✔ No changes detected. Workspace is clean.")
		return
	}
	fmt.Println("Done.")

	fmt.Print("Planning commits with AI... ")
	plan, err := planner.BuildStagingPlan(status, diff)
	if err != nil {
		fmt.Printf("\n✖ AI failed to generate commit plan: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done.\n")

	fmt.Println("==================================================")
	fmt.Printf("Proposed Branch Name: %s\n", plan.ProposedBranch)
	fmt.Printf("Current Branch:       %s\n", git.GetCurrentBranch())
	fmt.Println("==================================================")

	fmt.Println("\nProposed Commits:")
	for i, c := range plan.Commits {
		fmt.Printf("Batch %d:\n", i+1)
		fmt.Printf("  Commit Message: %s\n", c.Message)
		fmt.Println("  Files:")
		for _, f := range c.Files {
			fmt.Printf("    - %s\n", f)
		}
		fmt.Println()
	}

	reader := bufio.NewReader(os.Stdin)

	// Branch switch confirmation
	currentBranch := git.GetCurrentBranch()
	if plan.ProposedBranch != "" && plan.ProposedBranch != currentBranch {
		fmt.Printf("Create and switch to new branch '%s'? [Y/n]: ", plan.ProposedBranch)
		ans, _ := reader.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))
		if ans == "" || ans == "y" || ans == "yes" {
			if git.Checkout(plan.ProposedBranch) {
				fmt.Printf("✔ Switched to branch '%s'\n", plan.ProposedBranch)
			} else {
				fmt.Println("✖ Failed to switch branch. Continuing on current branch.")
			}
		}
	}

	// Staging batches execution
	fmt.Println("\nStarting Staging and Commit Batches...")
	for i, c := range plan.Commits {
		fmt.Printf("\nBatch %d/%d:\n", i+1, len(plan.Commits))
		fmt.Printf("  Files:   %s\n", strings.Join(c.Files, ", "))
		fmt.Printf("  Message: %s\n", c.Message)

		fmt.Print("Apply this commit batch? [Y/n]: ")
		ans, _ := reader.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))

		if ans == "" || ans == "y" || ans == "yes" {
			fmt.Print("Staging files... ")
			if git.Stage(c.Files) {
				fmt.Print("Committing... ")
				if git.Commit(c.Message) {
					fmt.Println("✔ Batch committed successfully")
				} else {
					fmt.Println("✖ Commit failed")
				}
			} else {
				fmt.Println("✖ Failed to stage files")
			}
		} else {
			fmt.Println("⚠ Batch skipped")
		}
	}
	fmt.Println("\n✔ Done! All actions processed.")
}
