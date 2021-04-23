// Package app implements github golang repo search app.
package app

import "github/aitigro/server"

// App represent app with http and cmd handles for github go lib search queries.
type App struct {
	done   <-chan bool
	server server.Server
	client server.Client
}

// NewApp returns app.
func NewApp() *App {
	return &App{
		done: make(chan bool),
	}
}

// Start starts app.
func (a *App) Start() <-chan bool {
	a.server.Start()
	a.client.RunQuery()

	return a.done
}
