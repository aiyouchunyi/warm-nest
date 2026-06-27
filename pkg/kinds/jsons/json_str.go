// Package jsons @Author larry
// @Date 2025/7/15 11:01
// @Desc

package jsons

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JsonStr string

// Value implements the driver.Valuer interface
func (j JsonStr) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JsonStr) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	return json.Unmarshal(bytes, j)
}

func (j JsonStr) ToString() string {
	return string(j)
}
