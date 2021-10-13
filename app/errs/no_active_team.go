package errs

import (
	"fmt"

	"github.com/teamyapp/teamy-backend/app/entity"
)

type NoActiveTeam entity.ID

var _ error = (*NoActiveTeam)(nil)

func (n NoActiveTeam) Error() string {
	return fmt.Sprintf("no active team for user: %d", n)
}
