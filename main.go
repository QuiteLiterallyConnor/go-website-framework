package main

import (
	"fmt"
	"go-website-framework/auth"
	"go-website-framework/server"
)

func main() {
	s := server.NewServer()

	s.Router.Static("/static", "./static")
	s.Router.LoadHTMLGlob("templates/*")
	// s.Router.LoadHTMLGlob("static/*")


	url := "https://dev.connorisseur.com"
	s.Auth = auth.NewAuth(url)
	s.AuthRoutes("")

	s.OnReceiveWebsocket(func(sessionUUID string, msg string) error {
		onWSReceive(sessionUUID, msg)
		return nil
	})

	port := ":8090"
	fmt.Println("Starting server on port", port)
	s.Serve(port)
}

func onWSReceive(sessionUUID, message string) {
	fmt.Printf("Message received: %s from %s\n", message, sessionUUID)
}
