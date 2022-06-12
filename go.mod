module github.com/teamyapp/teamy-backend

go 1.18

require (
	github.com/google/uuid v1.1.2
	github.com/google/wire v0.5.0
	github.com/gorilla/mux v1.8.0
	github.com/graph-gophers/graphql-go v1.2.0
	github.com/teamyapp/cloud v0.0.0-20220519073334-a938c55555c7
)

//replace github.com/teamyapp/cloud => ../cloud

require (
	github.com/go-gorp/gorp/v3 v3.0.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/lib/pq v1.10.5 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/rubenv/sql-migrate v1.1.1 // indirect
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	google.golang.org/grpc v1.38.0 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
)
