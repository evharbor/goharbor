/*
Package goharbor 是一个简单封装了EVHarbor RESTFULL API的golang包
*/
package goharbor

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"goharbor/grequests"
)

// ConfigKeyType 配置map的key数据类型
type ConfigKeyType uint16

const (
	//HTTP scheme
	HTTP string = "http"
	//HTTPS scheme
	HTTPS string = "https"
	// VERSION 配置选项key
	VERSION ConfigKeyType = iota // 0
	// SCHEME 配置选项key
	SCHEME ConfigKeyType = iota
	// HOST 配置选项key
	HOST ConfigKeyType = iota
	// ACCESSKEY 配置选项key
	ACCESSKEY ConfigKeyType = iota
	// SECRETKEY 配置选项key
	SECRETKEY ConfigKeyType = iota
)

//ConfigStruct 是一个配置相关结构体
type ConfigStruct struct {
	Version           string
	Scheme            string
	Host              string
	Accesskey         string
	Secretkey         string
	APIObjPrefix      string
	APIDirPrefix      string
	APIBucketPrefix   string
	APIMovePrefix     string
	APIMetadataPrefix string
	APIStatsPrefix    string
}

//DefaultConfigs is default config
var defaultConfigs = ConfigStruct{
	Version:           "v1",
	Scheme:            HTTPS,
	Host:              "obs.casearth.cn",
	APIObjPrefix:      "obj",
	APIDirPrefix:      "dir",
	APIBucketPrefix:   "buckets",
	APIMovePrefix:     "move",
	APIMetadataPrefix: "metadata",
	APIStatsPrefix:    "stats",
}

//GetDefaultConfig return default configs
func GetDefaultConfig() ConfigStruct {
	return defaultConfigs
}

// InitConfig 初始化并返回一个ConfigStruct对象
func InitConfig(c map[ConfigKeyType]string) (config ConfigStruct, err error) {
	config = GetDefaultConfig()

	for key, value := range c {
		switch key {
		case VERSION:
			config.Version = value
		case SCHEME:
			config.Scheme = value
		case HOST:
			config.Host = value
		case ACCESSKEY:
			config.Accesskey = value
		case SECRETKEY:
			config.Secretkey = value
		}
	}

	if config.Accesskey == "" || config.Secretkey == "" {
		err = errors.New("Valid values must be configured for both ACCESSKEY and SECRETKEY")
		return
	}
	err = nil
	return
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

// ObjChunk 对象分片
type ObjChunk struct {
	ChunkOffset int64
	ChunkSize   int64
	ObjSize     int64
	Chunk       []byte
}

// Results 请求结果类型
type Results struct {
	Ok       bool   `json:"ok,omitempty"`
	CodeText string `json:"code_text,omitempty"`
	Code     int    `json:"code,omitempty"`
}

// ResultsDetail json results
func (r Results) ResultsDetail() string {

	msg, _ := json.Marshal(r)
	return string(msg)
}

// Error 错误信息 实现error interface
func (r Results) Error() string {

	return r.CodeText
}

// Text 结果描述信息
func (r Results) Text() string {

	return r.CodeText
}

// HTTPCode 状态码
func (r Results) HTTPCode() int {

	return r.Code
}

func parseResponseMsg(e *Results, r *grequests.Response) {

	data := r.Bytes()
	mapData, err := JSON2map([]byte(data))
	if err != nil {
		e.CodeText = string(data)
		return
	}

	keys := [...]string{"code_text", "detail", "non_field_errors"}
	for _, k := range keys {
		val, ok := mapData[k]
		if ok {
			var msg interface{}

			switch val.(type) {
			case string:
				msg = val
			case []interface{}:
				msg = val.([]interface{})[0].(interface{})
			default:
				msg = val
			}

			text, err := json.Marshal(msg)
			if err == nil {
				e.CodeText = strings.Trim(string(text), "\"")
			}

			break
		}
	}
	return
}

// ResponseResult 获取请求结果
func ResponseResult(resp *grequests.Response) *Results {

	e := &Results{
		Ok:   resp.Ok,
		Code: resp.StatusCode,
	}

	ct, ok := resp.Header["Content-Type"]
	if ok && strings.HasPrefix(ct[0], "application/json") {
		parseResponseMsg(e, resp)
		return e
	}

	return e
}

// ChunkReturn 对象分片上传下载相关调用返回结果
type ChunkReturn struct {
	Results
	ObjChunk
}

//ClientStruct 是一个访问EVHarbor的客户端
type ClientStruct struct {
	API APIWrapper
}

// InitClient 初始化一个client
func InitClient(configs ConfigStruct) ClientStruct {
	client := ClientStruct{
		API: APIWrapper{configs: configs},
	}
	return client
}

// ObjMetadataReturn 对象或目录元数据返回结果
type ObjMetadataReturn struct {
	Results
	Data MetadataStruct `json:"data"`
}

// GetMetadata 获取元数据
// param bucketName: 桶名称
// param pathName: 桶下路径
func (client ClientStruct) GetMetadata(bucketName, pathName string) (*ObjMetadataReturn, error) {

	resp, err := client.API.GetMetadata(bucketName, pathName)
	if err != nil {
		return nil, err
	}

	ret := ObjMetadataReturn{}
	if resp.StatusCode == 200 {
		err2 := json.Unmarshal(resp.Bytes(), &ret)
		if err2 != nil {
			result := ResponseResult(resp)
			ret.Results = *result
		} else {
			ret.Ok = true
			ret.Code = resp.StatusCode
		}
		return &ret, nil
	}

	result2 := ResponseResult(resp)
	ret.Results = *result2
	return &ret, nil
}

// UploadOneChunk 上传一个对象数据块
// param bucketName: 桶名称
// param objPathName: 桶下全路径对象名称
// param offset: 数据块在对象中的字节偏移量
// param chunk: 数据块
func (client ClientStruct) UploadOneChunk(bucketName, objPathName string, offset int64, chunk []byte) (*Results, error) {

	dirPath, objName := CutPathAndName(objPathName)

	resp, err := client.API.UploadOneChunk(bucketName, dirPath, objName, offset, chunk)
	if err != nil {
		return nil, err
	}

	result := ResponseResult(resp)
	if resp.StatusCode == 200 {
		if result.CodeText == "" {
			result.CodeText = "Successful to upload an chunk of object"
		}
	} else {
		result.Ok = false
		if result.CodeText == "" {
			result.CodeText = "Failed to upload an chunk of object"
		}
	}

	return result, nil
}

// DownloadOneChunk 下载一个对象数据块
// param bucketName: 桶名称
// param objPathName: 桶下全路径对象名称
// param offset: 数据块在对象中的字节偏移量
// param chunk: 数据块
func (client ClientStruct) DownloadOneChunk(bucketName, objPathName string, offset int64, size int) (*ChunkReturn, error) {

	dirPath, objName := CutPathAndName(objPathName)

	resp, err := client.API.DownloadOneChunk(bucketName, dirPath, objName, offset, size)
	if err != nil {
		return nil, err
	}

	result := ResponseResult(resp)
	if resp.StatusCode != 200 {
		result.Ok = false
		if result.CodeText == "" {
			result.CodeText = "Failed to download an chunk of object"
		}
		return &ChunkReturn{Results: *result}, nil
	}

	cr := &ChunkReturn{Results: *result}
	cr.ChunkOffset = offset
	if cr.CodeText == "" {
		cr.CodeText = "Download successfull"
	}

	if val, ok := resp.Header["Evob_chunk_size"]; ok && len(val) > 0 {
		v, err := strconv.ParseInt(val[0], 10, 64)
		if err == nil {
			cr.ChunkSize = v
		}
	}

	if val, ok := resp.Header["Evob_obj_size"]; ok && len(val) > 0 {
		v, err := strconv.ParseInt(val[0], 10, 64)
		if err == nil {
			cr.ObjSize = v
		}
	}
	cr.Chunk = resp.Bytes()

	return cr, nil
}
