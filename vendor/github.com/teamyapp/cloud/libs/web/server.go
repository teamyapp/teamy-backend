package web

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func StartWebServer(router *mux.Router, port int) error {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", enableCORS(router.ServeHTTP))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), serveMux)
}

func enableCORS(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS, DELETE")
		writer.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		if request.Method == http.MethodOptions {
			return
		}

		handlerFunc(writer, request)
	}
}

func WriteJSON(writer http.ResponseWriter, body []byte) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)
	writer.Write(body)
}
