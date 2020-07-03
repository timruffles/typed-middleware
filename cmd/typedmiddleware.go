package main

import (
	"log"
	"os"

	"github.plaid.com/plaid/typedmiddleware/generator"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	if len(os.Args) < 2 {
		log.Fatal("Supply the middleware stack type as the first argument")
		return
	}

	target := os.Args[1]
	err = generator.Run(wd, os.Getenv("GOFILE"), target)

	if err != nil {
		log.Fatal(err)
		return
	}
}

