package main

import (
	"log"
	"os"
)

func main() {
	log.Println(os.Args)
	// use os.Args[0] to determine binary  path, then call it everytime a new request hit the route
}
