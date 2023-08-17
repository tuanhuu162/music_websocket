package main

import (
	"github.com/tuanhuu162/music_websocket/server"
)

func main() {
	app := server.NewApp()
	app.Listen(":8080")
}
