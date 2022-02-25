package intefaces

import "provider/domain/models"

type DataSource interface {
	Run() <-chan models.Data
	Stop()
}
