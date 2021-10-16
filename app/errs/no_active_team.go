package errs

import (
	"fmt"

	oneEntity "github.com/teamyapp/one/entity"
)

type NoActiveTeam oneEntity.ID

var _ error = (*NoActiveTeam)(nil)

func (n NoActiveTeam) Error() string {
	return fmt.Sprintf("no active team for user: %d", n)
}
