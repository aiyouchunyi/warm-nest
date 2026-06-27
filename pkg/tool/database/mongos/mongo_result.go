// Package mongos @Author larry
// @Date 2025/4/18 09:43
// @Desc

package mongos

import (
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"

	"warm-nest/pkg/tool/database/code"
)

func IsNoDocuments(err error) bool {
	return err != nil &&
		(errors.Is(err, mongo.ErrNoDocuments) ||
			code.DBNotFound.Is(err) ||
			strings.Contains(err.Error(), "record not found"))
}
