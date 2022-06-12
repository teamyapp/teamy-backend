module github.com/teamyapp/teamy-backend

go 1.18

require (
	github.com/google/uuid v1.3.0
	github.com/google/wire v0.5.0
	github.com/gorilla/mux v1.8.0
	github.com/graph-gophers/graphql-go v1.4.0
	github.com/teamyapp/cloud v0.0.0-20220612225926-bed0a52c4cba
)

//replace github.com/teamyapp/cloud => ../cloud

require (
	github.com/go-gorp/gorp/v3 v3.0.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/lib/pq v1.10.6 // indirect
	github.com/rubenv/sql-migrate v1.1.2 // indirect
	golang.org/x/net v0.0.0-20220607020251-c690dde0001d // indirect
	golang.org/x/sys v0.0.0-20220610221304-9f5ed59c137d // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220608133413-ed9918b62aac // indirect
	google.golang.org/grpc v1.47.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)
