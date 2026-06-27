// Package operator @Author larry
// @Date 2026/1/30 10:09
// @Desc

package impls

import (
	"fmt"

	"gorm.io/gorm"

	"warm-nest/pkg/tool/database/mysqls"
	"warm-nest/pkg/tool/machines/model/config"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/task"
	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/strings"
	"warm-nest/pkg/utils/times"
)

type IMachineRepository[T task.IMachineTask] interface {
	Sync(ctx *context.MachineContext) error                                   // 任务状态同步
	FindTask(taskId string) (T, error)                                        // 获取任务
	FindTasks(taskIds []string) ([]T, error)                                  // 获取任务列表
	MonitorTasks(statuses []string, config config.MachineConfig) ([]T, error) // 获取监控任务列表,监控状态、时间、其他参数
	Lock(ctx *context.MachineContext) (bool, error)                           // 锁定任务
	Unlock(ctx *context.MachineContext) (bool, error)                         // 解锁任务
	ReLock(lockTime int64) (int64, error)                                     // 释放锁,释放数量
}

type MachineRepository[T task.IMachineTask] struct {
	rootDB *gorm.DB
	scopes []func(*gorm.DB) *gorm.DB
}

func NewRepository[T task.IMachineTask]() IMachineRepository[T] {
	return NewRepositoryWithDB[T](mysqls.DB())
}

func NewRepositoryWithDB[T task.IMachineTask](rootDB *gorm.DB, scopes ...func(*gorm.DB) *gorm.DB) IMachineRepository[T] {
	return &MachineRepository[T]{
		rootDB: rootDB,
		scopes: scopes,
	}
}

func (r *MachineRepository[T]) db() *gorm.DB {
	var t T
	db := r.rootDB.Model(&t)
	for _, scope := range r.scopes {
		db = scope(db)
	}
	return db
}

func (r *MachineRepository[T]) taskField() string {
	t := reflects.New[T]()
	return strings.CamelToUnderline(t.GetIdentify().TaskField)
}

// Sync 任务状态同步
func (r *MachineRepository[T]) Sync(ctx *context.MachineContext) error {
	updateColumns := map[string]interface{}{
		"status":       ctx.Status,
		"status_infos": ctx.StatusInfos,
		"runtime":      ctx.Runtime,
	}
	if ctx.Change {
		updateColumns["updated_at"] = times.UnixMilli()
	}
	return r.db().Where(r.taskField()+" = ?", ctx.TaskId).
		UpdateColumns(updateColumns).Error
}

// FindTask 获取任务
func (r *MachineRepository[T]) FindTask(taskId string) (T, error) {
	t := reflects.New[T]()
	err := r.db().Where(r.taskField()+" = ?", taskId).First(&t).Error
	return t, err
}

// FindTasks 获取任务列表
func (r *MachineRepository[T]) FindTasks(taskIds []string) ([]T, error) {
	var tasks []T
	err := r.db().Where(r.taskField()+" IN ?", taskIds).Find(&tasks).Error
	return tasks, err
}

// MonitorTasks 获取监控任务列表
func (r *MachineRepository[T]) MonitorTasks(statuses []string, config config.MachineConfig) ([]T, error) {
	var tasks []T
	err := r.db().Where(fmt.Sprintf("status IN (?) AND updated_at >= ? AND `lock` = false"),
		statuses, times.BeforeMilli(config.MonitorTime)).Find(&tasks).Error
	return tasks, err
}

// Lock 锁定任务
func (r *MachineRepository[T]) Lock(ctx *context.MachineContext) (bool, error) {
	return mysqls.UpdateSucc(r.db().Where(r.taskField()+" = ? AND `lock` = false", ctx.TaskId).
		UpdateColumns(map[string]interface{}{
			"lock":      true,
			"locked_at": times.UnixMilli(),
		}))
}

// Unlock 解锁任务
func (r *MachineRepository[T]) Unlock(ctx *context.MachineContext) (bool, error) {
	return mysqls.UpdateSucc(r.db().Where(r.taskField()+" = ? AND `lock` = true", ctx.TaskId).
		UpdateColumns(map[string]interface{}{
			"lock":      false,
			"locked_at": 0,
		}))
}

// ReLock 释放锁
func (r *MachineRepository[T]) ReLock(lockTime int64) (int64, error) {
	return mysqls.UpdateRows(r.db().Where("`lock` = true AND locked_at > 0 AND locked_at <= ?", times.BeforeMilli(lockTime)).
		UpdateColumns(map[string]interface{}{
			"lock":      false,
			"locked_at": -1,
		}))
}
