// go-wfs project main.go
package main

import (
	"fmt"

	"github.com/jivanamara/go-wfs/server"
)

func main() {
	fmt.Println("Hello World!")
	server.StartServer(":9000")
}
