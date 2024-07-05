package main

import (
	"github.com/soltanoff/go_github_release_monitor_bot/internal"
)

func main() {
	if err := internal.Run(); err != nil {
		panic(err)
	}
}
