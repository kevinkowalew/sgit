package tui

import (
	"fmt"
	"sgit/internal/cmd"

	"github.com/fatih/color"
	"time"
)

type TUI struct {
	done chan bool
}

func New() *TUI {
	return &TUI{
		done: make(chan bool),
	}
}

func (t *TUI) ShowSpinner() {
	go func() {
		t.spinner()
	}()
}

func (t *TUI) spinner() {
	frames := []string{"|", "/", "-", "\\"}
	for {
		select {
		case <-t.done:
			return
		default:
			for _, frame := range frames {
				fmt.Printf("\r%s Synchronizing", frame)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (t TUI) clearLine() {
	fmt.Print("\033[H\033[2J")
}

func (t TUI) Handle(repoStateMap map[cmd.Repository]cmd.RepositoryState) error {
	type repoStatePair struct {
		repo  cmd.Repository
		state cmd.RepositoryState
	}
	t.done <- true
	t.clearLine()
	langToRepo := make(map[string][]repoStatePair, 0)

	for repo, state := range repoStateMap {
		e, _ := langToRepo[repo.Language]
		langToRepo[repo.Language] = append(e, repoStatePair{repo, state})
	}

	rainbow := []color.Attribute{
		color.FgBlue, color.FgMagenta, color.FgCyan,
	}

	i := 0
	for lang, pairs := range langToRepo {
		for _, pair := range pairs {
			d := color.New(
				rainbow[i%(len(rainbow)-1)],
				color.Bold,
			)
			d.Print(lang + " ")

			d = color.New(color.FgWhite)
			d.Print(pair.repo.Name + " ")

			switch pair.state {
			case cmd.UpToDate:
				d = color.New(color.FgGreen, color.Bold)
				d.Println(pair.state.String())
			case cmd.HasUncommittedChanges:
				d = color.New(color.FgYellow, color.Bold)
				d.Println(pair.state.String())
			default:
				d = color.New(color.FgRed, color.Bold)
				d.Println(pair.state.String())
			}

		}
		i += 1
	}

	return nil
}
