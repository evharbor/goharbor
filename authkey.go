package goharbor

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type jsonBodyStruct struct {
	PathOfURL string `json:"path_of_url"`
	Method    string `json:"method"`
	Deadline  int64  `json:"deadline"`
}

// AuthKey 访问密钥结构体
type AuthKey struct {
	AccessKey, SecretKey string
}

// Key 生成访问密钥
// param uri: 未编码的原始全路径（path?query）字符串
// param method: 请求方法 GET POST PUT PATCH等
// param timedelta: 安全凭证的有效期时间增量（基于当前时间戳），单位为秒s
func (ak AuthKey) Key(uri string, method string, timedelta int64) string {
	deadline := time.Now().Unix() + timedelta //获取时间戳

	body := jsonBodyStruct{PathOfURL: uri, Method: method, Deadline: deadline}
	data, _ := json.Marshal(body)
	dataBase64 := base64.URLEncoding.EncodeToString(data)

	h := hmac.New(sha1.New, []byte(ak.SecretKey))
	h.Write([]byte(dataBase64))
	key := base64.URLEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s %s:%s:%s", "evhb-auth", ak.AccessKey, key, dataBase64)
}
