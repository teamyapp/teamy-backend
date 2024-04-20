package dao

type Metrics interface {
	ReportDaoOperation(daoName string, operation string)
}
