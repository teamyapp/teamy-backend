module github.com/teamyapp/teamy-backend

go 1.17

require (
	github.com/google/wire v0.5.0
	github.com/graph-gophers/graphql-go v1.2.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/stretchr/testify v1.7.0
	github.com/teamyapp/one v0.0.0-20211107052442-dd702673447c
)

replace github.com/teamyapp/one => ../../tool/oneFramework

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang-migrate/migrate/v4 v4.15.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/lib/pq v1.10.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
