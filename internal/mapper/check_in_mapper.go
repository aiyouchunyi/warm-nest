// Package mapper @Author larry
// @Date 2026/06/15
// @Desc 打卡记录映射器

package mapper

import (
	"fmt"
	"sync"

	"warm-nest/pkg/tool/database/mysqls"

	"warm-nest/internal/model"
)

// CheckInMapper 打卡记录映射器
type CheckInMapper struct {
	mysqls.MysqlMapper[model.CheckIn]
}

var checkInMapper *CheckInMapper
var checkInOnce sync.Once

// GetCheckInMapper 获取打卡记录映射器单例
func GetCheckInMapper() *CheckInMapper {
	checkInOnce.Do(func() {
		checkInMapper = &CheckInMapper{
			MysqlMapper: mysqls.NewMapper[model.CheckIn](),
		}
	})
	return checkInMapper
}

// GetByUserDate 查某人某自然日的打卡（幂等判重，找不到返回 nil,nil）
func (m *CheckInMapper) GetByUserDate(elderUserId, checkInDate string) (*model.CheckIn, error) {
	var c model.CheckIn
	err := m.Model().
		Where("elder_user_id = ? AND check_in_date = ?", elderUserId, checkInDate).
		First(&c).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get checkin %s@%s: %w", elderUserId, checkInDate, err)
	}
	return &c, nil
}

// GetByCheckInId 按业务打卡ID查单条记录（单日详情用，找不到返回 nil,nil）
func (m *CheckInMapper) GetByCheckInId(checkInId string) (*model.CheckIn, error) {
	var c model.CheckIn
	err := m.Model().Where("check_in_id = ?", checkInId).First(&c).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get checkin by id %s: %w", checkInId, err)
	}
	return &c, nil
}

// ListByCheckInIds 按一批打卡ID查记录（消息列表带照片用，批量 IN 避免 N+1）
func (m *CheckInMapper) ListByCheckInIds(checkInIds []string) ([]model.CheckIn, error) {
	if len(checkInIds) == 0 {
		return nil, nil
	}
	var list []model.CheckIn
	err := m.Model().Where("check_in_id IN ?", checkInIds).Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list checkin by ids: %w", err)
	}
	return list, nil
}

// ListDatesByUser 查某人全部打卡日期升序（连续/累计天数算法用，只取 check_in_date 列）
func (m *CheckInMapper) ListDatesByUser(elderUserId string) ([]string, error) {
	var dates []string
	err := m.Model().
		Where("elder_user_id = ?", elderUserId).
		Order("check_in_date ASC").
		Pluck("check_in_date", &dates).Error
	if err != nil {
		return nil, fmt.Errorf("list checkin dates by %s: %w", elderUserId, err)
	}
	return dates, nil
}

// ListByMonth 查某人某月(YYYY-MM)的打卡记录（日历/月份详情用）
func (m *CheckInMapper) ListByMonth(elderUserId, yearMonth string) ([]model.CheckIn, error) {
	var list []model.CheckIn
	err := m.Model().
		Where("elder_user_id = ? AND check_in_date LIKE ?", elderUserId, yearMonth+"%").
		Order("check_in_date ASC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list checkin by month %s@%s: %w", elderUserId, yearMonth, err)
	}
	return list, nil
}

// CountByUser 查某人累计打卡天数
func (m *CheckInMapper) CountByUser(elderUserId string) (int64, error) {
	var cnt int64
	err := m.Model().Where("elder_user_id = ?", elderUserId).Count(&cnt).Error
	if err != nil {
		return 0, fmt.Errorf("count checkin by %s: %w", elderUserId, err)
	}
	return cnt, nil
}
