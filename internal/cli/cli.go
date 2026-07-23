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
	fmt.Println("  setup            Configure Jurnal API credentials and default branch")
	fmt.Println("  setup --global   Configure global Git user credentials (user.name & user.email)")
	fmt.Println("  init             Initialize Git repo, default branch, and starter files")
	fmt.Println("  stage            Plan and commit changes step-by-step")
	fmt.Println("  -v               Show version credits")
}

// HandleSetup configures credentials and settings interactively.
func HandleSetup(args []string) {
	reader := bufio.NewReader(os.Stdin)

	isGlobal := false
	for _, arg := range args {
		if arg == "--global" || arg == "--git" {
			isGlobal = true
			break
		}
	}

	if isGlobal {
		fmt.Println("\n--- Configuring Global Git User Credentials ---")
		fmt.Print("Git Global User Name: ")
		gitName, _ := reader.ReadString('\n')
		gitName = strings.TrimSpace(gitName)

		fmt.Print("Git Global User Email: ")
		gitEmail, _ := reader.ReadString('\n')
		gitEmail = strings.TrimSpace(gitEmail)

		if gitName != "" && gitEmail != "" {
			if git.ConfigureUser(gitName, gitEmail) {
				fmt.Println("✔ Git global user credentials set successfully")
			} else {
				fmt.Println("✖ Failed to set Git global configuration")
			}
		} else {
			fmt.Println("⚠ Git configuration skipped (name or email empty)")
		}
		return
	}

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

	fmt.Print("Default Git Branch (default: main): ")
	defaultBranch, _ := reader.ReadString('\n')
	defaultBranch = strings.TrimSpace(defaultBranch)
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	err := config.Save(url, key, model, defaultBranch)
	if err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✔ Credentials saved to ~/.jurnal.json")
	fmt.Println("Tip: Run 'jurnal setup --global' if you also want to configure global Git user credentials.")
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
		fmt.Println("Branch Action:")
		fmt.Printf("  [1] Switch to proposed branch ('%s') [Default]\n", plan.ProposedBranch)
		fmt.Printf("  [2] Stay on current branch ('%s')\n", currentBranch)
		fmt.Print("Choice [1/2]: ")

		ans, _ := reader.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))
		if ans == "" || ans == "1" || ans == "y" || ans == "yes" {
			if git.Checkout(plan.ProposedBranch) {
				fmt.Printf("✔ Switched to branch '%s'\n", plan.ProposedBranch)
			} else {
				fmt.Printf("✖ Failed to switch branch. Continuing on current branch '%s'.\n", currentBranch)
			}
		} else {
			fmt.Printf("ℹ Continuing commits on current branch '%s'\n", currentBranch)
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

// HandleInit initializes a git repository, sets default branch, and generates starter files.
func HandleInit(args []string) {
	reader := bufio.NewReader(os.Stdin)

	cfg, _ := config.Load()
	configuredBranch := cfg.DefaultBranch
	if configuredBranch == "" {
		configuredBranch = "main"
	}

	defaultBranch := configuredBranch
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		defaultBranch = strings.TrimSpace(args[0])
	} else {
		fmt.Printf("Default branch name [%s]: ", configuredBranch)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			defaultBranch = input
		}
	}

	if !git.IsRepository() {
		fmt.Printf("Initializing new Git repository on branch '%s'... ", defaultBranch)
		if git.Init(defaultBranch) {
			fmt.Println("Done.\n✔ Git repository initialized successfully.")
		} else {
			fmt.Println("\n✖ Failed to initialize Git repository.")
			os.Exit(1)
		}
	} else {
		currentBranch := git.GetCurrentBranch()
		if currentBranch != defaultBranch && !git.HasCommits() {
			fmt.Printf("Renaming branch '%s' -> '%s'... ", currentBranch, defaultBranch)
			if git.RenameBranch(defaultBranch) {
				fmt.Println("Done.")
			} else {
				fmt.Println("\n⚠ Failed to rename branch.")
			}
		}
		fmt.Printf("✔ Active Git repository detected on branch '%s'.\n", git.GetCurrentBranch())
	}

	// Helper function to check file existence
	fileExists := func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	}

	// Check and create .gitignore
	if !fileExists(".gitignore") {
		fmt.Print("Create default .gitignore? [Y/n]: ")
		ans, _ := reader.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))
		if ans == "" || ans == "y" || ans == "yes" {
			gitignoreContent := `# Binaries & Executables
*.exe
*.dll
*.so
*.dylib
bin/
dist/

# OS specific files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# IDEs & Editors
.idea/
.vscode/
*.swp
*.swo
*~

# Logs & Environment
*.log
.env
.env.local
`
			err := os.WriteFile(".gitignore", []byte(gitignoreContent), 0644)
			if err == nil {
				fmt.Println("✔ Created .gitignore")
			} else {
				fmt.Printf("✖ Failed to create .gitignore: %v\n", err)
			}
		}
	}

	// Check and create README.md
	if !fileExists("README.md") {
		fmt.Print("Create starter README.md? [Y/n]: ")
		ans, _ := reader.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))
		if ans == "" || ans == "y" || ans == "yes" {
			dirName := "Project"
			if cwd, err := os.Getwd(); err == nil {
				parts := strings.Split(cwd, string(os.PathSeparator))
				if len(parts) > 0 && parts[len(parts)-1] != "" {
					dirName = parts[len(parts)-1]
				}
			}
			readmeContent := fmt.Sprintf("# %s\n\nInitial repository setup with Jurnal CLI.\n", dirName)
			err := os.WriteFile("README.md", []byte(readmeContent), 0644)
			if err == nil {
				fmt.Println("✔ Created README.md")
			} else {
				fmt.Printf("✖ Failed to create README.md: %v\n", err)
			}
		}
	}

	// Option to run stage immediately
	fmt.Print("\nRun 'jurnal stage' now to plan initial commits? [Y/n]: ")
	ans, _ := reader.ReadString('\n')
	ans = strings.ToLower(strings.TrimSpace(ans))
	if ans == "" || ans == "y" || ans == "yes" {
		fmt.Println()
		HandleStage()
	} else {
		fmt.Println("\n✔ Initialization complete! Run 'jurnal stage' whenever you're ready to commit.")
	}
}

