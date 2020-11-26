package app

import "GoChallenge/server"

type App struct {
	done   <-chan bool
	server server.Server
	client server.Client
}

func NewApp() *App {
	return &App{
		done: make(chan bool),
	}
}

func (a *App) Start() <-chan bool {
	a.server.Start()
	a.client.RunQuery()

	return a.done
}
