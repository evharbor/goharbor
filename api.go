package goharbor

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"

	"goharbor/grequests"
)

// MetadataStruct 对象元数据
type MetadataStruct struct {
	PathName         string `json:"na"`                     //全路径
	Name             string `json:"name"`                   //对象或目录名称
	FileOrDir        bool   `json:"fod"`                    //对象（true），目录（false）
	ParentDirID      uint64 `json:"did"`                    //父目录id
	Size             uint64 `json:"si"`                     //大小，byte
	UploadTime       string `json:"ult"`                    // 上传时间
	UpdateTime       string `json:"upt,omitempty"`          // 最后修改时间
	DownloadCount    uint32 `json:"dlc"`                    //下载次数
	DownloadURL      string `json:"download_url,omitempty"` // 下载url
	AccessPermission string `json:"access_permission"`      // 访问权限
}

// APIWrapper EVHarbor API wrapper
type APIWrapper struct {
	configs ConfigStruct
}

// GetMetadata 获取元数据
// param bucketName: 桶名称
// param pathName: 桶下路径
func (api APIWrapper) GetMetadata(bucketName, pathName string) (*grequests.Response, error) {
	req := RequestStruct{configs: api.configs}
	builder := apiBuilderStruct{configs: api.configs}
	url := builder.buildMetadataAPI(bucketName, pathName, nil)
	r, err := req.Get(url, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// UploadOneChunk 上传一个对象数据块
// param bucketName: 桶名称
// param dirPath: 桶下对象所在路径
// param objName: 对象名称
// param offset: 数据块在对象中的字节偏移量
// param chunk: 数据块
func (api APIWrapper) UploadOneChunk(bucketName, dirPath, objName string, offset int64, chunk []byte) (*grequests.Response, error) {
	if strings.Contains(objName, "/") {
		return nil, errors.New("Object name can not contains '/'")
	}

	req := RequestStruct{configs: api.configs}
	builder := apiBuilderStruct{configs: api.configs}
	url := builder.buildObjAPI(bucketName, dirPath, objName, nil)

	files := []grequests.FileUpload{
		{
			FileContents: ioutil.NopCloser(bytes.NewReader(chunk)),
			FieldName:    "chunk",
			FileMime:     "application/octet-stream",
		},
	}
	ro := &grequests.RequestOptions{
		Data: map[string]string{
			"chunk_offset": strconv.FormatInt(offset, 10),
			"chunk_size":   strconv.Itoa(len(chunk)),
		},
		Files: files,
	}
	r, err := req.Post(url, ro)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// DownloadOneChunk 下载一个对象数据块
// param bucketName: 桶名称
// param dirPath: 桶下对象所在路径
// param objName: 对象名称
// param offset: 数据块在对象中的字节偏移量
// param size: 要下载的分片大小
func (api APIWrapper) DownloadOneChunk(bucketName, dirPath, objName string, offset int64, size int) (*grequests.Response, error) {
	req := RequestStruct{configs: api.configs}
	builder := apiBuilderStruct{configs: api.configs}
	params := &map[string]string{
		"offset": strconv.FormatInt(offset, 10),
		"size":   strconv.Itoa(size),
	}
	url := builder.buildObjAPI(bucketName, dirPath, objName, params)

	r, err := req.Get(url, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// DeleteObject 删除一个对象
// param bucketName: 桶名称
// param dirPath: 桶下对象所在路径
// param objName: 对象名称
func (api APIWrapper) DeleteObject(bucketName, dirPath, objName string) (*grequests.Response, error) {
	req := RequestStruct{configs: api.configs}
	builder := apiBuilderStruct{configs: api.configs}
	url := builder.buildObjAPI(bucketName, dirPath, objName, nil)

	r, err := req.Delete(url, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// MakeDir 创建一个目录
// param bucketName: 桶名称
// param dirPath: 桶下目录所在路径
// param dirName: 目录名称
func (api APIWrapper) MakeDir(bucketName, dirPath, dirName string) (*grequests.Response, error) {
	req := RequestStruct{configs: api.configs}
	builder := apiBuilderStruct{configs: api.configs}
	url := builder.buildDirAPI(bucketName, dirPath, dirName, nil)

	r, err := req.Post(url, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// DeleteDir 删除一个空目录
// param bucketName: 桶名称
// param dirPath: 桶下目录所在路径
// param dirName: 目录名称
func (api APIWrapper) DeleteDir(bucketName, dirPath, dirName string) (*grequests.Response, error) {
	req := RequestStruct{configs: api.configs}
	builder := apiBuilderStruct{configs: api.configs}
	url := builder.buildDirAPI(bucketName, dirPath, dirName, nil)

	r, err := req.Delete(url, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// ListDirOnePage 自定义获取一页目录下的子目录和对象信息
// param bucketName: 桶名称
// param dirPath: 桶下目录所在路径
// param dirName: 目录名称
// param offset limit: 自定义从offset偏移量处获取limit条信息；offset和limit大于0时，参数有效；
//  	否则按服务器默认返回数据
func (api APIWrapper) ListDirOnePage(bucketName, dirPath, dirName string, offset, limit int) (*grequests.Response, error) {

	builder := apiBuilderStruct{configs: api.configs}

	var params = make(map[string]string)
	if offset > 0 {
		params["offset"] = strconv.Itoa(offset)
	}

	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	url := builder.buildDirAPI(bucketName, dirPath, dirName, &params)

	return api.ListDirOnePageByURL(url)
}

// ListDirOnePageByURL 通过url获取一页目录下的子目录和对象信息
func (api APIWrapper) ListDirOnePageByURL(url string) (*grequests.Response, error) {

	req := RequestStruct{configs: api.configs}
	r, err := req.Get(url, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// MoveRenameObject 移动或重命名一个对象
// param bucket_name: 桶名称
// param dirPath: 桶下对象所在路径
// param objName: 对象名称
// param moveTo: 移动对象到此目录路径，""为不移动, "/"为根目录
// param rename: 重命名对象，，""为不重命名
func (api APIWrapper) MoveRenameObject(bucketName, dirPath, objName, moveTo, rename string) (*grequests.Response, error) {

	req := RequestStruct{configs: api.configs}
	builder := apiBuilderStruct{configs: api.configs}

	params := make(map[string]string)
	if moveTo != "" {
		params["move_to"] = moveTo
	}

	if rename != "" {
		params["rename"] = rename
	}

	url := builder.buildMoveAPI(bucketName, dirPath, objName, &params)

	r, err := req.Post(url, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// ObjectPermission 对象公有或私有访问权限设置
// param bucket_name: 桶名称
// param dirPath: 桶下对象所在路径
// param objName: 对象名称
// param share: 是否分享公开，用于设置对象公有或私有, true(公有)，false(私有)
// param days: 对象公开分享天数(share=true时有效)，0表示永久公开，负数表示不公开，默认为0
func (api APIWrapper) ObjectPermission(bucketName, dirPath, objName string, share bool, days int) (*grequests.Response, error) {

	req := RequestStruct{configs: api.configs}
	builder := apiBuilderStruct{configs: api.configs}

	params := make(map[string]string)
	if share {
		params["share"] = "true"
	} else {
		params["share"] = "false"
	}

	params["days"] = strconv.Itoa(days)

	url := builder.buildObjAPI(bucketName, dirPath, objName, &params)

	r, err := req.Patch(url, nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}
