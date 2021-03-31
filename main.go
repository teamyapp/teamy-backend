package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/teamyapp/template-go/app"
)

func main() {
	rand.Seed(time.Now().Unix())

	http.HandleFunc("/random", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)

		sum := app.Add(rand.Int(), rand.Int())
		writer.Write([]byte(fmt.Sprintf("%d\n", sum)))
	})
	fmt.Println("Server started at port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
