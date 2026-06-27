// Package widgets @Author larry
// @Date 2026/5/27 10:00
// @Desc 列级 DDL helper：基于 GORM Schema 解析渲染列定义，支持 Go 字段名和 db 列名两种入参，全部幂等

package widgets

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"warm-nest/pkg/tool/database/mysqls"
	strs "warm-nest/pkg/utils/strings"
)

// AddColumn 加列；自动从 model 字段顺序推导 AFTER 锚点 —— 取目标字段在 model 中的前一个**已落库**字段；
// 推导失败（首位 / 前序字段全未落库）报错，由调用方改用 AddColumnFirst 或 AddColumnAfter 显式指定。
//
// column：Go 字段名（如 "BatchId"）或 db 列名（如 "batch_id"），由 GORM Schema.LookUpField 自动识别。
//
// 末尾追加请用 AddColumnAfter(model, column, "")，与本接口语义解耦。
func AddColumn(model any, column string) error {
	db := mysqls.DB()
	sch, err := parseSchema(db, model)
	if err != nil {
		return err
	}
	target := sch.LookUpField(column)
	if target == nil {
		return fmt.Errorf("model field not found: %s", column)
	}
	afterDB, err := lookupAnchorBefore(model, sch, target)
	if err != nil {
		return err
	}
	return AddColumnAfter(model, column, afterDB)
}

// lookupAnchorBefore 在 model 字段顺序中向前回溯，找 target 之前**已落库**的最近字段 DBName。
// 跳过 IgnoreMigration、DBName 为空（关联/嵌入字段无对应列）。
func lookupAnchorBefore(model any, sch *schema.Schema, target *schema.Field) (string, error) {
	const hint = "; use AddColumnFirst or AddColumnAfter explicitly"
	db := mysqls.DB()
	idx := -1
	for i, f := range sch.Fields {
		if f == target {
			idx = i
			break
		}
	}
	if idx <= 0 {
		return "", fmt.Errorf("AddColumn: cannot derive AFTER anchor for %s (target is first field)"+hint, target.DBName)
	}
	for i := idx - 1; i >= 0; i-- {
		f := sch.Fields[i]
		if f.IgnoreMigration || f.DBName == "" {
			continue
		}
		if db.Migrator().HasColumn(model, f.DBName) {
			return f.DBName, nil
		}
	}
	return "", fmt.Errorf("AddColumn: cannot derive AFTER anchor for %s (no preceding field landed in table)"+hint, target.DBName)
}

// AddColumnAfter 在 afterColumn 之后追加列；afterColumn 为空等价末尾追加；幂等。
//
// column：同 AddColumn。afterColumn：Go 字段名或 db 列名，为空时不带 AFTER 子句。
func AddColumnAfter(model any, column, afterColumn string) error {
	db := mysqls.DB()
	tableName, dbName, fullType, err := resolveField(db, model, column)
	if err != nil {
		return err
	}
	if db.Migrator().HasColumn(model, dbName) {
		return nil
	}
	if afterColumn == "" {
		return db.Exec("ALTER TABLE ? ADD ? ?",
			clause.Table{Name: tableName}, clause.Column{Name: dbName}, fullType,
		).Error
	}
	afterDB, err := resolveColumnName(db, model, afterColumn)
	if err != nil {
		return err
	}
	return db.Exec("ALTER TABLE ? ADD ? ? AFTER ?",
		clause.Table{Name: tableName}, clause.Column{Name: dbName}, fullType, clause.Column{Name: afterDB},
	).Error
}

// AddColumnFirst 在表首插入列；幂等。
func AddColumnFirst(model any, column string) error {
	db := mysqls.DB()
	tableName, dbName, fullType, err := resolveField(db, model, column)
	if err != nil {
		return err
	}
	if db.Migrator().HasColumn(model, dbName) {
		return nil
	}
	return db.Exec("ALTER TABLE ? ADD ? ? FIRST",
		clause.Table{Name: tableName}, clause.Column{Name: dbName}, fullType,
	).Error
}

// AddColumnsAfter 在 afterColumn 之后按顺序追加多列；链式，columns[0] 紧随 afterColumn。
// afterColumn 为空时按顺序追加到表末尾（首列等价 AddColumn）。
// 已存在的列会跳过；锚点 after 为已存在则正常使用，新列会插在该锚点之后。
func AddColumnsAfter(model any, columns []string, afterColumn string) error {
	prev := afterColumn
	for _, col := range columns {
		if err := AddColumnAfter(model, col, prev); err != nil {
			return err
		}
		prev = col
	}
	return nil
}

// DropColumn 删除列；幂等：列不存在直接返回 nil。
func DropColumn(model any, column string) error {
	db := mysqls.DB()
	dbName, err := resolveColumnName(db, model, column)
	if err != nil {
		return err
	}
	if !db.Migrator().HasColumn(model, dbName) {
		return nil
	}
	return db.Migrator().DropColumn(model, dbName)
}

// RenameColumn 重命名列；幂等：oldName 不存在或 newName 已存在均返回 nil。
// oldName/newName 支持 Go 字段名或 db 列名（基于 model Schema 解析）。
func RenameColumn(model any, oldName, newName string) error {
	db := mysqls.DB()
	oldDB, err := resolveColumnName(db, model, oldName)
	if err != nil {
		return err
	}
	newDB, err := resolveColumnName(db, model, newName)
	if err != nil {
		return err
	}
	if !db.Migrator().HasColumn(model, oldDB) {
		return nil
	}
	if db.Migrator().HasColumn(model, newDB) {
		return nil
	}
	return db.Migrator().RenameColumn(model, oldDB, newDB)
}

// ModifyColumn 按 model 当前 tag 定义重新定义列（不改名）；列不存在则报错。
// 用于 model 字段类型/comment 变更后同步到表结构。
func ModifyColumn(model any, column string) error {
	db := mysqls.DB()
	dbName, err := resolveColumnName(db, model, column)
	if err != nil {
		return err
	}
	if !db.Migrator().HasColumn(model, dbName) {
		return fmt.Errorf("ModifyColumn: column not exist: %s", dbName)
	}
	return db.Migrator().AlterColumn(model, dbName)
}

// parseSchema 包内统一的 model schema 解析入口；错误信息统一带 "parse model failed" 前缀。
func parseSchema(db *gorm.DB, model any) (*schema.Schema, error) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return nil, fmt.Errorf("parse model failed: %w", err)
	}
	return stmt.Schema, nil
}

// resolveField 解析 model + 列输入 → 表名、db 列名、完整列定义（含 size/comment/default）。
// 列输入支持 Go 字段名和 db 列名；底层走 GORM Schema.LookUpField（同时匹配 Name 和 DBName）。
func resolveField(db *gorm.DB, model any, column string) (tableName, dbName string, fullType clause.Expr, err error) {
	sch, err := parseSchema(db, model)
	if err != nil {
		return
	}
	field := sch.LookUpField(column)
	if field == nil {
		err = fmt.Errorf("model field not found: %s", column)
		return
	}
	tableName = sch.Table
	dbName = field.DBName
	fullType = db.Migrator().FullDataTypeOf(field)
	return
}

// resolveColumnName 仅解析输入到 db 列名，不渲染列定义。
// LookUpField 命中（model 仍存在该字段）走 Schema 给出的权威 DBName；
// 未命中（典型场景：DropColumn 时 model 字段已删）兜底走 CamelToSnake 把驼峰转下划线，
// 避免直接把 Go 字段名塞进 SQL 导致 column not exist。
func resolveColumnName(db *gorm.DB, model any, column string) (string, error) {
	sch, err := parseSchema(db, model)
	if err != nil {
		return "", err
	}
	if field := sch.LookUpField(column); field != nil {
		return field.DBName, nil
	}
	return strs.CamelToSnake(column), nil
}
