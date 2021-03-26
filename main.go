package main

import (
	"github_seeker/app"
	"log"
)

func main() {
	log.Println("Hello. I'm here")

	a := app.NewApp()

	<-a.Start()
}
