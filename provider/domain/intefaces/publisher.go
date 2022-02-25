package intefaces

import "provider/domain/models"

type Publisher interface {
	Publish(data models.Data) error
}
