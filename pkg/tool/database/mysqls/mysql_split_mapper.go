// Package mysqls @Author larry
// File mysql_mapper.go
// @Date 2024/5/22 14:12:00
// @Desc DB映射器
package mysqls

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/database/code"
	"warm-nest/pkg/tool/database/query"
	"warm-nest/pkg/tool/database/utils"
	"warm-nest/pkg/tool/optionals"
	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/strings"
	"warm-nest/pkg/utils/times"
)

type MysqlIndexMapper[T MySQLSplit] interface {
	Model(tableIndex int, opts ...Option) *gorm.DB
	Silent() *mysqlIndexMapper[T]

	Log(tableIndex int) *logrus.Entry
	TableName(tableIndex int) string

	Count(tableIndex int) (int64, error)
	GetAll(tableIndex int) ([]T, error)

	GetById(tableIndex int, id string) (T, error)
	TryGetById(tableIndex int, id string) (optionals.Optional[T], error)
	GetByIds(tableIndex int, id []string) ([]T, error)

	GetByUColumn(tableIndex int, column string, value string) (T, error)
	TryGetByUColumn(tableIndex int, column string, value string) (optionals.Optional[T], error)
	ExistByUColumn(tableIndex int, column string, value string) (bool, error)
	GetByUColumns(tableIndex int, column string, values ...string) ([]T, error)

	Create(tableIndex int, model *T) error
	CreateBatch(tableIndex int, models []T) error
	Update(tableIndex int, model *T) error
	UpdateBatch(tableIndex int, models []T) error
	Save(tableIndex int, model *T) error
	SaveBatch(tableIndex int, models []T) error
	DeleteById(tableIndex int, id string) error
	DeleteByIds(tableIndex int, ids []string) error
	Delete(tableIndex int, model *T) error
	DeleteBatch(tableIndex int, models []T) error

	Query(tableIndex int, req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) ([]T, error)
	QueryAll(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) ([]T, error)
	QueryOne(tableIndex int, req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (T, error)
	TryQueryOne(tableIndex int, req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (optionals.Optional[T], error)
	QueryTotalAll(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (int64, error)
	QueryTotal(tableIndex int, req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (int64, error)
	QueryPage(tableIndex int, req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (query.QueryResult, error)

	CreateByTx(tx *gorm.DB, tableIndex int, model *T) error
	UpdateByTx(tx *gorm.DB, tableIndex int, model *T) error
	SaveByTx(tx *gorm.DB, tableIndex int, model *T) error
	DeleteByTx(tx *gorm.DB, tableIndex int, model *T) error
}

type mysqlIndexMapper[T MySQLSplit] struct {
	silent bool
}

func NewIndexMapper[T MySQLSplit]() MysqlIndexMapper[T] {
	return &mysqlIndexMapper[T]{}
}

// Model 获取模型
func (m *mysqlIndexMapper[T]) Model(tableIndex int, opts ...Option) *gorm.DB {
	record := reflects.New[T]()
	return DB(append(opts, m.getSilent())...).Table(record.GetTableNameByIndex(tableIndex)).Model(&record)
}

// Silent 忽略日志
func (m *mysqlIndexMapper[T]) Silent() *mysqlIndexMapper[T] {
	m.silent = true
	return m
}

func (m *mysqlIndexMapper[T]) getSilent() Option {
	defer func() {
		m.silent = false
	}()
	return Silent(m.silent)
}

func (m *mysqlIndexMapper[T]) Log(tableIndex int) *logrus.Entry {
	record := reflects.New[T]()
	return logrus.WithFields(logrus.Fields{
		"tableName": record.GetTableNameByIndex(tableIndex),
	})
}

func (m *mysqlIndexMapper[T]) TableName(tableIndex int) string {
	record := reflects.New[T]()
	return record.GetTableNameByIndex(tableIndex)
}

func (m *mysqlIndexMapper[T]) QueryAll(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) ([]T, error) {
	var records []T
	for i := 0; i < 10; i++ {
		partRecords, err := m.Query(i, req, fn...)
		if err != nil {
			logrus.WithError(err).Errorf("Query part table %d failed!", i)
			return nil, err
		}
		records = append(records, partRecords...)
	}
	return records, nil
}

func (m *mysqlIndexMapper[T]) Query(tableIndex int, req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) ([]T, error) {
	log := m.Log(tableIndex).WithFields(logrus.Fields{
		"req": req,
	})
	expressions, err := Parse(req, fn...)
	if err != nil {
		log.WithError(err).Error("[MYSQL] Query parse err!")
		return nil, errors.NewWithArgs(code.DBParamInvalid, err)
	}
	var models []T
	err = m.Model(tableIndex).Clauses(expressions...).Find(&models).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] Query failed!")
		return nil, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return models, nil
}

func (m *mysqlIndexMapper[T]) QueryOne(tableIndex int, req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (T, error) {
	log := m.Log(tableIndex).WithFields(logrus.Fields{
		"req": req,
	})
	req.LimitOne()
	var model T
	expressions, err := Parse(req, fn...)
	if err != nil {
		log.WithError(err).Error("[MYSQL] Query parse err!")
		return model, errors.NewWithArgs(code.DBParamInvalid, err)
	}
	err = m.Model(tableIndex).Clauses(expressions...).First(&model).Error
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

func (m *mysqlIndexMapper[T]) TryQueryOne(tableIndex int, req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (optionals.Optional[T], error) {
	log := m.Log(tableIndex).WithFields(logrus.Fields{
		"req": req,
	})
	req.LimitOne()
	var model T
	expressions, err := Parse(req, fn...)
	if err != nil {
		log.WithError(err).Error("[MYSQL] Query parse err!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBParamInvalid, err)
	}
	err = m.Model(tableIndex).Clauses(expressions...).First(&model).Error
	if err != nil {
		if IsRecordNotFound(err) {
			return optionals.Empty[T](), nil
		}
		log.WithError(err).Error("[MYSQL] Query failed!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBAbnormal, err)
	}
	return optionals.Of(model), nil
}

func (m *mysqlIndexMapper[T]) QueryTotalAll(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (int64, error) {
	var total int64
	for i := 0; i < 10; i++ {
		partTotal, err := m.QueryTotal(i, req, fn...)
		if err != nil {
			logrus.WithError(err).Errorf("QueryTotal part table %d failed!", i)
			return 0, err
		}
		total += partTotal
	}
	return total, nil
}

func (m *mysqlIndexMapper[T]) QueryTotal(tableIndex int, req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (int64, error) {
	log := m.Log(tableIndex).WithFields(logrus.Fields{
		"req": req,
	})
	expressions, err := ToConditions(req, fn...)
	if err != nil {
		log.WithError(err).Error("[MYSQL] QueryTotal parse err!")
		return 0, errors.NewWithArgs(code.DBParamInvalid, err)
	}
	var count int64
	err = m.Model(tableIndex).Clauses(expressions...).Count(&count).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] QueryTotal failed!")
		return 0, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return count, nil
}

// QueryPage 查询模型结果
func (m *mysqlIndexMapper[T]) QueryPage(tableIndex int, req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) (query.QueryResult, error) {
	total, err := m.QueryTotal(tableIndex, req, fn...)
	if err != nil {
		return query.QueryResult{}, err
	}
	models, err := m.Query(tableIndex, req, fn...)
	if err != nil {
		return query.QueryResult{}, err
	}
	return query.Result(total, models, req), nil
}

// Count 查询模型数量
func (m *mysqlIndexMapper[T]) Count(tableIndex int) (int64, error) {
	var count int64
	err := m.Model(tableIndex).Count(&count).Error
	if err != nil {
		m.Log(tableIndex).WithError(err).Error("[MYSQL] Count failed!")
		return 0, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return count, nil
}

// GetAll 查询全量数据
func (m *mysqlIndexMapper[T]) GetAll(tableIndex int) ([]T, error) {
	var models []T
	err := m.Model(tableIndex).Find(&models).Error
	if err != nil {
		m.Log(tableIndex).WithError(err).Error("[MYSQL] Get failed!")
		return nil, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return models, nil
}

// GetById 根据ID获取模型
func (m *mysqlIndexMapper[T]) GetById(tableIndex int, id string) (T, error) {
	var model T
	if id == "" {
		m.Log(tableIndex).Error("[MYSQL] GetById id is empty!")
		return model, errors.NewWithArgs(code.DBParamInvalid, "id is empty!")
	}

	log := m.Log(tableIndex).WithFields(logrus.Fields{
		"id": id,
	})
	err := m.Model(tableIndex).First(&model, id).Error
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
func (m *mysqlIndexMapper[T]) TryGetById(tableIndex int, id string) (optionals.Optional[T], error) {
	if id == "" {
		m.Log(tableIndex).Error("[MYSQL] TryGetById id is empty!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBParamInvalid, "id is empty!")
	}

	log := m.Log(tableIndex).WithFields(logrus.Fields{
		"id": id,
	})
	var model T
	err := m.Model(tableIndex).First(&model, id).Error
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
func (m *mysqlIndexMapper[T]) GetByIds(tableIndex int, ids []string) ([]T, error) {
	if len(ids) == 0 {
		m.Log(tableIndex).Warn("[MYSQL] GetByIds ids param err!")
		return nil, nil
	}

	log := m.Log(tableIndex).WithFields(logrus.Fields{
		"ids": ids,
	})
	var models []T
	err := m.Model(tableIndex).Where("id IN ?", ids).Find(&models).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] GetByIds failed!")
		return nil, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return models, nil
}

// GetByUColumn 根据唯一字段获取模型
func (m *mysqlIndexMapper[T]) GetByUColumn(tableIndex int, column string, value string) (T, error) {
	log := m.Log(tableIndex).WithFields(logrus.Fields{
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
	err := m.Model(tableIndex).Where(strings.CamelToUnderline(column)+" = ?", value).First(&model).Error
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
func (m *mysqlIndexMapper[T]) TryGetByUColumn(tableIndex int, column string, value string) (optionals.Optional[T], error) {
	log := m.Log(tableIndex).WithFields(logrus.Fields{
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
	err := m.Model(tableIndex).Where(strings.CamelToUnderline(column)+" = ?", value).First(&model).Error
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
func (m *mysqlIndexMapper[T]) ExistByUColumn(tableIndex int, column string, value string) (bool, error) {
	optional, err := m.TryGetByUColumn(tableIndex, column, value)
	if err != nil {
		return false, err
	}
	return optional.IsPresent(), nil
}

// GetByUColumns 根据唯一字段列表获取模型
func (m *mysqlIndexMapper[T]) GetByUColumns(tableIndex int, column string, values ...string) ([]T, error) {
	log := m.Log(tableIndex).WithFields(logrus.Fields{
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
	err := m.Model(tableIndex).Where(strings.CamelToUnderline(column)+" IN ?", values).Find(&models).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] GetByUColumns failed!")
		return nil, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return models, nil
}

// Create 创建模型
func (m *mysqlIndexMapper[T]) Create(tableIndex int, model *T) error {
	return m.CreateByTx(DB(m.getSilent()), tableIndex, model)
}

// CreateBatch 批量创建模型
func (m *mysqlIndexMapper[T]) CreateBatch(tableIndex int, models []T) error {
	return DB(m.getSilent()).Transaction(func(tx *gorm.DB) error {
		for index := range models {
			model := &models[index]
			err := m.CreateByTx(tx, tableIndex, model)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Update 更新模型
func (m *mysqlIndexMapper[T]) Update(tableIndex int, model *T) error {
	return m.UpdateByTx(DB(m.getSilent()), tableIndex, model)
}

// UpdateBatch 批量更新模型
func (m *mysqlIndexMapper[T]) UpdateBatch(tableIndex int, models []T) error {
	return DB(m.getSilent()).Transaction(func(tx *gorm.DB) error {
		for index := range models {
			model := &models[index]
			err := m.UpdateByTx(tx, tableIndex, model)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Save 保存模型，如果模型存在则更新，否则创建
func (m *mysqlIndexMapper[T]) Save(tableIndex int, model *T) error {
	return m.SaveByTx(DB(m.getSilent()), tableIndex, model)
}

// SaveBatch 批量保存模型，如果模型存在则更新，否则创建
func (m *mysqlIndexMapper[T]) SaveBatch(tableIndex int, models []T) error {
	return DB(m.getSilent()).Transaction(func(tx *gorm.DB) error {
		for index := range models {
			model := &models[index]
			err := m.SaveByTx(tx, tableIndex, model)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteById 根据ID删除模型
func (m *mysqlIndexMapper[T]) DeleteById(tableIndex int, id string) error {
	if id == "" {
		m.Log(tableIndex).Error("[MYSQL] DeleteById id is empty!")
		return errors.NewWithArgs(code.DBParamInvalid, "id is empty!")
	}
	log := m.Log(tableIndex).WithFields(logrus.Fields{
		"id": id,
	})
	var model T
	result := DB(m.getSilent()).Table(m.TableName(tableIndex)).Where("id = ?", id).Delete(&model)
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
func (m *mysqlIndexMapper[T]) DeleteByIds(tableIndex int, ids []string) error {
	if len(ids) == 0 {
		m.Log(tableIndex).Warn("[MYSQL] DeleteByIds ids is empty!")
		return nil
	}
	log := m.Log(tableIndex).WithFields(logrus.Fields{
		"ids": ids,
	})
	var model T
	err := DB(m.getSilent()).Table(m.TableName(tableIndex)).Where("id IN ?", ids).Delete(&model).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] DeleteByIds failed!")
		return errors.NewWithArgs(code.DBAbnormal, err)
	}
	return nil
}

// Delete 删除模型
func (m *mysqlIndexMapper[T]) Delete(tableIndex int, model *T) error {
	return m.DeleteByTx(DB(m.getSilent()), tableIndex, model)
}

// DeleteBatch 批量删除模型
func (m *mysqlIndexMapper[T]) DeleteBatch(tableIndex int, models []T) error {
	return DB(m.getSilent()).Transaction(func(tx *gorm.DB) error {
		for index := range models {
			model := &models[index]
			err := m.DeleteByTx(tx, tableIndex, model)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// CreateByTx 创建模型
func (m *mysqlIndexMapper[T]) CreateByTx(tx *gorm.DB, tableIndex int, model *T) error {
	if model == nil {
		m.Log(tableIndex).Error("[MYSQL] Create model is nil!")
		return errors.NewWithArgs(code.DBParamInvalid, "model is nil!")
	}

	log := m.Log(tableIndex).WithFields(logrus.Fields{
		"model": model,
	})
	err := tx.Table(m.TableName(tableIndex)).Create(model).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] Create failed!")
		return errors.NewWithArgs(code.DBAbnormal, err)
	}
	return nil
}

// UpdateByTx 更新模型
func (m *mysqlIndexMapper[T]) UpdateByTx(tx *gorm.DB, tableIndex int, model *T) error {
	if model == nil {
		m.Log(tableIndex).Error("[MYSQL] Update model is nil!")
		return errors.NewWithArgs(code.DBParamInvalid, "model is nil!")
	}
	log := m.Log(tableIndex).WithFields(logrus.Fields{
		"model": model,
	})
	id, err := reflects.GetModelField(model, "ID")
	if err != nil {
		log.Error("[MYSQL] Update model ID err!")
		return errors.NewWithArgs(code.DBParamInvalid, err)
	}
	_ = reflects.SetModelField(model, "UpdatedAt", times.UnixMilli())

	err = tx.Table(m.TableName(tableIndex)).Select("*").Where("id = ?", id).UpdateColumns(model).Error
	if err != nil {
		log.WithError(err).Error("[MYSQL] Update failed!")
		return errors.NewWithArgs(code.DBAbnormal, err)
	}
	return nil
}

// SaveByTx 保存模型，如果模型存在则更新，否则创建
func (m *mysqlIndexMapper[T]) SaveByTx(tx *gorm.DB, tableIndex int, model *T) error {
	if model == nil {
		m.Log(tableIndex).Error("[MYSQL] Save model is nil!")
		return errors.NewWithArgs(code.DBParamInvalid, "model is nil!")
	}
	isNew, err := utils.IsNew(model)
	if err != nil {
		return err
	}
	if isNew {
		return m.CreateByTx(tx, tableIndex, model)
	}
	return m.UpdateByTx(tx, tableIndex, model)
}

// DeleteByTx 删除模型
func (m *mysqlIndexMapper[T]) DeleteByTx(tx *gorm.DB, tableIndex int, model *T) error {
	if model == nil {
		logrus.Errorf("[MYSQL] Delete model is nil!")
		return errors.NewWithArgs(code.DBParamInvalid, "model is nil!")
	}
	log := m.Log(tableIndex).WithFields(logrus.Fields{
		"model": model,
	})
	id, err := utils.GetID(model)
	if err != nil {
		log.Error("[MYSQL] Delete model ID err!")
		return errors.NewWithArgs(code.DBParamInvalid, err)
	}
	result := tx.Table(m.TableName(tableIndex)).Delete(model)
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
