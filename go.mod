module github.com/teamyapp/teamy-backend

go 1.18

require (
	github.com/google/wire v0.5.0
	github.com/gorilla/mux v1.8.0
	github.com/graph-gophers/graphql-go v1.4.0
	github.com/teamyapp/cloud v0.0.0-20220814184025-aed66a5f89d7
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.1
)

//replace github.com/teamyapp/cloud => ../cloud

require (
	github.com/go-gorp/gorp/v3 v3.0.2 // indirect
	github.com/go-ini/ini v1.66.6 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/lib/pq v1.10.6 // indirect
	github.com/minio/minio-go v6.0.14+incompatible // indirect
	github.com/mitchellh/go-homedir v1.0.0 // indirect
	github.com/rubenv/sql-migrate v1.1.2 // indirect
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/net v0.0.0-20220805013720-a33c5aa5df48 // indirect
	golang.org/x/sys v0.0.0-20220804214406-8e32c043e418 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220805133916-01dd62135a58 // indirect
)
