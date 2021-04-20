package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/teamyapp/template-go/app"
)

const version = 1

func main() {
	rand.Seed(time.Now().Unix())

	http.HandleFunc("/random", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)

		randIntA := rand.Int()
		randIntB := rand.Int()

		sum := app.Add(randIntA, randIntB)
		writer.Write([]byte(fmt.Sprintf(
			strings.TrimPrefix(`
Version = %d
Random Int A = %d
Random Int B = %d
Sum = %d`, "\n"), version, randIntA, randIntB, sum)))
	})
	fmt.Println("Server started at port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
