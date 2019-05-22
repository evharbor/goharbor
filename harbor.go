/*
Package goharbor 是一个简单封装了EVHarbor RESTFULL API的golang包
*/
package goharbor

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
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
	if int64(len(cr.Chunk)) != cr.ChunkSize {
		cr.Ok = false
		cr.CodeText = "应返回的数据长度和实际下载的数据长度不一致"
	}

	return cr, nil
}

// ObjReturn 对象上传或下载结果
type ObjReturn struct {
	Results
	Offset  int64 // 已完成对象上传或下载的偏移量
	ObjSize int64 // 对象大小
}

// IsDone 对象上传或下载是否完成
func (o ObjReturn) IsDone() bool {
	if o.Ok && o.Offset == o.ObjSize {
		return true
	}
	return false
}

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

// DownLoadObject 下载一个对象
// param bucketName: 桶名称
// param objPathName: 桶下全路径对象名称
// param savePath: 下载的对象保存的目录路径
// param saveFilename: 下载对象保存的新文件名，为空字符串，使用对象名称
// param startOffset: 从对象的此偏移量处开始下载
func (client ClientStruct) DownLoadObject(bucketName, objPathName, savePath string, saveFilename string, startOffset int64) (*ObjReturn, error) {
	var offset int64
	var readSize = 1024 * 1024 * 10 //10Mb
	if startOffset < 0 {
		offset = 0
	} else {
		offset = startOffset
	}

	// 目录路径不存在存在则创建
	dirPath := filepath.Clean(savePath)
	if exist, _ := DirExists(dirPath); !exist {
		err := os.MkdirAll(dirPath, os.ModeDir)
		if err != nil {
			return nil, err
		}
	}

	// 保存文件名
	var fileName string
	if saveFilename == "" {
		_, fileName = filepath.Split(objPathName)
	} else {
		if strings.IndexByte(saveFilename, filepath.Separator) >= 0 {
			return nil, errors.New("saveFilename不能包含路径分隔符")
		}
		fileName = saveFilename
	}
	filePathName := filepath.Join(dirPath, fileName)
	saveFile, err := os.Create(filePathName)
	if err != nil {
		return nil, err
	}
	defer saveFile.Close()

	ret := &ObjReturn{ObjSize: -1}
	var retErr error
	for {
		r, err := client.DownloadOneChunk(bucketName, objPathName, offset, readSize)
		if err != nil {
			retErr = err
			break
		}

		if !r.Ok {
			ret.Results = r.Results
			break
		}
		ret.ObjSize = r.ObjSize
		if offset >= ret.ObjSize {
			retErr = errors.New("offset超出了对象大小")
			break
		}

		s, err := saveFile.Seek(offset, os.SEEK_SET)
		if err != nil {
			retErr = err
			break
		}
		if s != offset {
			retErr = errors.New("seek文件偏移量错误")
			break
		}
		writedSize, err := saveFile.Write(r.Chunk)
		if err != nil {
			retErr = err
			break
		}
		offset += int64(writedSize)
		if offset >= ret.ObjSize { // 下载完成
			ret.CodeText = "download ok"
			ret.Ok = true
			break
		}
	}
	ret.Offset = offset
	return ret, retErr
}

// UploadObject 上传一个对象
// param bucketName: 桶名称
// param objPathName: 桶下全路径对象名称
// param fileName: 要上传的文件路径
// param startOffset: 从文件的此偏移量处开始上传
func (client ClientStruct) UploadObject(bucketName, objPathName, fileName string, startOffset int64) (*ObjReturn, error) {
	var offset int64
	var readSize = 1024 * 1024 * 10 //10Mb
	if startOffset < 0 {
		offset = 0
	} else {
		offset = startOffset
	}

	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 文件偏移量设置
	s, err := file.Seek(offset, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	if s != offset {
		return nil, errors.New("seek文件偏移量错误")
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	ret := &ObjReturn{ObjSize: fileInfo.Size()}
	var retErr error
	inputReader := bufio.NewReader(file)
	readSize = 1024 * 1024 * 5    //5Mb
	buf := make([]byte, readSize) //5Mb
	for {
		retSize, err := io.ReadFull(inputReader, buf)
		if (err != nil) && (err != io.ErrUnexpectedEOF) {
			retErr = err
			break
		}
		r, err := client.UploadOneChunk(bucketName, objPathName, offset, buf[0:retSize])
		if err != nil {
			retErr = err
			break
		}
		if !r.Ok {
			ret.Results = *r
			break
		}
		offset += int64(retSize)
		if offset >= ret.ObjSize {
			ret.CodeText = "upload ok"
			ret.Ok = true
			break
		}
	}
	ret.Offset = offset
	return ret, retErr
}
