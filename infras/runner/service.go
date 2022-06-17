package runner

type Service interface {
	Start(runner *ServiceRunner) error
}
