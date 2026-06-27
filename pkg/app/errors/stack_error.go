// Package errors @Author larry
// @Date 2024/4/11 15:20:00
// @Desc 栈错误(底层重复错误码使用)
package errors

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type stackError struct {
	err   error
	code  string
	cause error
	*stack
}

func NewWithStack(codeError CodeError, args ...interface{}) error {
	return &stackError{
		err:   NewWithArgs(codeError, args...),
		code:  codeError.Code,
		stack: Callers(),
	}
}

func NewWithStackLog(msg string, codeError CodeError, args ...interface{}) error {
	err := &stackError{
		err:   NewWithArgs(codeError, args...),
		code:  codeError.Code,
		stack: Callers(),
	}
	logrus.Errorf("%s. %+v", msg, err)
	return err
}

func WrapWithStack(err error, codeError CodeError, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &stackError{
		err:   NewWithArgs(codeError, args...),
		code:  codeError.Code,
		cause: err,
		stack: Callers(),
	}
}

func WrapWithStackLog(cause error, msg string, codeError CodeError, args ...interface{}) error {
	if cause == nil {
		return nil
	}
	err := &stackError{
		err:   NewWithArgs(codeError, args...),
		code:  codeError.Code,
		cause: cause,
		stack: Callers(),
	}
	logrus.Errorf("%s. %+v", msg, err)
	return err
}

func (w *stackError) Error() string {
	return w.err.Error()
}

func (w *stackError) Cause() error { return w.cause }

func (w *stackError) Unwrap() error { return w.cause }

func (w *stackError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "error=%s cause=%+v", w.Error(), w.Cause())
			w.stack.Format(s, verb)
			return
		}
		fallthrough
	case 's', 'q':
		fmt.Fprintf(s, "%s", w.Error())
	}
}

func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}
