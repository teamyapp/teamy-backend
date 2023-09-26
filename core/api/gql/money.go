package gql

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Money struct {
	deps  *Dependencies
	money entity.Money
}

func (m Money) Currency() entity.Currency {
	return m.money.Currency
}

func (m Money) Amount() int32 {
	return int32(m.money.Amount)
}

func (m Money) Tag() string {
	return m.money.Tag
}

func newMoney(deps *Dependencies, money entity.Money) Money {
	return Money{
		deps:  deps,
		money: money,
	}
}
