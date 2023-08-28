package japicore

import (
	"fmt"
	"net/http"
)

func processError(block string, caughtError error) {
	fmt.Printf("***** Error in block: %s *****\n", block)
	fmt.Println(caughtError)
	fmt.Println("***** End Error Report *****")

}

func processHttpPostError(block string, caughtError error, w http.ResponseWriter) {

	fmt.Printf("***** Error in block: %s *****\n", block)
	fmt.Println(caughtError)
	fmt.Println("***** End Error Report *****")
	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte(caughtError.Error()))
	if err != nil {
		processError(fmt.Sprintf("processHttpPostError for %s", block), err)
	}
}
