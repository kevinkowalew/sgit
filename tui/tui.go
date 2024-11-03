package tui

import (
	"github.com/manifoldco/promptui"
	"sgit/internal/cmd"
)

type TUI struct {
}

func New() *TUI {
	return &TUI{}
}

func (t TUI) Handle(repository cmd.GithubRepository, state cmd.RepositoryState) error {
	if state == cmd.UpToDate || state == cmd.HasUncommittedChanges {
		return nil
	}
	return nil
}

func (t TUI) HandleUpToDateWithLocalChanges() error {
	return nil
}

func showPrompt(title string, items []string) (string, error) {
	p := promptui.Select{
		Label: title,
		Items: items,
	}
	_, a, err := p.Run()
	return a, err
}
