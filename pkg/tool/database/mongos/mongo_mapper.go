// Package mongos @Author larry
// @Date 2025/4/17 19:34
// @Desc

package mongos

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/database/base"
	"warm-nest/pkg/tool/database/code"
	"warm-nest/pkg/tool/database/mongos/widget"
	"warm-nest/pkg/tool/database/query"
	"warm-nest/pkg/tool/database/utils"
	"warm-nest/pkg/tool/optionals"
	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/slices"
	"warm-nest/pkg/utils/strings"
	"warm-nest/pkg/utils/times"
)

type MongoMapper[T any] interface {
	base.DBMapper[T]
	Collection() *mongo.Collection
	FindOne(filter interface{}, opts ...*options.FindOneOptions) (T, error)
	Find(filter interface{}, opts ...*options.FindOptions) ([]T, error)
	Exist(filter interface{}) (bool, error)

	Query(req query.QueryReq, fn ...func(c query.Custom) (bson.M, error)) ([]T, error)
	QueryOne(req query.QueryReq, fn ...func(c query.Custom) (bson.M, error)) (T, error)
	TryQueryOne(req query.QueryReq, fn ...func(c query.Custom) (bson.M, error)) (optionals.Optional[T], error)
	QueryTotal(req query.QueryReq, fn ...func(c query.Custom) (bson.M, error)) (int64, error)
	QueryPage(req query.QueryReq, fn ...func(c query.Custom) (bson.M, error)) (query.QueryResult, error)
}

type mongoMapper[T any] struct {
	database   string
	collection string
}

func NewMongoMapper[T any](collection string) MongoMapper[T] {
	mapper := &mongoMapper[T]{
		database:   strings.LeftStr(collection, strings.DotSplitChar),
		collection: strings.RightStr(collection, strings.DotSplitChar),
	}
	err := widget.CreateIndex(mapper.Collection(), reflects.New[T]())
	if err != nil {
		logrus.WithError(err).Error("CreateIndex failed!")
	}
	return mapper
}

func (m *mongoMapper[T]) Collection() *mongo.Collection {
	return Client().Database(m.database).Collection(m.collection)
}

func (m *mongoMapper[T]) Log() *logrus.Entry {
	var model T
	return logrus.WithFields(logrus.Fields{
		"model": reflects.ModelName(model),
	})
}

func (m *mongoMapper[T]) Name() string {
	var model T
	return reflects.ModelName(model)
}

func (m *mongoMapper[T]) Query(req query.QueryReq, fn ...func(c query.Custom) (bson.M, error)) ([]T, error) {
	filter, err := Parse(req, fn...)
	if err != nil {
		m.Log().WithError(err).Errorf("[MONGO] Query parse failed!")
		return nil, errors.NewWithArgs(code.DBParamInvalid, err)
	}
	return m.Find(filter, FindOptions(req))
}

func (m *mongoMapper[T]) QueryOne(req query.QueryReq, fn ...func(c query.Custom) (bson.M, error)) (T, error) {
	var result T
	filter, err := Parse(req, fn...)
	if err != nil {
		m.Log().WithError(err).Errorf("[MONGO] QueryOne parse failed!")
		return result, errors.NewWithArgs(code.DBParamInvalid, err)
	}
	err = m.Collection().FindOne(nil, filter, FindOneOptions(req)).Decode(&result)
	if err != nil {
		if IsNoDocuments(err) {
			return result, errors.NewWithArgs(code.DBNotFound, req)
		}
		m.Log().WithError(err).Errorf("[MONGO] QueryOne failed!")
		return result, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return result, nil
}

func (m *mongoMapper[T]) TryQueryOne(req query.QueryReq, fn ...func(c query.Custom) (bson.M, error)) (optionals.Optional[T], error) {
	var result T
	filters, err := Parse(req, fn...)
	if err != nil {
		m.Log().WithError(err).Errorf("[MONGO] TryQueryOne parse failed!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBParamInvalid, err)
	}
	err = m.Collection().FindOne(nil, filters, FindOneOptions(req)).Decode(&result)
	if err != nil {
		if IsNoDocuments(err) {
			return optionals.Empty[T](), nil
		}
		m.Log().WithError(err).Errorf("[MONGO] TryQueryOne failed!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBAbnormal, err)
	}
	return optionals.Of(result), nil
}

func (m *mongoMapper[T]) QueryTotal(req query.QueryReq, fn ...func(c query.Custom) (bson.M, error)) (int64, error) {
	filter, err := Parse(req, fn...)
	if err != nil {
		m.Log().WithError(err).Errorf("[MONGO] QueryTotal parse failed!")
		return 0, errors.NewWithArgs(code.DBParamInvalid, err)
	}
	count, err := m.Collection().CountDocuments(nil, filter)
	if err != nil {
		m.Log().WithError(err).Errorf("[MONGO] QueryTotal failed!")
		return 0, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return count, nil
}

func (m *mongoMapper[T]) QueryPage(req query.QueryReq, fn ...func(c query.Custom) (bson.M, error)) (query.QueryResult, error) {
	total, err := m.QueryTotal(req, fn...)
	if err != nil {
		return query.QueryResult{}, nil
	}
	result, err := m.Query(req, fn...)
	if err != nil {
		m.Log().WithError(err).Errorf("[MONGO] QueryPage find failed!")
		return query.QueryResult{}, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return query.Result(total, result, req), nil
}

func (m *mongoMapper[T]) Count() (int64, error) {
	count, err := m.Collection().CountDocuments(nil, bson.M{})
	if err != nil {
		m.Log().WithError(err).Errorf("[MONGO] Count failed!")
		return 0, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return count, nil
}

func (m *mongoMapper[T]) GetAll() ([]T, error) {
	return m.Find(bson.M{})
}

func (m *mongoMapper[T]) GetById(id string) (T, error) {
	var result T
	if id == "" {
		m.Log().Error("[MONGO] GetById id is empty!")
		return result, errors.NewWithArgs(code.DBParamInvalid, "id is empty!")
	}

	log := m.Log().WithFields(logrus.Fields{
		"id": id,
	})
	objectID, _ := primitive.ObjectIDFromHex(id)
	err := m.Collection().FindOne(nil, bson.M{"_id": objectID}).Decode(&result)
	if err != nil {
		if IsNoDocuments(err) {
			log.Errorf("[MONGO] GetById empty!")
			return result, errors.NewWithArgs(code.DBNotFound, id)
		}
		log.WithError(err).Errorf("[MONGO] GetById failed!")
		return result, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return result, nil
}

func (m *mongoMapper[T]) TryGetById(id string) (optionals.Optional[T], error) {
	var result T
	if id == "" {
		m.Log().Error("[MONGO] TryGetById id is empty!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBParamInvalid, "id is empty!")
	}

	log := m.Log().WithFields(logrus.Fields{
		"id": id,
	})
	objectID, _ := primitive.ObjectIDFromHex(id)
	err := m.Collection().FindOne(nil, bson.M{"_id": objectID}).Decode(&result)
	if err != nil {
		if IsNoDocuments(err) {
			return optionals.Empty[T](), nil
		}
		log.WithError(err).Errorf("[MONGO] TryGetById failed!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBAbnormal, err)
	}
	return optionals.Of(result), nil
}

func (m *mongoMapper[T]) GetByIds(ids []string) ([]T, error) {
	var result []T
	if len(ids) == 0 {
		m.Log().Error("[MONGO] GetByIds ids is empty!")
		return result, errors.NewWithArgs(code.DBParamInvalid, "ids is empty!")
	}

	log := m.Log().WithFields(logrus.Fields{
		"ids": ids,
	})

	filter := bson.M{"_id": bson.M{"$in": slices.Map(ids, func(id string) primitive.ObjectID {
		objectID, _ := primitive.ObjectIDFromHex(id)
		return objectID
	})}}
	result, err := m.Find(filter)
	if err != nil {
		log.WithError(err).Errorf("[MONGO] GetByIds failed!")
		return nil, err
	}
	return result, nil
}

func (m *mongoMapper[T]) GetByUColumn(column string, value string) (T, error) {
	log := m.Log().WithFields(logrus.Fields{
		"column": column,
		"value":  value,
	})
	var result T
	if column == "" {
		m.Log().Error("[MONGO] GetByUColumn column is empty!")
		return result, errors.NewWithArgs(code.DBParamInvalid, "column is empty!")
	}
	if value == "" {
		m.Log().Error("[MONGO] GetByUColumn value is empty!")
		return result, errors.NewWithArgs(code.DBParamInvalid, "value is empty!")
	}
	err := m.Collection().FindOne(nil, bson.M{strings.LowerFirst(column): value}).Decode(&result)
	if err != nil {
		if IsNoDocuments(err) {
			log.Errorf("[MONGO] GetByUColumn empty!")
			return result, errors.NewWithArgs(code.DBNotFound, value)
		}
		log.WithError(err).Errorf("[MONGO] GetByUColumn failed!")
		return result, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return result, nil
}

func (m *mongoMapper[T]) TryGetByUColumn(column string, value string) (optionals.Optional[T], error) {
	var result T
	if column == "" {
		m.Log().Error("[MONGO] TryGetByUColumn column is empty!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBParamInvalid, "column is empty!")
	}
	if value == "" {
		m.Log().Error("[MONGO] TryGetByUColumn value is empty!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBParamInvalid, "value is empty!")
	}

	log := m.Log().WithFields(logrus.Fields{
		"column": column,
		"value":  value,
	})
	err := m.Collection().FindOne(nil, bson.M{strings.LowerFirst(column): value}).Decode(&result)
	if err != nil {
		if IsNoDocuments(err) {
			return optionals.Empty[T](), nil
		}
		log.WithError(err).Errorf("[MONGO] TryGetByUColumn failed!")
		return optionals.Empty[T](), errors.NewWithArgs(code.DBAbnormal, err)
	}
	return optionals.Of(result), nil
}

func (m *mongoMapper[T]) ExistByUColumn(column string, value string) (bool, error) {
	if column == "" {
		m.Log().Error("[MONGO] ExistByUColumn column is empty!")
		return false, errors.NewWithArgs(code.DBParamInvalid, "column is empty!")
	}
	if value == "" {
		m.Log().Error("[MONGO] ExistByUColumn value is empty!")
		return false, errors.NewWithArgs(code.DBParamInvalid, "value is empty!")
	}

	log := m.Log().WithFields(logrus.Fields{
		"column": column,
		"value":  value,
	})
	count, err := m.Collection().CountDocuments(nil, bson.M{strings.LowerFirst(column): value})
	if err != nil {
		log.WithError(err).Errorf("[MONGO] ExistByUColumn failed!")
		return false, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return count > 0, nil
}

func (m *mongoMapper[T]) GetByUColumns(column string, values ...string) ([]T, error) {
	var result []T
	if column == "" {
		m.Log().Error("[MONGO] GetByUColumns column is empty!")
		return result, errors.NewWithArgs(code.DBParamInvalid, "column is empty!")
	}
	if len(values) == 0 {
		m.Log().Error("[MONGO] GetByUColumns values is empty!")
		return result, errors.NewWithArgs(code.DBParamInvalid, "values is empty!")
	}

	log := m.Log().WithFields(logrus.Fields{
		"column": column,
		"values": values,
	})
	filter := bson.M{strings.LowerFirst(column): bson.M{"$in": values}}
	result, err := m.Find(filter)
	if err != nil {
		log.WithError(err).Errorf("[MONGO] GetByUColumns failed!")
		return nil, err
	}
	return result, nil
}

func (m *mongoMapper[T]) Create(model *T) error {
	if model == nil {
		m.Log().Error("[MONGO] Create model is nil!")
		return errors.NewWithArgs(code.DBParamInvalid, "model is nil!")
	}

	_ = reflects.SetModelField(model, "CreatedAt", times.UnixMilli())
	_ = reflects.SetModelField(model, "UpdatedAt", times.UnixMilli())

	log := m.Log().WithFields(logrus.Fields{
		"model": model,
	})
	result, err := m.Collection().InsertOne(nil, model)
	if err != nil {
		log.WithError(err).Errorf("[MONGO] Create failed!")
		return errors.NewWithArgs(code.DBAbnormal, err)
	}

	_ = reflects.SetModelField(model, "ID", result.InsertedID.(primitive.ObjectID).Hex())
	return nil
}

func (m *mongoMapper[T]) CreateBatch(models []T) error {
	for index := range models {
		model := &models[index]
		err := m.Create(model)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoMapper[T]) Update(model *T) error {
	if model == nil {
		m.Log().Error("[MONGO] Update model is nil!")
		return errors.NewWithArgs(code.DBParamInvalid, "model is nil!")
	}

	_ = reflects.SetModelField(model, "UpdatedAt", times.UnixMilli())
	log := m.Log().WithFields(logrus.Fields{
		"model": model,
	})
	id, err := utils.GetID(model)
	if err != nil {
		log.Error("[MYSQL] Update model ID err!")
		return errors.NewWithArgs(code.DBParamInvalid, err)
	}
	objectID, _ := primitive.ObjectIDFromHex(id)
	_ = reflects.SetModelField(model, "ID", "")
	_, err = m.Collection().ReplaceOne(nil, bson.M{"_id": objectID}, model)
	_ = reflects.SetModelField(model, "ID", id)
	if err != nil {
		log.WithError(err).Errorf("[MONGO] Update failed!")
		return errors.NewWithArgs(code.DBAbnormal, err)
	}
	return nil
}

func (m *mongoMapper[T]) UpdateBatch(models []T) error {
	for index := range models {
		model := &models[index]
		err := m.Update(model)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoMapper[T]) Save(model *T) error {
	if model == nil {
		m.Log().Error("[MONGO] Save model is nil!")
		return errors.NewWithArgs(code.DBParamInvalid, "model is nil!")
	}
	isNew, err := utils.IsNew(model)
	if err != nil {
		return err
	}
	if isNew {
		return m.Create(model)
	}
	return m.Update(model)
}

func (m *mongoMapper[T]) SaveBatch(models []T) error {
	for index := range models {
		model := &models[index]
		err := m.Save(model)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoMapper[T]) DeleteById(id string) error {
	if id == "" {
		m.Log().Error("[MONGO] DeleteById id is empty!")
		return errors.NewWithArgs(code.DBParamInvalid, "id is empty!")
	}
	log := m.Log().WithFields(logrus.Fields{
		"id": id,
	})
	objectID, _ := primitive.ObjectIDFromHex(id)
	result, err := m.Collection().DeleteOne(nil, bson.M{"_id": objectID})
	if err != nil {
		log.WithError(err).Errorf("[MONGO] DeleteById failed!")
		return errors.NewWithArgs(code.DBAbnormal, err)
	}
	if result.DeletedCount == 0 {
		log.Errorf("[MONGO] DeleteById affect zero!")
		return errors.NewWithArgs(code.DBAffectZero, id)
	}
	return nil
}

func (m *mongoMapper[T]) DeleteByIds(ids []string) error {
	if len(ids) == 0 {
		m.Log().Warn("[MONGO] DeleteByIds ids is empty!")
		return nil
	}

	log := m.Log().WithFields(logrus.Fields{
		"ids": ids,
	})
	filter := bson.M{"_id": bson.M{"$in": slices.Map(ids, func(id string) primitive.ObjectID {
		objectID, _ := primitive.ObjectIDFromHex(id)
		return objectID
	})}}
	_, err := m.Collection().DeleteMany(nil, filter)
	if err != nil {
		log.WithError(err).Errorf("[MONGO] DeleteByIds failed!")
		return errors.NewWithArgs(code.DBAbnormal, err)
	}
	return nil
}

func (m *mongoMapper[T]) Delete(model *T) error {
	if model == nil {
		m.Log().Error("[MONGO] Delete model is nil!")
		return errors.NewWithArgs(code.DBParamInvalid, "model is nil!")
	}
	id, err := utils.GetID(model)
	if err != nil {
		logrus.WithError(err).Error("[MONGO] Delete model ID err!")
		return errors.NewWithArgs(code.DBParamInvalid, err)
	}
	return m.DeleteById(id)
}

func (m *mongoMapper[T]) DeleteBatch(models []T) error {
	if len(models) == 0 {
		m.Log().Warn("[MONGO] DeleteBatch models is empty!")
		return nil
	}
	for index := range models {
		model := &models[index]
		err := m.Delete(model)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoMapper[T]) FindOne(filter interface{}, opts ...*options.FindOneOptions) (T, error) {
	var result T
	if filter == nil {
		m.Log().Error("[MONGO] FindOne filter is nil!")
		return result, errors.NewWithArgs(code.DBParamInvalid, "filter is nil!")
	}
	err := m.Collection().FindOne(nil, filter, opts...).Decode(&result)
	if err != nil {
		if IsNoDocuments(err) {
			return result, errors.NewWithArgs(code.DBNotFound, filter)
		}
		m.Log().WithError(err).Errorf("[MONGO] FindOne failed!")
		return result, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return result, nil
}

func (m *mongoMapper[T]) Find(filter interface{}, opts ...*options.FindOptions) ([]T, error) {
	if filter == nil {
		m.Log().Error("[MONGO] Find filter is nil!")
		return nil, errors.NewWithArgs(code.DBParamInvalid, "filter is nil!")
	}
	var result []T
	cursor, err := m.Collection().Find(nil, filter, opts...)
	if err != nil {
		m.Log().WithError(err).Errorf("[MONGO] find failed!")
		return result, errors.NewWithArgs(code.DBAbnormal, err)
	}
	defer func() {
		if err = cursor.Close(nil); err != nil {
			m.Log().WithError(err).Errorf("[MONGO] find close cursor failed!")
		}
	}()

	for cursor.Next(nil) {
		var item T
		if err = cursor.Decode(&item); err != nil {
			m.Log().WithError(err).Errorf("[MONGO] find decode item failed!")
			return result, errors.NewWithArgs(code.DBAbnormal, err)
		}
		result = append(result, item)
	}
	return result, nil
}

func (m *mongoMapper[T]) Exist(filter interface{}) (bool, error) {
	if filter == nil {
		m.Log().Error("[MONGO] Exist filter is nil!")
		return false, errors.NewWithArgs(code.DBParamInvalid, "filter is nil!")
	}
	count, err := m.Collection().CountDocuments(nil, filter)
	if err != nil {
		m.Log().WithError(err).Errorf("[MONGO] Exist failed!")
		return false, errors.NewWithArgs(code.DBAbnormal, err)
	}
	return count > 0, nil
}
