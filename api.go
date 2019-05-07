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
	r, err := req.Put(url, ro)
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
