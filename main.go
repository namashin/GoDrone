package main

import (
	"go-tello/app/controllers"
	"log"
)

func main() {
	log.Fatal(controllers.StartWebServer())
}
