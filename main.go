package main

import (
	"fmt"
	"github.com/teamyapp/template-go/app"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/random", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)

		sum := app.Add(app.RandInt(), app.RandInt())
		writer.Write([]byte(fmt.Sprintf("%d\n", sum)))
	})
	fmt.Println("Server started at port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
