package goharbor

import (
	"encoding/json"
	"os"
	"strings"
)

// PathExists 文件或目录是否存在
// return:
// 		true and nil,说明文件或文件夹存在
//		false and nil, 不存在
// 		error != nil ,则不确定是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// DirExists 目录是否存在
// return:
// 		true and nil,文件夹存在
//		false and nil, 不存在
// 		error != nil ,则不确定是否存在
func DirExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.Mode().IsDir() {
			return true, nil
		}
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CutPathAndName 切割一个路径，return 父路经和文件目录名称
func CutPathAndName(pathName string) (string, string) {

	i := strings.LastIndex(pathName, "/")
	if i >= 0 {
		return pathName[:i], pathName[i+1:]
	}
	return "", pathName
}

// JSON2map json转换为map
func JSON2map(data []byte) (map[string]interface{}, error) {

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetValueFromJSON 从json中获取指定键的值
func GetValueFromJSON(key, json string) (value interface{}, ok bool, err error) {
	mapData, e := JSON2map([]byte(json))
	if e != nil {
		err = e
		ok = false
		value = nil
		return
	}
	value, ok = mapData[key]
	err = nil
	return
}
