package app

type App struct {
	Port       int
	Identifier string
}

func NewApp(port int, hostname string) *App {
	return &App{Port: port, Identifier: hostname}
}
