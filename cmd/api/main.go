package main

import (
	"context"
	"log"

	"sakeofher/internal/app"
)

func main() {
	if err := app.RunAPI(context.Background()); err != nil {
		log.Fatal(err)
	}
}
