package main

import "github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/app"

func main() {
	if err := app.Run(":8081"); err != nil {
		panic(err)
	}
}