package main

import (
	"GoChallenge/app"
	"log"
)

func main() {
	log.Println("Hello. I'm here")

	a := app.NewApp()

	<-a.Start()
}
