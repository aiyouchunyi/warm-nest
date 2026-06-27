// Package utils @Author larry
// @Date 2025/12/22 11:11
// @Desc

package loader

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml"

	"warm-nest/pkg/tool/caches"
	"warm-nest/pkg/utils/times"
)

// LoadAll loads a struct of type V from a TOML file given the filename and the item key.
func LoadAll[V any](filename string) (V, error) {
	var result V
	tree, err := LoadTomlTree(filename)
	if err != nil {
		return result, err
	}
	if err2 := tree.Unmarshal(&result); err2 != nil {
		return result, fmt.Errorf("toml %s unmarshal failed: %s", filename, err2)
	}
	return result, nil
}

// LoadAllTo loads a struct of type V from a TOML file into the provided cfg pointer.
func LoadAllTo[V any](filename string, cfg *V) error {
	tree, err := LoadTomlTree(filename)
	if err != nil {
		return err
	}
	if err2 := tree.Unmarshal(cfg); err2 != nil {
		return fmt.Errorf("toml %s unmarshal failed: %s", filename, err2)
	}
	return nil
}

// Load loads a struct of type V from a TOML file given the filename and the item key.
func Load[V any](filename string, path string) (V, error) {
	var result V
	err := LoadTo(filename, path, &result)
	return result, err
}

// LoadTo loads a struct of type V from a TOML file given the filename and the item key.
func LoadTo[V any](filename string, path string, cfg *V) error {
	tree, err := LoadTomlTree(filename)
	if err != nil {
		return err
	}
	item := tree.Get(path)
	if itemTree, ok := item.(*toml.Tree); ok {
		if err2 := itemTree.Unmarshal(cfg); err2 != nil {
			return fmt.Errorf("toml %s path %s unmarshal failed: %s", filename, path, err2)
		}
		return nil
	}
	return nil
}

// LoadMap loads a map[string]V from a TOML file given the filename and the item key.
func LoadMap[V any](filename string, path string) (map[string]V, error) {
	tree, err := LoadTomlTree(filename)
	if err != nil {
		return nil, err
	}
	itemCfg := tree.Get(path)
	result := make(map[string]V)
	if externalMap, ok := itemCfg.(*toml.Tree); ok {
		for key, value := range externalMap.ToMap() {
			if v, ok2 := value.(V); ok2 {
				result[key] = v
			} else {
				return nil, fmt.Errorf("toml %s path %s key %s type assertion failed", filename, path, key)
			}
		}
		return result, nil
	}
	return result, nil
}

// LoadArray loads an array of V from a TOML file given the filename and the item key.
func LoadArray[V any](filename, path string) ([]V, error) {
	tree, err := LoadTomlTree(filename)
	if err != nil {
		return nil, err
	}
	arr, ok := tree.Get(path).([]*toml.Tree)
	if !ok {
		return nil, fmt.Errorf("toml %s path %s not found or not array", filename, path)
	}
	var result []V
	for _, elem := range arr {
		var v V
		if err2 := elem.Unmarshal(&v); err2 != nil {
			return nil, fmt.Errorf("toml %s path %s unmarshal failed: %s", filename, path, err2)
		}
		result = append(result, v)
	}
	return result, nil
}

// LoadTomlTree loads a TOML file and returns its toml.Tree representation.
func LoadTomlTree(filename string) (*toml.Tree, error) {
	cacheKey := caches.CacheKey("config", filename)
	return caches.GetOrLoad[*toml.Tree](cacheKey, times.FiveMinuteInSec, func() (interface{}, error) {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("read file %s failed: %s", filename, err)
		}
		tree, err := toml.LoadBytes(data)
		if err != nil {
			return nil, fmt.Errorf("toml file %s load failed: %s", filename, err)
		}
		return tree, nil
	})
}
