package gql

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Money struct {
}

func (m Money) Currency() entity.Currency {
	panic("implement me")
}

func (m Money) Amount() int32 {
	panic("implement me")
}
