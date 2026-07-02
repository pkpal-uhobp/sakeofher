package main

import (
	"context"
	"log"

	"sakeofher/internal/app"
)

func main() {
	if err := app.RunWorker(context.Background()); err != nil {
		log.Fatal(err)
	}
}
