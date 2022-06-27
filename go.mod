module github.com/teamyapp/teamy-backend

go 1.18

require (
	github.com/google/uuid v1.3.0
	github.com/google/wire v0.5.0
	github.com/graph-gophers/graphql-go v1.4.0
	github.com/teamyapp/cloud v0.0.0-20220619182653-336d1be89ddf
)

replace github.com/teamyapp/cloud => ../cloud

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
	golang.org/x/net v0.0.0-20220617184016-355a448f1bc9 // indirect
	golang.org/x/sys v0.0.0-20220615213510-4f61da869c0c // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220617124728-180714bec0ad // indirect
	google.golang.org/grpc v1.47.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)
