package tui

import (
	"fmt"
	"sgit/internal/interactor"
	"strings"
	"time"

	"github.com/fatih/color"
)

func Output(repos []interactor.Repo) {
	rainbow := []color.Attribute{
		color.FgBlue, color.FgMagenta, color.FgCyan,
	}

	langIndex := make(map[string]int, 0)
	for _, repo := range repos {
		z, ok := langIndex[repo.Language]
		if !ok {
			langIndex[repo.Language] = len(langIndex)
			z += 1
		}

		d := color.New(
			rainbow[z%(len(rainbow)-1)],
			color.Bold,
		)
		d.Print(repo.Language + " ")

		d = color.New(color.FgWhite)
		d.Println(
			fmt.Sprintf("%s/%s ", repo.Owner, repo.Name),
		)
	}
}

var last = 0.0

func PrintProgress(percentage float64) {
	if percentage == 1.00 && last == 0.0 {
		// makes sinle item progress updates smoother
		for _, p := range []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0} {
			printProgress(p)
			time.Sleep(90 * time.Millisecond)
		}
	} else {
		printProgress(percentage)
	}
}

func printProgress(percentage float64) {
	width := 50
	barWidth := int(percentage * float64(width))
	bar := strings.Repeat("#", barWidth) + strings.Repeat(" ", width-barWidth)
	fmt.Printf("\r[%s] %3.0f%%", bar, percentage*100)
	last = percentage

	if percentage == 1.0 {
		fmt.Printf("\r%s\r", strings.Repeat(" ", width+8))
	}
}
