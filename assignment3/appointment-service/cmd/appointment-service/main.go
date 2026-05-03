package main

import "github.com/CoffeeSi/golang2AITU/assignment3/appointment-service/internal/app"

func main() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
