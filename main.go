package main

import (
	"go-tello/app/controllers"
	"log"
)

func main() {
	err := controllers.StartWebServer()
	log.Fatal(err)
}
