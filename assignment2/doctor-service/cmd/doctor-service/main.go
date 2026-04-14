package main

import "github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/app"

func main() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
