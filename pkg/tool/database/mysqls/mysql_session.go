// Package mysqls @Author Larry
// @Date 2024/10/10 14:24
// @Desc

package mysqls

type MysqlSession struct {
	silent bool // 是否静默模式
}

type Option func(s *MysqlSession)

func Silent(silent bool) Option {
	return func(s *MysqlSession) {
		s.silent = silent
	}
}

func NewSession(opts ...Option) *MysqlSession {
	s := &MysqlSession{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
