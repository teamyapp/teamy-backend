package main

import (
	"fmt"

	"github.com/teamyapp/teamy-backend/app/api/gql"
)

func main() {
	server, err := gql.NewServer(9000)
	if err != nil {
		panic(err)
	}

	fmt.Println("GraphQL server started at 9000")
	panic(server.ListenAndServe())
}
