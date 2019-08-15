/*
Package goharbor 是一个简单封装了EVHarbor RESTFULL API的golang包
*/
package goharbor

import (
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

// GetConfigs 获取配置信息
func (client ClientStruct) GetConfigs() ConfigStruct {
	return client.API.configs
}

// ObjMetadataReturn 对象或目录元数据返回结果
type ObjMetadataReturn struct {
	Results
	BucketName string         `json:"bucket_name,omitempty"`
	DirPath    string         `json:"dir_path,omitempty"`
	Obj        MetadataStruct `json:"obj,omitempty"`
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

		writedSize, err := saveFile.WriteAt(r.Chunk, offset)
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
	// inputReader := bufio.NewReader(file)
	readSize = 1024 * 1024 * 5    //5Mb
	buf := make([]byte, readSize) //5Mb
	for {
		retSize, err := io.ReadFull(file, buf)
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

// DeleteObject 删除一个对象
// param bucketName: 桶名称
// param objPathName: 桶下全路径对象名称
func (client ClientStruct) DeleteObject(bucketName, objPathName string) (*Results, error) {

	dirPath, objName := CutPathAndName(objPathName)
	resp, err := client.API.DeleteObject(bucketName, dirPath, objName)
	if err != nil {
		return nil, err
	}

	result := ResponseResult(resp)
	if resp.StatusCode != 204 {
		result.Ok = false
		if result.CodeText == "" {
			result.CodeText = "Failed to delete object"
		}
	}
	return result, nil
}

// MoveRenameReturn 移动重命名返回结构体
type MoveRenameReturn struct {
	Results
	BucketName string         `json:"bucket_name,omitempty"`
	DirPath    string         `json:"dir_path,omitempty"`
	Obj        MetadataStruct `json:"obj,omitempty"`
}

// MoveRenameObject 移动或重命名一个对象
// param bucket_name: 桶名称
// param objPathName: 桶下全路径对象名称
// param moveTo: 移动对象到此目录路径，""为不移动, "/"为根目录
// param rename: 重命名对象，，""为不重命名
func (client ClientStruct) MoveRenameObject(bucketName, objPathName, moveTo, rename string) (*MoveRenameReturn, error) {

	dirPath, objName := CutPathAndName(objPathName)
	resp, err := client.API.MoveRenameObject(bucketName, dirPath, objName, moveTo, rename)
	if err != nil {
		return nil, err
	}

	ret := MoveRenameReturn{}
	if resp.StatusCode == 201 {
		ret.Ok = true
		ret.Code = resp.StatusCode
		json.Unmarshal(resp.Bytes(), &ret)
		return &ret, nil
	}

	result2 := ResponseResult(resp)
	ret.Results = *result2
	return &ret, nil
}

// MoveObject 移动一个对象
// param bucket_name: 桶名称
// param objPathName: 桶下全路径对象名称
// param moveTo: 移动对象到此目录路径，""为不移动, "/"为根目录
func (client ClientStruct) MoveObject(bucketName, objPathName, moveTo string) (*MoveRenameReturn, error) {

	return client.MoveRenameObject(bucketName, objPathName, moveTo, "")
}

// RenameObject 重命名一个对象
// param bucket_name: 桶名称
// param objPathName: 桶下全路径对象名称
// param rename: 重命名对象，，""为不重命名
func (client ClientStruct) RenameObject(bucketName, objPathName, rename string) (*MoveRenameReturn, error) {

	return client.MoveRenameObject(bucketName, objPathName, "", rename)
}

// ObjectPermission 对象公有或私有访问权限设置
// param bucket_name: 桶名称
// param objPathName: 桶下全路径对象名称
// param share: 是否分享公开，用于设置对象公有或私有, true(公有)，false(私有)
// param days: 对象公开分享天数(share=true时有效)，0表示永久公开，负数表示不公开
func (client ClientStruct) ObjectPermission(bucketName, objPathName string, share bool, days int) (*Results, error) {

	dirPath, objName := CutPathAndName(objPathName)
	resp, err := client.API.ObjectPermission(bucketName, dirPath, objName, share, days)
	if err != nil {
		return nil, err
	}

	result := ResponseResult(resp)
	if resp.StatusCode != 200 {
		result.Ok = false
		if result.CodeText == "" {
			result.CodeText = "Failed to share object"
		}
	}
	return result, nil
}

// MakeDir 创建一个目录
// param bucketName: 桶名称
// param dirPath: 桶下目录所在路径
// param dirName: 目录名称
func (client ClientStruct) MakeDir(bucketName, dirPath, dirName string) (*Results, error) {

	var dir DirStruct
	slice := []string{dirPath, dirName}
	pathName := buildPath(slice)
	configs := client.GetConfigs()
	dir.Init(bucketName, pathName, configs)
	return dir.MakeDir()
}

// DeleteDir 删除一个空目录
// param bucketName: 桶名称
// param dirPath: 桶下目录路径
func (client ClientStruct) DeleteDir(bucketName, dirPath string) (*Results, error) {

	var dir DirStruct
	configs := client.GetConfigs()
	dir.Init(bucketName, dirPath, configs)
	return dir.DeleteDir()
}

// ListDirOnePage 自定义获取一页目录下的子目录和对象信息
// param bucketName: 桶名称
// param dirPathName: 桶下全路径目录名称
// param offset limit: 自定义从offset偏移量处获取limit条信息；offset和limit大于0时，参数有效；
//  	否则按服务器默认返回数据
func (client ClientStruct) ListDirOnePage(bucketName, dirPathName string, offset, limit int) (*ListDirReturn, error) {
	var dir DirStruct
	configs := client.GetConfigs()
	dir.Init(bucketName, dirPathName, configs)
	return dir.ListDirOnePage(offset, limit)
}

// Dir 获取一个目录结构体实例
func (client ClientStruct) Dir(bucketName, dirPathName string) *DirStruct {
	var dir DirStruct
	configs := client.GetConfigs()
	dir.Init(bucketName, dirPathName, configs)
	return &dir
}

// DirStruct 目录结构体
type DirStruct struct {
	bucketName string
	pathName   string
	path       string
	name       string
	configs    ConfigStruct
	curPage    *ListDirReturn
}

// Init 初始化一个目录
// param bucketName: 同名称
// param dirPathName: 全路径目录名
func (dir *DirStruct) Init(bucketName, dirPathName string, configs ConfigStruct) {
	dir.bucketName = bucketName
	dir.pathName = dirPathName
	dirPath, dirName := CutPathAndName(dirPathName)
	dir.path = dirPath
	dir.name = dirName
	dir.configs = configs
	return
}

// GetBucketName 获取桶名称
func (dir DirStruct) GetBucketName() string {
	return dir.bucketName
}

// GetDirPath 获取目录父目录路径
func (dir DirStruct) GetDirPath() string {
	return dir.path
}

// GetDirName 获取目录名称
func (dir DirStruct) GetDirName() string {
	return dir.name
}

// GetDirPathName 获取全路径目录名
func (dir DirStruct) GetDirPathName() string {
	if dir.pathName == "" {
		dir.pathName = buildPath([]string{dir.path, dir.name})
	}
	return dir.pathName
}

// MakeDir 创建一个目录
// param bucketName: 桶名称
// param dirPath: 桶下目录所在路径
// param dirName: 目录名称
func (dir DirStruct) MakeDir() (*Results, error) {

	API := APIWrapper{configs: dir.configs}
	resp, err := API.MakeDir(dir.GetBucketName(), dir.GetDirPath(), dir.GetDirName())
	if err != nil {
		return nil, err
	}

	result := ResponseResult(resp)
	if resp.StatusCode == 201 {
		return result, nil
	}

	if resp.StatusCode == 400 {
		s := resp.String()
		existing, ok, err := GetValueFromJSON("existing", s)
		if err == nil && ok {
			if existing.(bool) == true {
				result.Ok = true
				result.CodeText = "Directory already exists"
				return result, nil
			}
		}
	}

	result.Ok = false
	if result.CodeText == "" {
		result.CodeText = "Failed to Create directory"
	}

	return result, nil
}

// DeleteDir 删除一个空目录
// param bucketName: 桶名称
// param dirPath: 桶下目录路径
func (dir DirStruct) DeleteDir() (*Results, error) {

	API := APIWrapper{configs: dir.configs}
	resp, err := API.DeleteDir(dir.GetBucketName(), dir.GetDirPath(), dir.GetDirName())
	if err != nil {
		return nil, err
	}

	result := ResponseResult(resp)
	if resp.StatusCode != 204 {
		result.Ok = false
	}
	return result, nil
}

type page struct {
	Current int `json:"current,omitempty"`
	Final   int `json:"final,omitempty"`
}

// ListDirReturn 列举目录返回值类型
type ListDirReturn struct {
	Results
	BucketName string           `json:"bucket_name,omitempty"`
	DirPath    string           `json:"dir_path,omitempty"` // 全路径目录名
	Count      int              `json:"count,omitempty"`    // 目录下子目录和对象总数
	Next       string           `json:"next,omitempty"`     // 下一页数据url
	Previous   string           `json:"previous,omitempty"` // 上一页数据url
	Page       page             `json:"page,omitempty"`
	Files      []MetadataStruct `json:"files"`
}

// HasNext 是否有下一页
func (ldr ListDirReturn) HasNext() bool {
	return ldr.Next != ""
}

// HasPrevious 是否有上一页
func (ldr ListDirReturn) HasPrevious() bool {
	return ldr.Previous != ""
}

// PreviousURL 上一页url
func (ldr ListDirReturn) PreviousURL() string {
	return ldr.Previous
}

// NextURL 下一页url
func (ldr ListDirReturn) NextURL() string {
	return ldr.Next
}

// CurPageNum 当前页码
func (ldr ListDirReturn) CurPageNum() int {
	return ldr.Page.Current
}

// FinalPageNum 最后一页页码
func (ldr ListDirReturn) FinalPageNum() int {
	return ldr.Page.Final
}

// ListDirOnePage 自定义获取一页目录下的子目录和对象信息
// param offset limit: 自定义从offset偏移量处获取limit条信息；offset和limit大于0时，参数有效；
//  	否则按服务器默认返回数据
func (dir DirStruct) ListDirOnePage(offset, limit int) (*ListDirReturn, error) {

	API := APIWrapper{configs: dir.configs}
	resp, err := API.ListDirOnePage(dir.GetBucketName(), dir.GetDirPath(), dir.GetDirName(), offset, limit)
	if err != nil {
		return nil, err
	}

	return dir.buildListDirReturn(resp)
}

// buildListDirReturn 列举目录返回值构建
func (dir DirStruct) buildListDirReturn(resp *grequests.Response) (*ListDirReturn, error) {

	ret := ListDirReturn{}
	if resp.StatusCode == 200 {
		err2 := json.Unmarshal(resp.Bytes(), &ret)
		if err2 != nil {
			return nil, err2
		}
		ret.Ok = true
		ret.Code = resp.StatusCode
		return &ret, nil
	}

	result := ResponseResult(resp)
	if result.CodeText == "" {
		result.CodeText = "获取信息失败"
	}
	ret.Results = *result
	return &ret, nil
}

// ListFirstPage 列举目录下子目录和对象信息第一页数据
// param numPerPage: 每页数据量
func (dir *DirStruct) ListFirstPage(numPerPage int) (*ListDirReturn, error) {
	r, err := dir.ListDirOnePage(0, numPerPage)
	if err == nil && r.Ok {
		dir.curPage = r
	}

	return r, err
}

// NextPage 获取下一页数据
// 如果希望自定义每页返回数据量，需要先调用ListFirstPage()获取指定数量的第一页数据, 后可通过NextPage()循环获取下一页数据
func (dir *DirStruct) NextPage() (*ListDirReturn, error) {
	if dir.curPage == nil {
		return dir.ListFirstPage(0)
	}
	if !dir.curPage.HasNext() {
		ret := ListDirReturn{}
		ret.Ok = false
		ret.CodeText = "没有下一页了"
		return &ret, nil
	}
	url := dir.curPage.NextURL()

	API := APIWrapper{configs: dir.configs}
	r, err := API.ListDirOnePageByURL(url)
	if err != nil {
		return nil, err
	}

	ret, err := dir.buildListDirReturn(r)
	if err == nil && r.Ok {
		dir.curPage = ret
	}
	return ret, err
}

// PreviousPage 获取下一页数据
func (dir *DirStruct) PreviousPage() (*ListDirReturn, error) {

	if (dir.curPage == nil) || (!dir.curPage.HasPrevious()) {
		ret := ListDirReturn{}
		ret.Ok = false
		ret.CodeText = "没有上一页"
		return &ret, nil
	}
	url := dir.curPage.PreviousURL()

	API := APIWrapper{configs: dir.configs}
	r, err := API.ListDirOnePageByURL(url)
	if err != nil {
		return nil, err
	}

	ret, err := dir.buildListDirReturn(r)
	if err == nil && r.Ok {
		dir.curPage = ret
	}
	return ret, err
}
