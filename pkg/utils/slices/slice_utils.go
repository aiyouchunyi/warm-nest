// Package slices @Author larry
// File slice_utils.go
// @Date 2024/5/6 14:13:00
// @Desc
package slices

import "sort"

// ToMap 将切片转换为map
// @param models 切片
// @param keyFun key生成函数
func ToMap[K comparable, V any](models []V, keyFun func(V) K) map[K]V {
	modelsMap := make(map[K]V)
	for _, model := range models {
		modelsMap[keyFun(model)] = model
	}
	return modelsMap
}

func ToGroupMap[K comparable, V any](models []V, keyFun func(V) K) map[K][]V {
	modelsMap := make(map[K][]V)
	for _, model := range models {
		key := keyFun(model)
		modelsMap[key] = append(modelsMap[key], model)
	}
	return modelsMap
}

// ToMapMap 将切片转换为mapmap
func ToMapMap[K, k2 comparable, V any](models []V, keyFun func(V) K, keyFun2 func(V) k2) map[K]map[k2]V {
	modelsMap := make(map[K]map[k2]V)
	for _, model := range models {
		key := keyFun(model)
		if _, ok := modelsMap[key]; !ok {
			modelsMap[key] = make(map[k2]V)
		}
		modelsMap[key][keyFun2(model)] = model
	}
	return modelsMap
}

// Merge  合并两个切片
// @param a 切片a
// @param b 切片b
func Merge[T any](arrays ...[]T) []T {
	var a []T
	for _, arr := range arrays {
		a = append(a, arr...)
	}
	return a
}

// Union 两个切面并集
// @param a 切片a
// @param b 切片b
// @return []T 并集
func Union[T comparable](arrays ...[]T) []T {
	m := make(map[T]struct{})
	for _, arr := range arrays {
		for _, v := range arr {
			m[v] = struct{}{}
		}
	}
	results := make([]T, 0, len(m))
	for k := range m {
		results = append(results, k)
	}
	return results
}

// HasIntersection 两个切片是否有交集
func HasIntersection[T comparable](a, b []T) bool {
	return len(Intersection(a, b)) != 0
}

// Intersection 两个切片交集
// @param a 切片a
// @param b 切片b
// @return []T 交集
func Intersection[T comparable](a, b []T) []T {
	m := make(map[T]struct{})
	for _, v := range a {
		m[v] = struct{}{}
	}
	var results []T
	for _, v := range b {
		if _, ok := m[v]; ok {
			results = append(results, v)
		}
	}
	return results
}

func Contain[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func EmptyContain[T comparable](s []T, e T) bool {
	if s == nil || len(s) == 0 {
		return true
	}
	return Contain(s, e)
}

func ContainV2[T any](s []T, e T, equal func(a, b T) bool) bool {
	for _, v := range s {
		if equal(v, e) {
			return true
		}
	}
	return false
}

func Contains[T comparable](s []T, es ...T) bool {
	for _, e := range es {
		if !Contain(s, e) {
			return false
		}
	}
	return true
}

// Remove 从切片中删除元素
// @param s 切片
// @param es 元素列表
// @return []T 删除元素后的切片
func Remove[T comparable](s []T, es ...T) []T {
	out := make([]T, 0, len(s))
	for _, v := range s {
		if Contain(es, v) {
			continue
		}
		out = append(out, v)
	}
	return out
}

func RemoveAll[T comparable](s []T, es []T) []T {
	return Remove(s, es...)
}

// RemoveV2 从切片中删除元素
// @param s 切片
// @param e 元素
// @param equals 比较函数
// @return []T 删除元素后的切片
func RemoveV2[T any](s []T, e T, equals func(a, b T) bool) []T {
	out := make([]T, 0, len(s))
	for _, v := range s {
		if equals(v, e) {
			continue
		}
		out = append(out, v)
	}
	return out
}

// Map 对切片中的每个元素执行函数
// @param s 切片
// @param f 函数
// @return []R 结果切片
func Map[T any, R any](s []T, f func(T) R) []R {
	results := make([]R, 0, len(s))
	for _, v := range s {
		results = append(results, f(v))
	}
	return results
}

func MapPtr[T any](s []T) []*T {
	results := make([]*T, 0, len(s))
	for i := range s {
		results = append(results, &s[i])
	}
	return results
}

// FlatMap 对切片中的每个元素执行函数，然后将结果合并
// @param s 切片
// @param f 函数
// @return []R 结果切片
func FlatMap[T any, R any](s []T, f func(T) []R) []R {
	var results []R
	for _, v := range s {
		results = append(results, f(v)...)
	}
	return results
}

// Unique 去重
func Unique[T comparable](s []T) []T {
	m := make(map[T]struct{})
	results := make([]T, 0, len(s))
	for _, v := range s {
		if _, ok := m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		results = append(results, v)
	}
	return results
}

func UniqueV2[T any](s []T, equals func(a, b T) bool) []T {
	results := make([]T, 0)
	for _, v := range s {
		unique := true
		for _, r := range results {
			if equals(v, r) {
				unique = false
				break
			}
		}
		if unique {
			results = append(results, v)
		}
	}
	return results
}

// Repeat 获取重复元素列表
func Repeat[T comparable](s []T) []T {
	m := make(map[T]struct{})
	results := make([]T, 0, len(s))
	for _, v := range s {
		if _, ok := m[v]; ok {
			results = append(results, v)
			continue
		}
		m[v] = struct{}{}
	}
	return Unique(results)
}

// Filter 过滤
func Filter[T any](s []T, f func(T) bool) []T {
	var results []T
	for _, v := range s {
		if f(v) {
			results = append(results, v)
		}
	}
	return results
}

// Peek 对切片中的每个元素执行函数
func Peek[T any](s []T, f func(*T)) {
	for i := range s {
		v := &s[i]
		f(v)
	}
}

func Group[T any, K comparable](s []T, f func(T) K) map[K][]T {
	m := make(map[K][]T)
	for _, v := range s {
		k := f(v)
		m[k] = append(m[k], v)
	}
	return m
}

func Find[T any](s []T, f func(T) bool) *T {
	for _, v := range s {
		if f(v) {
			return &v
		}
	}
	return nil
}

func FindOne[T any](s []T, f func(T) bool) (T, bool) {
	for _, v := range s {
		if f(v) {
			return v, true
		}
	}
	var t T
	return t, false
}

func Exist[T any](s []T, f func(T) bool) bool {
	for _, v := range s {
		if f(v) {
			return true
		}
	}
	return false
}

func FindOrDefault[T any](s []T, f func(T) bool, defaultValue T) T {
	if v, ok := FindOne(s, f); ok {
		return v
	}
	return defaultValue
}

func FindOrGet[T any](s []T, f func(T) bool, get func() T) T {
	if v, ok := FindOne(s, f); ok {
		return v
	}
	return get()
}

func Sort[T any](s []T, less func(i, j int) bool) []T {
	sort.Slice(s, less)
	return s
}

// Sub a - b
func Sub[T comparable](a []T, b []T) []T {
	var result []T
	for _, v := range a {
		if !Contain(b, v) {
			result = append(result, v)
		}
	}
	return result
}

func Index[T comparable](s []T, e T) int {
	for i, v := range s {
		if v == e {
			return i
		}
	}
	return -1
}

func IndexV2[T any](s []T, find func(a T) bool) int {
	for i, v := range s {
		if find(v) {
			return i
		}
	}
	return -1
}

func AddIfAbsent[T comparable](s []T, e T) []T {
	if !Contain(s, e) {
		return append(s, e)
	}
	return s
}

func AddIfAbsentV2[T any](s []T, e T, equals func(a, b T) bool) []T {
	if !ContainV2(s, e, equals) {
		return append(s, e)
	}
	return s
}

func Set[T comparable](s []T, e T) []T {
	for i, v := range s {
		if v == e {
			s[i] = e
			return s
		}
	}
	return append(s, e)
}

func SetV2[T any](s []T, e T, equals func(a, b T) bool) []T {
	for i, v := range s {
		if equals(v, e) {
			s[i] = e
			return s
		}
	}
	return append(s, e)
}

func Count[T comparable](s []T, e T) int {
	count := 0
	for _, v := range s {
		if v == e {
			count++
		}
	}
	return count
}

func CountV2[T comparable](s []T, f func(a T) bool) int {
	count := 0
	for _, v := range s {
		if f(v) {
			count++
		}
	}
	return count
}

func First[T any](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}
	return s[0]
}

func Equals[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		if !Contain(b, v) {
			return false
		}
	}
	return true
}

func Done[T any](s []T, f func(T)) {
	for _, v := range s {
		f(v)
	}
}

func Both[T any](s []T, f func(T) bool) bool {
	for _, v := range s {
		if !f(v) {
			return false
		}
	}
	return true
}

func ForEach[T any](s []T, f func(T)) {
	for _, v := range s {
		f(v)
	}
}
