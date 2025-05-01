package main

import (
	"context"
)

func main() {
	ctx := context.Background()

	geminiClient, err := newClient(ctx)
	if err != nil {
		panic(err)
	}

	gui, err := newGui(ctx, geminiClient)
	if err != nil {
		panic(err)
	}

	gui.Show(ctx)

}
