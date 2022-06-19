package web

import (
	"net/http"
)

type Route struct {
	Path        string
	Method      string
	HandlerFunc http.HandlerFunc
}
