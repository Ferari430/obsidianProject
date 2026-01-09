package main

import "github.com/Ferari430/obsidianProject/cmd/app"

func main() {
	application := app.NewApp()
	_ = application
	application.Start()
	select {}
}
