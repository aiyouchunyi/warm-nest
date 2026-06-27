// Package mysqls @Author larry
// File mysql_mapper.go
// @Date 2024/5/22 14:12:00
// @Desc DB映射器
package mysqls

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/database/base"
	"warm-nest/pkg/tool/database/code"
	"warm-nest/pkg/tool/database/query"
	"warm-nest/pkg/tool/database/utils"
	"warm-nest/pkg/tool/optionals"
	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/strings"
	"warm-nest/pkg/utils/times"
)

type MysqlMapper[T any] interface {
	base.DBMapper[T]
	Model(opts ...Option) *gorm.DB
	Silent() *mysqlMapper[T]

	Query(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) ([]T, error)
	QueryAll(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) ([]T, error)
	QueryOne(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (T, error)
	TryQueryOne(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (optionals.Optional[T], error)
	QueryTotal(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (int64, error)
	QueryPage(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (query.QueryResult, error)

	CreateByTx(tx *gorm.DB, model *T) error
	CreateBatchByTx(tx *gorm.DB, model []T) error
	UpdateByTx(tx *gorm.DB, model *T) error
	UpdateBatchByTx(tx *gorm.DB, model []T) error
	SaveByTx(tx *gorm.DB, model *T) error
	SaveBatchByTx(tx *gorm.DB, model []T) error
	DeleteByTx(tx *gorm.DB, model *T) error
}

type mysqlMapper[T any] struct {
	silent bool
}

func NewMapper[T any]() MysqlMapper[T] {
	return &mysqlMapper[T]{}
}

// Model 获取模型
func (m *mysqlMapper[T]) Model(opts ...Option) *gorm.DB {
	var model T
	return DB(append(opts, m.getSilent())...).Model(&model)
}

// Silent 忽略日志
func (m *mysqlMapper[T]) Silent() *mysqlMapper[T] {
	m.silent = true
	return m
}

func (m *mysqlMapper[T]) getSilent() Option {
	defer func() {
		m.silent = false
	}()
	return Silent(m.silent)
}

func (m *mysqlMapper[T]) Log() *logrus.Entry {
	var model T
	return logrus.WithFields(logrus.Fields{
		"model": reflects.ModelName(model),
	})
}

func (m *mysqlMapper[T]) Name() string {
	var model T
	return reflects.ModelName(model)
}

func (m *mysqlMapper[T]) Query(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) ([]T, error) {
	log := m.Log().WithFields(logrus.Fields{
		"req": req,
	})
	expressions, err := Parse(req, fn...)
	if err != nil {
		log.WithError(err).Error("[MYSQL] Query parse err!")
		return nil, errors.NewWithArgs(code.DBParamInvalid, err)
	}
	var models []T
	err = m.Model().Clauses(expressions...).Find(&models).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] Query failed!")
		return nil, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return models, nil
}

// QueryAll 查询所有模型结果
func (m *mysqlMapper[T]) QueryAll(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) ([]T, error) {
	req.All()
	return m.Query(req, fn...)
}

func (m *mysqlMapper[T]) QueryOne(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (T, error) {
	log := m.Log().WithFields(logrus.Fields{
		"req": req,
	})
	req.LimitOne()
	var model T
	expressions, err := Parse(req, fn...)
	if err != nil {
		log.WithError(err).Error("[MYSQL] Query parse err!")
		return model, errors.NewWithArgs(code.DBParamInvalid, err)
	}
	err = m.Model().Clauses(expressions...).First(&model).Error
	if err != nil {
		if IsRecordNotFound(err) {
			log.Error("[MYSQL] Query empty!")
			return model, errors.NewWithArgs(code.DBNotFound, req)
		}
		log.WithError(err).Error("[MYSQL] Query failed!")
		return model, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return model, nil
}

func (m *mysqlMapper[T]) TryQueryOne(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (optionals.Optional[T], error) {
	log := m.Log().WithFields(logrus.Fields{
		"req": req,
	})
	req.LimitOne()
	var model T
	expressions, err := Parse(req, fn...)
	if err != nil {
		log.WithError(err).Error("[MYSQL] Query parse err!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBParamInvalid, err)
	}
	err = m.Model().Clauses(expressions...).First(&model).Error
	if err != nil {
		if IsRecordNotFound(err) {
			return optionals.Empty[T](), nil
		}
		log.WithError(err).Error("[MYSQL] Query failed!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBAbnormal, err)
	}
	return optionals.Of(model), nil
}

func (m *mysqlMapper[T]) QueryTotal(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (int64, error) {
	log := m.Log().WithFields(logrus.Fields{
		"req": req,
	})
	expressions, err := ToConditions(req, fn...)
	if err != nil {
		log.WithError(err).Error("[MYSQL] QueryTotal parse err!")
		return 0, errors.NewWithArgs(code.DBParamInvalid, err)
	}
	var count int64
	err = m.Model().Clauses(expressions...).Count(&count).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] QueryTotal failed!")
		return 0, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return count, nil
}

// QueryPage 查询模型结果
func (m *mysqlMapper[T]) QueryPage(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (query.QueryResult, error) {
	total, err := m.QueryTotal(req, fn...)
	if err != nil {
		return query.QueryResult{}, err
	}
	models, err := m.Query(req, fn...)
	if err != nil {
		return query.QueryResult{}, err
	}
	return query.Result(total, models, req), nil
}

// Count 查询模型数量
func (m *mysqlMapper[T]) Count() (int64, error) {
	var count int64
	err := m.Model().Count(&count).Error
	if err != nil {
		m.Log().WithError(err).Error("[MYSQL] Count failed!")
		return 0, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return count, nil
}

// GetAll 查询全量数据
func (m *mysqlMapper[T]) GetAll() ([]T, error) {
	var models []T
	err := m.Model().Find(&models).Error
	if err != nil {
		m.Log().WithError(err).Error("[MYSQL] Get failed!")
		return nil, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return models, nil
}

// GetById 根据ID获取模型
func (m *mysqlMapper[T]) GetById(id string) (T, error) {
	var model T
	if id == "" {
		m.Log().Error("[MYSQL] GetById id is empty!")
		return model, errors.NewWithArgs(code.DBParamInvalid, "id is empty!")
	}

	log := m.Log().WithFields(logrus.Fields{
		"id": id,
	})
	err := m.Model().First(&model, id).Error
	if err != nil {
		if IsRecordNotFound(err) {
			log.Error("[MYSQL] GetById empty!")
			return model, errors.NewWithArgs(code.DBNotFound, id)
		}
		log.WithError(err).Error("[MYSQL] GetById failed!")
		return model, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return model, nil
}

// TryGetById 尝试根据ID获取模型
func (m *mysqlMapper[T]) TryGetById(id string) (optionals.Optional[T], error) {
	if id == "" {
		m.Log().Error("[MYSQL] TryGetById id is empty!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBParamInvalid, "id is empty!")
	}

	log := m.Log().WithFields(logrus.Fields{
		"id": id,
	})
	var model T
	err := m.Model().First(&model, id).Error
	if err != nil {
		if IsRecordNotFound(err) {
			return optionals.Empty[T](), nil
		}
		log.WithError(err).Error("[MYSQL] TryGetById failed!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBAbnormal, err)
	}
	return optionals.Of(model), nil
}

// GetByIds 根据ID列表获取模型
func (m *mysqlMapper[T]) GetByIds(ids []string) ([]T, error) {
	if len(ids) == 0 {
		m.Log().Warn("[MYSQL] GetByIds ids param err!")
		return nil, nil
	}

	log := m.Log().WithFields(logrus.Fields{
		"ids": ids,
	})
	var models []T
	err := m.Model().Where("id IN ?", ids).Find(&models).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] GetByIds failed!")
		return nil, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return models, nil
}

// GetByUColumn 根据唯一字段获取模型
func (m *mysqlMapper[T]) GetByUColumn(column string, value string) (T, error) {
	log := m.Log().WithFields(logrus.Fields{
		"column": column,
		"value":  value,
	})
	var model T
	if column == "" {
		log.Error("[MYSQL] GetByUColumn column is empty!")
		return model, errors.NewWithArgs(code.DBParamInvalid, "column is empty!")
	}
	if value == "" {
		log.Error("[MYSQL] GetByUColumn value is empty!")
		return model, errors.NewWithArgs(code.DBParamInvalid, "value is empty!")
	}
	err := m.Model().Where(strings.CamelToUnderline(column)+" = ?", value).First(&model).Error
	if err != nil {
		if IsRecordNotFound(err) {
			log.Error("[MYSQL] GetByUColumn empty!")
			return model, errors.NewWithArgs(code.DBNotFound, strings.MergeHyphen(column, value))
		}
		log.WithError(err).Error("[MYSQL] GetByUColumn failed!")
		return model, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return model, nil
}

// TryGetByUColumn 尝试根据唯一字段获取模型
func (m *mysqlMapper[T]) TryGetByUColumn(column string, value string) (optionals.Optional[T], error) {
	log := m.Log().WithFields(logrus.Fields{
		"column": column,
		"value":  value,
	})
	var model T
	if column == "" {
		log.Error("[MYSQL] TryGetByUColumn column is empty!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBParamInvalid, "column is empty!")
	}
	if value == "" {
		log.Error("[MYSQL] TryGetByUColumn value is empty!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBParamInvalid, "value is empty!")
	}
	err := m.Model().Where(strings.CamelToUnderline(column)+" = ?", value).First(&model).Error
	if err != nil {
		if IsRecordNotFound(err) {
			return optionals.Empty[T](), nil
		}
		log.WithError(err).Error("[MYSQL] TryGetByUColumn failed!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBAbnormal, err)
	}
	return optionals.Of(model), nil
}

// ExistByUColumn 判断唯一字段是否存在
func (m *mysqlMapper[T]) ExistByUColumn(column string, value string) (bool, error) {
	optional, err := m.TryGetByUColumn(column, value)
	if err != nil {
		return false, err
	}
	return optional.IsPresent(), nil
}

// GetByUColumns 根据唯一字段列表获取模型
func (m *mysqlMapper[T]) GetByUColumns(column string, values ...string) ([]T, error) {
	log := m.Log().WithFields(logrus.Fields{
		"column": column,
		"values": values,
	})
	if column == "" {
		log.Error("[MYSQL] GetByUColumns column is empty!")
		return nil, errors.NewWithArgs(code.DBParamInvalid, "column is empty!")
	}
	if len(values) == 0 {
		log.Error("[MYSQL] GetByUColumns values is empty!")
		return nil, errors.NewWithArgs(code.DBParamInvalid, "values is empty!")
	}
	var models []T
	err := m.Model().Where(strings.CamelToUnderline(column)+" IN ?", values).Find(&models).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] GetByUColumns failed!")
		return nil, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return models, nil
}

// Create 创建模型
func (m *mysqlMapper[T]) Create(model *T) error {
	return m.CreateByTx(DB(m.getSilent()), model)
}

// CreateBatch 批量创建模型
func (m *mysqlMapper[T]) CreateBatch(models []T) error {
	return DB(m.getSilent()).Transaction(func(tx *gorm.DB) error {
		for index := range models {
			model := &models[index]
			err := m.CreateByTx(tx, model)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// CreateBatchByTx 批量创建模型
func (m *mysqlMapper[T]) CreateBatchByTx(tx *gorm.DB, models []T) error {
	for index := range models {
		model := &models[index]
		err := m.CreateByTx(tx, model)
		if err != nil {
			return err
		}
	}
	return nil
}

// Update 更新模型
func (m *mysqlMapper[T]) Update(model *T) error {
	return m.UpdateByTx(DB(m.getSilent()), model)
}

// UpdateBatch 批量更新模型
func (m *mysqlMapper[T]) UpdateBatch(models []T) error {
	return DB(m.getSilent()).Transaction(func(tx *gorm.DB) error {
		for index := range models {
			model := &models[index]
			err := m.UpdateByTx(tx, model)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateBatchByTx 批量更新模型
func (m *mysqlMapper[T]) UpdateBatchByTx(tx *gorm.DB, models []T) error {
	for index := range models {
		model := &models[index]
		err := m.UpdateByTx(tx, model)
		if err != nil {
			return err
		}
	}
	return nil
}

// Save 保存模型，如果模型存在则更新，否则创建
func (m *mysqlMapper[T]) Save(model *T) error {
	return m.SaveByTx(DB(m.getSilent()), model)
}

// SaveBatch 批量保存模型，如果模型存在则更新，否则创建
func (m *mysqlMapper[T]) SaveBatch(models []T) error {
	return DB(m.getSilent()).Transaction(func(tx *gorm.DB) error {
		for index := range models {
			model := &models[index]
			err := m.SaveByTx(tx, model)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// SaveBatchByTx 批量保存模型，如果模型存在则更新，否则创建
func (m *mysqlMapper[T]) SaveBatchByTx(tx *gorm.DB, models []T) error {
	for index := range models {
		model := &models[index]
		err := m.SaveByTx(tx, model)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteById 根据ID删除模型
func (m *mysqlMapper[T]) DeleteById(id string) error {
	if id == "" {
		m.Log().Error("[MYSQL] DeleteById id is empty!")
		return errors.NewWithArgs(code.DBParamInvalid, "id is empty!")
	}
	log := m.Log().WithFields(logrus.Fields{
		"id": id,
	})
	var model T
	result := DB(m.getSilent()).Where("id = ?", id).Delete(&model)
	if result.Error != nil {
		log.WithError(result.Error).Error("[MYSQL] DeleteById failed!")
		return errors.NewWithArgs(code.DBAbnormal, result.Error)
	}
	if result.RowsAffected == 0 {
		log.Error("[MYSQL] DeleteById affect zero!")
		return errors.NewWithArgs(code.DBAffectZero, id)
	}
	return nil
}

// DeleteByIds 根据ID列表批量删除模型
func (m *mysqlMapper[T]) DeleteByIds(ids []string) error {
	if len(ids) == 0 {
		m.Log().Warn("[MYSQL] DeleteByIds ids is empty!")
		return nil
	}
	log := m.Log().WithFields(logrus.Fields{
		"ids": ids,
	})
	var model T
	err := DB(m.getSilent()).Where("id IN ?", ids).Delete(&model).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] DeleteByIds failed!")
		return errors.NewWithArgs(code.DBAbnormal, err)
	}
	return nil
}

// Delete 删除模型
func (m *mysqlMapper[T]) Delete(model *T) error {
	return m.DeleteByTx(DB(m.getSilent()), model)
}

// DeleteBatch 批量删除模型
func (m *mysqlMapper[T]) DeleteBatch(models []T) error {
	return DB(m.getSilent()).Transaction(func(tx *gorm.DB) error {
		for index := range models {
			model := &models[index]
			err := m.DeleteByTx(tx, model)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteBatchByTx 批量删除模型
func (m *mysqlMapper[T]) DeleteBatchByTx(tx *gorm.DB, models []T) error {
	for index := range models {
		model := &models[index]
		err := m.DeleteByTx(tx, model)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateByTx 创建模型
func (m *mysqlMapper[T]) CreateByTx(tx *gorm.DB, model *T) error {
	if model == nil {
		m.Log().Error("[MYSQL] Create model is nil!")
		return errors.NewWithArgs(code.DBParamInvalid, "model is nil!")
	}

	log := m.Log().WithFields(logrus.Fields{
		"model": model,
	})
	err := tx.Create(model).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] Create failed!")
		return errors.NewWithArgs(code.DBAbnormal, err)
	}
	return nil
}

// UpdateByTx 更新模型
func (m *mysqlMapper[T]) UpdateByTx(tx *gorm.DB, model *T) error {
	if model == nil {
		m.Log().Error("[MYSQL] Update model is nil!")
		return errors.NewWithArgs(code.DBParamInvalid, "model is nil!")
	}
	log := m.Log().WithFields(logrus.Fields{
		"model": model,
	})
	id, err := reflects.GetModelField(model, "ID")
	if err != nil {
		log.Error("[MYSQL] Update model ID err!")
		return errors.NewWithArgs(code.DBParamInvalid, err)
	}
	_ = reflects.SetModelField(model, "UpdatedAt", times.UnixMilli())

	err = tx.Select("*").Where("id = ?", id).UpdateColumns(model).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] Update failed!")
		return errors.NewWithArgs(code.DBAbnormal, err)
	}
	return nil
}

// SaveByTx 保存模型，如果模型存在则更新，否则创建
func (m *mysqlMapper[T]) SaveByTx(tx *gorm.DB, model *T) error {
	if model == nil {
		m.Log().Error("[MYSQL] Save model is nil!")
		return errors.NewWithArgs(code.DBParamInvalid, "model is nil!")
	}
	isNew, err := utils.IsNew(model)
	if err != nil {
		return err
	}
	if isNew {
		return m.CreateByTx(tx, model)
	}
	return m.UpdateByTx(tx, model)
}

// DeleteByTx 删除模型
func (m *mysqlMapper[T]) DeleteByTx(tx *gorm.DB, model *T) error {
	if model == nil {
		logrus.Errorf("[MYSQL] Delete model is nil!")
		return errors.NewWithArgs(code.DBParamInvalid, "model is nil!")
	}
	log := m.Log().WithFields(logrus.Fields{
		"model": model,
	})
	id, err := utils.GetID(model)
	if err != nil {
		log.Error("[MYSQL] Delete model ID err!")
		return errors.NewWithArgs(code.DBParamInvalid, err)
	}
	result := tx.Delete(model)
	if result.Error != nil {
		log.WithError(result.Error).Error("[MYSQL] Delete failed!")
		return errors.NewWithArgs(code.DBAbnormal, result.Error)
	}

	if result.RowsAffected == 0 {
		log.Error("[MYSQL] Delete affect zero!")
		return errors.NewWithArgs(code.DBAffectZero, id)
	}
	return nil
}
