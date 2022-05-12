module github.com/teamyapp/teamy-backend

go 1.18

require (
	github.com/google/uuid v1.1.2
	github.com/google/wire v0.5.0
	github.com/graph-gophers/graphql-go v1.2.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/teamyapp/cloud v0.0.0-20220512061748-cba9459a3918
)

//replace github.com/teamyapp/cloud => /Users/harryliu/Documents/projects/teamyapp/apps/cloud

require (
	github.com/go-gorp/gorp/v3 v3.0.2 // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/lib/pq v1.10.5 // indirect
	github.com/rubenv/sql-migrate v1.1.1 // indirect
)
