package server

import (
	"fmt"
	"net/http"
)

func StartServer(bindAddress string) {
	setUpRoutes()
	err := http.ListenAndServe(bindAddress, nil)
	if err != nil {
		panic(fmt.Sprintf("Problem starting web server: %v", err))
	}
}
