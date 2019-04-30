/*
Package goharbor 是一个简单封装了EVHarbor RESTFULL API的golang包
*/
package goharbor

import (
	"errors"

	"github.com/levigross/grequests"
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

// GetMetadata 获取元数据
// param bucketName: 桶名称
// param pathName: 桶下路径
func (client ClientStruct) GetMetadata(bucketName, pathName string) (*grequests.Response, error) {
	return client.API.GetMetadata(bucketName, pathName)
}
