// Package maps @Author larry
// @Date 2024/4/22 14:39:00
// @Desc map 工具类
package maps

// Keys 获取map的key
// @param m map
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values 获取map的value
// @param m map
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// ValuesByKeys 获取map的value
// @param m map
// @param keys key列表
// @return value列表
func ValuesByKeys[K comparable, V any](m map[K]V, keys []K) []V {
	values := make([]V, 0, len(keys))
	for _, k := range keys {
		if v, ok := m[k]; ok {
			values = append(values, v)
		}
	}
	return values
}

// GetOrDefault 获取map的value，如果key不存在则返回默认值
// @param m map
// @param key key
// @param defaultValue 默认值
func GetOrDefault[K comparable, V any](m map[K]V, key K, defaultValue V) V {
	if v, ok := m[key]; ok {
		return v
	}
	return defaultValue
}

func GetDefault(m map[string]string, key string) string {
	return GetOrDefault(m, key, key)
}

// PutAll 将m2的元素放入m
// @param m map
// @param m2 map
// @return map
func PutAll[K comparable, V any](m map[K]V, m2 map[K]V) map[K]V {
	for k, v := range m2 {
		m[k] = v
	}
	return m
}

// Switch map的key和value互换
// @param m map
// @return map
func Switch[K comparable, V comparable](m map[K]V) map[V]K {
	m2 := make(map[V]K)
	for k, v := range m {
		m2[v] = k
	}
	return m2
}

// MapValue map的value进行转换
// @param m map
// @param f 转换函数
// @return map
func MapValue[K comparable, V any, M any](m map[K]V, f func(v V) M) map[K]M {
	m2 := make(map[K]M)
	for k, v := range m {
		m2[k] = f(v)
	}
	return m2
}

// Equal 判断两个map是否相等
// @param a map
// @param b map
// @param eq 判断两个value是否相等的函数
func Equal[K comparable, V any](a, b map[K]V, eq func(v1 V, v2 V) bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		v2, ok := b[k]
		if !ok {
			return false
		}
		if !eq(v, v2) {
			return false
		}
	}
	return true
}

// PutIfAbsent 如果key不存在则放入map
func PutIfAbsent[K comparable, V any](m map[K]V, key K, value V) (V, bool) {
	if v, ok := m[key]; ok {
		return v, false
	}
	m[key] = value
	return value, true
}

func FilterByKeys[K comparable, V any](m map[K]V, keys []K) map[K]V {
	m2 := make(map[K]V)
	for _, k := range keys {
		if v, ok := m[k]; ok {
			m2[k] = v
		}
	}
	return m2
}

func Merge[K comparable, V any](m1, m2 map[K]V) map[K]V {
	m3 := make(map[K]V)
	for k, v := range m1 {
		m3[k] = v
	}
	for k, v := range m2 {
		m3[k] = v
	}
	return m3
}

func Convert[K comparable, V comparable](m map[K]V) map[V]K {
	m2 := make(map[V]K)
	for k, v := range m {
		m2[v] = k
	}
	return m2
}
