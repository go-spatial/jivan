package server

import (
	"fmt"
	"net/http"

	"github.com/jivanamara/go-wfs/provider"
)

var P provider.Provider

func StartServer(bindAddress string, p provider.Provider) {
	fmt.Println("Listening on ", bindAddress)
	P = p
	setUpRoutes()
	err := http.ListenAndServe(bindAddress, nil)
	if err != nil {
		panic(fmt.Sprintf("Problem starting web server: %v", err))
	}
}
