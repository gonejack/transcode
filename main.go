package main

import (
	"log"
)

func main() {
	if e := new(trans).run(); e != nil {
		log.Fatal(e)
	}
}
