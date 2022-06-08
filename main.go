package main

import (
	"log"
)

func main() {
	err := new(transcode).run()
	if err != nil {
		log.Fatal(err)
	}
}
