package main

import (
	"automappath/maps"
	"log"
)

type coordinate struct {
	x, y int
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	maps.GetPath()
}
