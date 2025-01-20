package main

import (
	"fmt"
	"go-website-framework/auth"
	"go-website-framework/server"
)

func main() {
	s := server.NewServer()

	url := "https://dev.connorisseur.com"
	port := ":8080"

	s.Auth = auth.NewAuth(url)
	s.AuthRoutes(port)

	staticFileMapping := map[string]string{
		"/index":     "index.html",
		"/websocket": "static/websocket.html",
		"/page":      "static/page.html",
	}

	s.StaticFiles(staticFileMapping)
	s.StaticDirectory("static")

	s.OnReceiveWebsocket(func(session_uuid string, msg string) error {
		onWSReceive(session_uuid, msg)
		fmt.Printf("%v", s.Clients)
		return nil
	})

	fmt.Println("Starting WebSocket server on port", port)
	s.Serve(port)
}

func onWSReceive(session_uuid, message string) {
	fmt.Printf("Message received: %s from %s\n", message, session_uuid)
}
