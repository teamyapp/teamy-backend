package sqldb

import (
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Story struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
}

var _ dao.Task = (*Story)(nil)

func ()  {
	
}
