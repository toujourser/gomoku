package main

import (
	_ "github.com/toujourser/gomoku/config"
	"github.com/toujourser/gomoku/internal/app"
	_ "github.com/toujourser/gomoku/pkg/redis"
)

func main() {
	app.InitServer()
}
