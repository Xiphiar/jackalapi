package http

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

func Start(port int, get Handlers, post Handlers) {
	router := httprouter.New()
	handler := cors.Default().Handler(router)

	for getKey, getFunc := range get {
		router.GET(getKey, getFunc)
	}

	for postKey, postFunc := range post {
		router.POST(postKey, postFunc)
	}

	fmt.Printf("🌍 Started Jackal API: http://0.0.0.0:%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), handler)
	if err != nil {
		fmt.Println(err)
		return
	}

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server Closed\n")
		return
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
