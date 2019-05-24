package goharbor

import (
	"net/url"
	"strings"
)

type apiBuilderStruct struct {
	configs ConfigStruct
}

func buildPath(slice []string) string {
	a := []string{}
	for _, value := range slice {
		if v := strings.Trim(value, "/"); v != "" {
			a = append(a, value)
		}
	}
	return strings.Join(a, "/")
}

func (builder apiBuilderStruct) getConfigs() *ConfigStruct {
	return &(builder.configs)
}

// buildURL return url.URL结构体对象
// If you do not intend to use the `params` you can just pass nil
func (builder apiBuilderStruct) buildURL(urlPath string, params *map[string]string) url.URL {
	configs := builder.getConfigs()

	query := ""
	if params != nil {
		v := url.Values{}
		for key, value := range *params {
			v.Add(key, value)
		}
		query = v.Encode() //编码后的query string
	}
	u := url.URL{
		Scheme:   configs.Scheme,
		Host:     configs.Host,
		Path:     urlPath, //未编码的path
		RawQuery: query,
	}

	return u
}

// buildObjAPI 构建对象api url（已编码）
// If you do not intend to use the `params` you can just pass nil
func (builder apiBuilderStruct) buildObjAPI(bucketName, dirPath, objName string, params *map[string]string) string {
	configs := builder.getConfigs()

	slice := []string{"api", configs.Version, configs.APIObjPrefix, bucketName, dirPath, objName}
	path := buildPath(slice) + "/"

	url := builder.buildURL(path, params)
	return url.String()
}

// buildMetadataAPI 构建对象api url（已编码）
// If you do not intend to use the `params` you can just pass nil
func (builder apiBuilderStruct) buildMetadataAPI(bucketName, pathName string, params *map[string]string) string {
	configs := builder.getConfigs()

	slice := []string{"api", configs.Version, configs.APIMetadataPrefix, bucketName, pathName}
	path := buildPath(slice) + "/"

	url := builder.buildURL(path, params)
	return url.String()
}

// buildDirAPI 构建目录api url（已编码）
// If you do not intend to use the `params` you can just pass nil
func (builder apiBuilderStruct) buildDirAPI(bucketName, dirPath, dirName string, params *map[string]string) string {
	configs := builder.getConfigs()

	slice := []string{"api", configs.Version, configs.APIDirPrefix, bucketName, dirPath, dirName}
	path := buildPath(slice) + "/"

	url := builder.buildURL(path, params)
	return url.String()
}

// buildBucketAPI 构建存储桶api url（已编码）
// If you do not intend to use the `params` you can just pass nil
func (builder apiBuilderStruct) buildBucketAPI(bucketID string, params *map[string]string) string {
	configs := builder.getConfigs()

	slice := []string{"api", configs.Version, configs.APIBucketPrefix, bucketID}
	path := buildPath(slice) + "/"

	url := builder.buildURL(path, params)
	return url.String()
}

// buildMoveAPI 构建移动重命名api url（已编码）
// If you do not intend to use the `params` you can just pass nil
func (builder apiBuilderStruct) buildMoveAPI(bucketName, dirPath, objName string, params *map[string]string) string {
	configs := builder.getConfigs()

	slice := []string{"api", configs.Version, configs.APIMovePrefix, bucketName, dirPath, objName}
	path := buildPath(slice) + "/"

	url := builder.buildURL(path, params)
	return url.String()
}

// buildStatsAPI 构建存储桶资源统计api url（已编码）
// If you do not intend to use the `params` you can just pass nil
func (builder apiBuilderStruct) buildStatsAPI(bucketName string, params *map[string]string) string {
	configs := builder.getConfigs()

	slice := []string{"api", configs.Version, configs.APIStatsPrefix, bucketName}
	path := buildPath(slice) + "/"

	url := builder.buildURL(path, params)
	return url.String()
}
