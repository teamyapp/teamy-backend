package core

import (
	_ "embed"
)

//go:embed authorization.yml
var AuthorizationConfig string
