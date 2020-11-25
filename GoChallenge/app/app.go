package app

type App struct {
	done   <-chan bool
	server Server
	client Client
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
