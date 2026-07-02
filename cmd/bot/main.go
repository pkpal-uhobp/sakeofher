package main

import (
	"context"
	"log"

	"sakeofher/internal/app"
)

func main() {
	if err := app.RunBot(context.Background()); err != nil {
		log.Fatal(err)
	}
}
