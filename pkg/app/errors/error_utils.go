// Package errors @Author larry
// @Date 2024/12/19 10:28
// @Desc

package errors

import (
	"errors"
	"reflect"
)

func NilDo(err error, f func() error) error {
	if err == nil {
		return f()
	}
	return err
}

func NilDoResp[T any](err error, f func() (T, error)) (T, error) {
	if err == nil {
		return f()
	}
	var t T
	return t, err
}

func NilDoNil(err error, f func()) {
	if err == nil {
		f()
	}
}

func ErrMsg(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func Equal(e1 error, e2 error) bool {
	if e1 == nil && e2 == nil {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	var ce1 *CodeError
	var ce2 *CodeError
	if errors.As(e1, &ce1) && errors.As(e2, &ce2) {
		if ce1.Code == ce2.Code {
			return true
		}
	}
	return e1.Error() == e2.Error()
}

func Union(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func As(err error, target interface{}) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		panic("errors: target must be a non-nil pointer")
	}

	targetType := val.Elem().Type()
	for err != nil {
		if reflect.TypeOf(err).AssignableTo(targetType) {
			val.Elem().Set(reflect.ValueOf(err))
			return true
		}

		if x, ok := err.(interface{ Unwrap() error }); ok {
			err = x.Unwrap()
		} else {
			break
		}
	}
	return false
}

func Code(err error) string {
	if err == nil {
		return ""
	}
	var ce *CodeError
	if errors.As(err, &ce) {
		return ce.Code
	}
	return ""
}

func Msg(err error) string {
	if err == nil {
		return ""
	}
	var ce *CodeError
	if errors.As(err, &ce) {
		return ce.Msg
	}
	return err.Error()
}
