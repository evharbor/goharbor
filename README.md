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
bucketName := "6666"
objPathName := "ddd/动次打次/uploadtest666.exe"
fileName := "C:\\Users\\admin\\Downloads\\VSCodeUserSetup-x64-1.33.1.exe"
r, err := client.UploadObject(bucketName, objPathName, fileName, 0)
if err != nil {
	fmt.Println(err)
}
if r.IsDone() {
	fmt.Println("上传成功")
}else{
	// 未上传完成，继续断点处上传
	r, err := client.UploadObject(bucketName, objPathName, fileName, r.Offset)
}

```

#### 下载一个对象
```go
bucketName := "6666"
objPathName := "ddd/动次打次/uploadtest.exe"
savePath := "C:\\Users\\admin\\Downloads"
newSaveFileName := ""
r, err := client.DownLoadObject(bucketName, objPathName, savePath, newSaveFileName, 0)
if err != nil {
	fmt.Println(err)
}
if r.IsDone() {
	fmt.Println("下载成功")
}else{
	// 从中断处继续下载
	r, err := client.DownLoadObject(bucketName, objPathName, savePath, newSaveFileName, r.Offset)
}
```

#### 对象或目录原数据
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

#### 创建一个目录
```go
bucketName := "6666"
parentDirPath := "ddd"
dirName := "makedir"
r, err := client.MakeDir(bucketName, parentDirPath, dirName)
if err != nil {
	fmt.Println(err)
}
if r.Ok {
	fmt.Println("创建目录成功:", r.CodeText)
}else{
	fmt.Println("创建目录失败:", r.CodeText)
}
```

#### 删除一个空目录
```go
r, err := client.DeleteDir("6666", "ddd/makedir")
if err != nil {
	fmt.Println(err)
}
if r.Ok {
	fmt.Println("删除目录成功:", r.CodeText)
} else {
	fmt.Println("删除目录失败:", r.CodeText)
}
```

#### 列举目录下的子目录和对象信息
```go
bucketName := "6666"
dirPathName := "/"
numPerPage := 100

dir := client.Dir(bucketName, dirPathName)
r, err := dir.ListFirstPage(numPerPage)
if err != nil {
	fmt.Println(err)
}
if r.Ok {
	fmt.Println(r.Files)
	// 下一页
	if r.HasNext() {
		r2, err2 := dir.NextPage()
		if err2 != nil {
			fmt.Println(err2)
		}
		if r2.Ok {
			fmt.Println(r2.Files)
			// 下一页
			if r2.HasNext() {
				r4, err4 := dir.NextPage()
			}
		}
	}
	// 上一页
	if r.HasPrevious() {
		r3, err3 := dir.PreviousPage()
		if err3 != nil {
			fmt.Println(err3)
		}
		if r3.Ok {
			fmt.Println(r3.Files)
		}
	}
} else {
	fmt.Println(r.CodeText)
}
```

#### 移动或重命名一个对象
* 移动并重命名对象
```go
bucketName := "6666"
objPathName := "ddd/动次打次/test.py"
newName := "rename_test.py"
moveTo := "/" // 桶下根目录
r, err := client.MoveRenameObject(bucketName, objPathName, moveTo, newName)
if err != nil {
	fmt.Println(err)
}
if r.Ok {
	fmt.Println("移动并重命名对象成功")
	fmt.Println(r.Obj) // 修改后的对象元数据
} else {
	fmt.Println(r.CodeText)
}
```
* 移动对象
```go
bucketName := "6666"
objPathName := "rename_test.py"
moveTo := "ddd"
r, err := client.MoveObject(bucketName, objPathName, moveTo)
if err != nil {
	fmt.Println(err)
}
if r.Ok {
	fmt.Println("移动对象成功")
	fmt.Println(r.Obj) // 修改后的对象元数据
} else {
	fmt.Println(r.CodeText)
}
```
* 重命名对象
```go 
bucketName := "6666"
objPathName := "ddd/rename_test.py"
newName := "test2.py"
r, err := client.RenameObject(bucketName, objPathName, newName)
if err != nil {
	fmt.Println(err)
}
if r.Ok {
	fmt.Println("重命名对象成功")
	fmt.Println(r.Obj) // 修改后的对象元数据
} else {
	fmt.Println(r.CodeText)
}
```

#### 对象公有或私有访问权限设置
```go
bucketName := "6666"
objPathName := "ddd/test2.py"
// 永久公开对象访问权限
r, err := client.ObjectPermission(bucketName, objPathName, true, 0)
// 公开对象访问权限7天
r, err = client.ObjectPermission(bucketName, objPathName, true, 7)
// 私有对象访问权限7天
r, err = client.ObjectPermission(bucketName, objPathName, false, 0)
if err != nil {
	fmt.Println(err)
}
if r.Ok {
	fmt.Println("设置对象访问权限成功")
} else {
	fmt.Println(r.CodeText)
}
```
