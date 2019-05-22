### Requires
go get -u golang.org/x/net/publicsuffix  
被墙的话可以从github下载net(https://github.com/golang/net.git)放到 {your GOPATH}/src/golang.org/x/下。
go get -u github.com/google/go-querystring   

### Install
go get -u github.com/evharbor/goharbor

### Usage
import "github.com/evharbor/goharbor"

### Quick Start
#### 创建客户端

```go
package main

import (
    harbor "github.com/evharbor/goharbor"
)

func main() {
	configs, err := harbor.InitConfig(map[harbor.ConfigKeyType]string{
		// harbor.VERSION:   "v2",      // 默认v1
		harbor.SCHEME: harbor.HTTP,     // 默认HTTPS
		harbor.HOST:   "10.0.86.213",   // 默认obs.casearth.cn
		harbor.ACCESSKEY: "fddcdd54341511e9bd0ec800a000655d",
		harbor.SECRETKEY: "8a01eb85aab4f653ffdd13ee0834f0861042e253",
	})
	if err != nil {
		fmt.Println("InitConfig error:", err)
	}

	client := harbor.InitClient(configs)
}
```

#### 上传一个对象数据块
```go
str := "test"
data := []byte(str)
r, err := client.UploadOneChunk("wwww", "cc/UploadOneChunk2", 0, data)
if err != nil {
    fmt.Println(err)
} else {
    fmt.Println(r.ResultsDetail())
}
```

#### 下载一个对象数据块
```go
var offset int64 = 1024*1024
r2, err := client.DownloadOneChunk("6666", "test.py", offset, 1024)
if err != nil {
	fmt.Println(err)
} else {
	if !r2.Ok {
		fmt.Println(r2.ResultsDetail())
		return
	}

	var data []byte = r2.Chunk
	saveFile, saveError := os.Create("C:\\Users\\admin\\Downloads\\testd.txt")
	if saveError != nil {
		return // exit the function on error
	}
	defer saveFile.Close()
	s, seekErr := saveFile.Seek(offset, os.SEEK_SET)
	if seekErr != nil || s != offset {
		fmt.Println(seekErr)
	}
	saveFile.Write(data)
}

```

#### 上传一个对象
```go
r, err := client.UploadObject("6666", "ddd/动次打次/uploadtest.exe", "C:\\Users\\admin\\Downloads\\VSCodeUserSetup-x64-1.33.1.exe", 0)
if err != nil {
	fmt.Println(err)
}
if r.IsDone() {
	fmt.Println("上传成功")
}
```

#### 下载一个对象
```go
	r, err = client.DownLoadObject("6666", "ddd/动次打次/uploadtest.exe", "C:\\Users\\admin\\Downloads", "", 0)
	if err != nil {
		fmt.Println(err)
	}
	if r.IsDone() {
		fmt.Println("下载成功")
	}
```

#### 对象原数据
```go
r, err := client.GetMetadata("wwww", "cc/UploadOneChunk2")
if err != nil {
	fmt.Println(err)
} else {
    fmt.Println(r.Data.Name)
    fmt.Println(r.Data.FileOrDir)
    fmt.Println(r.Data.Size)
}
```


