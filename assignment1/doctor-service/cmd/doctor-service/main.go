package main

import "github.com/CoffeeSi/golang2AITU/assignment1/doctor-service/internal/app"

func main() {
	// app.Run(":8080")
	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}