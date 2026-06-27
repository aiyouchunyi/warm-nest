// Package widgets @Author larry
// @Date 2026/5/27 10:30
// @Desc 索引 DDL helper：基于 GORM Schema 解析字段，原生 ALTER TABLE 拼接，支持单/复合索引，全部幂等
//
// 命名规范：普通索引 i_xxx，唯一索引 u_xxx；与项目 struct tag 规范保持一致。

package widgets

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"warm-nest/pkg/tool/database/mysqls"
	strs "warm-nest/pkg/utils/strings"
)

// AddIndex 创建普通索引；支持单字段或多字段（复合索引）；幂等。
//
// name：索引名（建议 i_xxx）。
// fields：Go 字段名或 db 列名，至少一个；多个 = 复合索引，顺序即索引列顺序。
func AddIndex(model any, name string, fields ...string) error {
	return createIndex(model, name, false, fields)
}

// AddUniqueIndex 创建唯一索引；支持单字段或多字段（复合唯一）；幂等。
//
// name：索引名（建议 u_xxx）。
func AddUniqueIndex(model any, name string, fields ...string) error {
	return createIndex(model, name, true, fields)
}

// DropIndex 删除索引；幂等：索引不存在直接返回 nil。
func DropIndex(model any, name string) error {
	db := mysqls.DB()
	if !db.Migrator().HasIndex(model, name) {
		return nil
	}
	return db.Migrator().DropIndex(model, name)
}

// HasIndex 索引是否存在。
func HasIndex(model any, name string) bool {
	return mysqls.DB().Migrator().HasIndex(model, name)
}

// RenameIndex 重命名索引；幂等：oldName 不存在或 newName 已存在均返回 nil。
func RenameIndex(model any, oldName, newName string) error {
	db := mysqls.DB()
	if !db.Migrator().HasIndex(model, oldName) {
		return nil
	}
	if db.Migrator().HasIndex(model, newName) {
		return nil
	}
	return db.Migrator().RenameIndex(model, oldName, newName)
}

func createIndex(model any, name string, unique bool, fields []string) error {
	if name == "" {
		return fmt.Errorf("index name required")
	}
	if len(fields) == 0 {
		return fmt.Errorf("index %s requires at least one field", name)
	}
	db := mysqls.DB()
	if db.Migrator().HasIndex(model, name) {
		return nil
	}
	tableName, columns, err := resolveColumns(db, model, fields)
	if err != nil {
		return err
	}
	values := []any{clause.Column{Name: name}, clause.Table{Name: tableName}}
	placeholders := make([]string, 0, len(columns))
	for _, c := range columns {
		placeholders = append(placeholders, "?")
		values = append(values, clause.Column{Name: c})
	}
	keyword := "CREATE INDEX"
	if unique {
		keyword = "CREATE UNIQUE INDEX"
	}
	sql := keyword + " ? ON ? (" + strings.Join(placeholders, ", ") + ")"
	return db.Exec(sql, values...).Error
}

// resolveColumns 把若干 Go 字段名/db 列名批量解析为 db 列名 + 表名（一次 Parse）。
// LookUpField 未命中时兜底走 CamelToSnake，与 widget_alter_column.resolveColumnName 语义一致，
// 避免把驼峰字段名直接塞进索引 SQL。
func resolveColumns(db *gorm.DB, model any, fields []string) (tableName string, columns []string, err error) {
	sch, err := parseSchema(db, model)
	if err != nil {
		return
	}
	tableName = sch.Table
	columns = make([]string, 0, len(fields))
	for _, f := range fields {
		if field := sch.LookUpField(f); field != nil {
			columns = append(columns, field.DBName)
			continue
		}
		columns = append(columns, strs.CamelToSnake(f))
	}
	return
}
