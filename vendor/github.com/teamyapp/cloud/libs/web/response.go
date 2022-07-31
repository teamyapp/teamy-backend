package web

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteJSON(writer http.ResponseWriter, body interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusAccepted)

	buf, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Write(buf)
}
