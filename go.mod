module github.com/teamyapp/teamy-backend

go 1.18

require (
	github.com/google/wire v0.5.0
	github.com/graph-gophers/graphql-go v1.4.0
	github.com/teamyapp/cloud v0.0.0-20220706074550-502105c67e49
	google.golang.org/grpc v1.47.0
	google.golang.org/protobuf v1.28.0
)

//replace github.com/teamyapp/cloud => ../cloud

require (
	github.com/go-gorp/gorp/v3 v3.0.2 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/lib/pq v1.10.6 // indirect
	github.com/rubenv/sql-migrate v1.1.2 // indirect
	golang.org/x/net v0.0.0-20220630215102-69896b714898 // indirect
	golang.org/x/sys v0.0.0-20220704084225-05e143d24a9e // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220630174209-ad1d48641aa7 // indirect
)
