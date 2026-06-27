// Package mongos @Author larry
// @Date 2025/4/18 10:17
// @Desc

package widget

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateIndex 创建索引的函数
func CreateIndex(collection *mongo.Collection, model any) error {
	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("CreateIndex: type %T is not a struct", model)
	}

	compoundIndexes := make(map[string]struct {
		Keys   bson.D
		Unique bool
	})

	// 遍历结构体字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		bsonTag := field.Tag.Get("bson")
		if bsonTag == "" || bsonTag == "-" {
			continue
		}
		bsonField := strings.Split(bsonTag, ",")[0]

		// 解析 mongo 标签
		if tag, ok := field.Tag.Lookup("gorm"); ok {
			indexName, unique := parseMongoTag(tag)
			if indexName != "" {
				if _, exists := compoundIndexes[indexName]; !exists {
					compoundIndexes[indexName] = struct {
						Keys   bson.D
						Unique bool
					}{Keys: bson.D{}, Unique: unique}
				}
				compoundIndexes[indexName] = struct {
					Keys   bson.D
					Unique bool
				}{
					Keys:   append(compoundIndexes[indexName].Keys, bson.E{Key: bsonField, Value: 1}),
					Unique: compoundIndexes[indexName].Unique || unique,
				}
			}
		}
	}

	// 创建索引
	for name, index := range compoundIndexes {
		indexModel := mongo.IndexModel{
			Keys:    index.Keys,
			Options: options.Index().SetName(name).SetUnique(index.Unique),
		}
		if _, err := collection.Indexes().CreateOne(context.TODO(), indexModel); err != nil {
			return err
		}
		log.Printf("Created index: %s (unique: %v)\n", name, index.Unique)
	}
	return nil
}

// 解析 mongo 标签
func parseMongoTag(tag string) (indexName string, unique bool) {
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "uniqueIndex:") {
			return strings.TrimPrefix(part, "uniqueIndex:"), true
		}
		if strings.HasPrefix(part, "index:") {
			return strings.TrimPrefix(part, "index:"), false
		}
	}
	return "", false
}
