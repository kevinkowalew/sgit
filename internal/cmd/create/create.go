package create

import (
	"bufio"
	"fmt"
	"os"
	"sgit/internal/interactor"
	"sgit/internal/tui"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "create",
	Short: "create a repo",
	Long:  "create a repo",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	name := showNamePrompt()
	private := showPrivatePrompt()
	tui.PrintProgress(0.0)

	i := interactor.New()
	repo, err := i.CreateRepo(cmd.Context(), name, private)
	if err != nil {
		return fmt.Errorf("interactor.CreateRepo failed: %w", err)
	}

	exists, err := i.Exists(*repo)
	if err != nil {
		return err
	}

	if !exists {
		if err := i.Clone(*repo); err != nil {
			return fmt.Errorf("interactor.Clone failed: %w, err", err)
		}
	}

	tui.PrintProgress(1.0)
	return nil
}

func showNamePrompt() string {
	reader := bufio.NewReader(os.Stdin)

	for {
		d := color.New(color.FgGreen, color.Bold)
		d.Print("Name: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if len(input) > 0 {
			return input
		} else {
			fmt.Println("You must enter a non-empty repo name.")
		}
	}
}

func showPrivatePrompt() bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		d := color.New(color.FgGreen, color.Bold)
		d.Print("Private (y/n): ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" {
			return true
		} else if input == "n" {
			return false
		} else {
			fmt.Println("Invalid input. Please enter y or n.")
		}

	}
}
