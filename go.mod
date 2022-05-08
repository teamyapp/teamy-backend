module github.com/teamyapp/teamy-backend

go 1.18

require (
	github.com/google/uuid v1.1.2
	github.com/google/wire v0.5.0
	github.com/graph-gophers/graphql-go v1.2.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/teamyapp/one v0.0.0-20211210090909-75d5a83f5504
)

//replace github.com/teamyapp/one => ../../infra/oneFramework

require (
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/lib/pq v1.10.3 // indirect
	github.com/rubenv/sql-migrate v0.0.0-20211023115951-9f02b1e13857 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1 // indirect
	gopkg.in/gorp.v1 v1.7.2 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
