// Package database @Author larry
// @Date 2025/4/18 17:16
// @Desc

package base

import (
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/tool/optionals"
)

type DBMapper[T any] interface {
	Log() *logrus.Entry
	Name() string

	Count() (int64, error)
	GetAll() ([]T, error)

	GetById(id string) (T, error)
	TryGetById(id string) (optionals.Optional[T], error)
	GetByIds(id []string) ([]T, error)

	GetByUColumn(column string, value string) (T, error)
	TryGetByUColumn(column string, value string) (optionals.Optional[T], error)
	ExistByUColumn(column string, value string) (bool, error)
	GetByUColumns(column string, values ...string) ([]T, error)

	Create(model *T) error
	CreateBatch(models []T) error
	Update(model *T) error
	UpdateBatch(models []T) error
	Save(model *T) error
	SaveBatch(models []T) error
	DeleteById(id string) error
	DeleteByIds(ids []string) error
	Delete(model *T) error
	DeleteBatch(models []T) error
}
