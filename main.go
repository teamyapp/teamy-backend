package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {

		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("Hello World!"))
	})
	fmt.Println("Server started at port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
