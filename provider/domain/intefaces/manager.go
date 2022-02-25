package intefaces

type Manager interface {
	Run(src DataSource, producer Publisher)
	Stop()
}
