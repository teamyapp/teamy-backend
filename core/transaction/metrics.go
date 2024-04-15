package transaction

import (
	"time"
)

type Result string

const (
	CommittedResult  Result = "committed"
	RolledBackResult Result = "rolledBack"
)

type Metrics interface {
	ReportTransactionBegin()
	ReportTransactionCommit()
	ReportTransactionRollback()
	ReportTransactionDuration(duration time.Duration, result Result)
}
