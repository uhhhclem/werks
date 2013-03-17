package main

import (
	"gamework"
)

func main() {
	g := gamework.InitTestGameWithTestEngine()
	gamework.PlayToConsole(g)
}
