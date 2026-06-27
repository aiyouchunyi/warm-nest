// Package strings @Author larry
// @Date 2025/2/5 14:05
// @Desc

package strings

import (
	"regexp"
)

const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

var emailRegexp = regexp.MustCompile(emailRegex)

func IsEmail(email string) bool {
	return emailRegexp.MatchString(email)
}
