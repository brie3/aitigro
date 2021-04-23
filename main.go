package main

import (
	"github/aitigro/app"
	"log"
)

func main() {
	log.Println("Hello. I'm here")

	a := app.NewApp()

	<-a.Start()
}
