//Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/aixiaoxiang/bast/guid"
)

//FileInfo struct  文件信息
type FileInfo struct {
	FileName string `json:"fileName"`
	RawName  string `json:"rawName"`
	Chunks   int    `json:"chunks"`
}

//FileDefault 默认文件路由
func FileDefault(dir string) {
	//文件读服务
	FileServer("/f/", dir)
	//文件上传服务-支持多线程、大文件
	Post("/files/upload", FileUploadHandle(dir))
	//文件合并服务
	Post("/files/merge", MergeHandle(dir))
}

//FileUploadHandle 文件上传的处理程序
func FileUploadHandle(dir string) func(ctx *Context) {
	return func(ctx *Context) {
		FileUpload(ctx, dir)
	}
}

//FileUpload 客户端上传多个文件,并带有请求参数
func FileUpload(ctx *Context, dir string) {
	//ctx.OutJSON(realFiles, SerOK, "上传成功")
	realFiles, err := FileHandleUpload(ctx, dir, false)
	if err != nil {
		ctx.JSONWithCode(err.Error(), SerError)
	} else {
		ctx.JSONWithCodeMsg(realFiles, SerOK, "上传成功")
	}
}

//FileHandleUpload 客户端上传多个文件,并带有请求参数
func FileHandleUpload(ctx *Context, dir string, returnRealFile bool) ([]FileInfo, error) {
	//ctx.ParseForm()
	ctx.ParseMultipartForm(32 << 40) //最大内存为64M
	mp := ctx.Request.MultipartForm
	if mp == nil {
		return nil, errors.New("上传格式不对")
	}
	if mp.File == nil || len(mp.File) == 0 {
		return nil, errors.New("没有上传文件")
	}
	//m5 := md5.New()
	var realFiles []FileInfo
	for _, v := range mp.File {
		for _, f := range v {
			fn := ctx.GetString("fn") //fn
			id := ctx.GetString("id") //id
			fn += id
			chunk, err := ctx.GetInt("chunk")   //chunk
			chunks, err := ctx.GetInt("chunks") //chunks
			fmt.Printf("chunk=%d,chunks=%d", chunk, chunks)
			fileName := f.Filename
			//m5.Write([]byte(fileName))
			//m5FileName := hex.EncodeToString(m5.Sum(nil))
			if fn == "" {
				fn = guid.GUID()
			}
			file, err := f.Open()
			if err != nil {
				return nil, errors.New("上传失败")
			}
			defer file.Close()
			//fn += m5FileName
			fn += path.Ext(fileName)
			rfn := fn
			if chunks > 0 {
				fn += "." + strconv.Itoa(chunk)
			}
			//fp := config.FileDir() + fn
			fp := dir + fn
			if returnRealFile {
				realFiles = append(realFiles, FileInfo{FileName: fp, RawName: fileName, Chunks: chunks})
			} else {
				realFiles = append(realFiles, FileInfo{FileName: "f/" + rfn, RawName: fileName, Chunks: chunks})
			}
			exist := PathExist(dir)
			if !exist {
				err := os.Mkdir(dir, os.ModePerm)
				if err != nil {
					return nil, errors.New("文件夹不存在")
				}
			}
			writer, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				//ctx.OutJSON("服务无法创建文件", SerError)
				return nil, errors.New("服务无法创建文件")
			}
			defer writer.Close()
			io.Copy(writer, file)
		}
	}
	//ctx.OutJSON(realFiles, SerOK, "上传成功")
	return realFiles, nil
}

//MergeHandle 合并文件的处理程序
func MergeHandle(dir string) func(ctx *Context) {
	return func(ctx *Context) {
		Merge(ctx, dir)
	}
}

//Merge 合并分片的文件
func Merge(ctx *Context, dir string) {
	var data []FileInfo
	err := ctx.JSONObj(&data)
	if err != nil {
		ctx.Failed("亲！数据有误")
		return
	}
	lg := len(data)
	files := make([]FileInfo, 0, lg)
	for i := 0; i < lg; i++ {
		o := &data[i]
		if err := mergeFile(o, dir); err == nil {
			files = append(files, *o)
		}
	}
	ctx.JSON(files)

}

//执行文件分片合并，并删除分片文件-内部使用
func mergeFile(fObj *FileInfo, dir string) error {
	fileName := fObj.FileName
	chunks := fObj.Chunks
	if chunks <= 0 {
		return nil
	}
	if fileName != "" {
		i := strings.Index(fileName, "/") + 1
		fileName = fileName[i:]
		//fObj.FileName = fileName
	}
	if fileName != "" && chunks > 0 {
		//f := config.FileDir() + fileName
		f := dir + fileName
		writer, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			//ctx.OutJSON("服务无法创建文件", SerError)
			return errors.New("服务无法创建文件")
		}
		defer writer.Close()
		for i := 0; i < chunks; i++ {
			fp := f + "." + strconv.Itoa(i)
			file, err := os.Open(fp)
			if err != nil {
				//ctx.OutJSON("合并分片失败", SerError)
				return errors.New("合并分片失败")
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err == nil {
				os.Remove(fp)
			}
		}
		return nil
	}
	//ctx.OutError("亲！文件名有误")
	return errors.New("亲！文件名有误")
}

// PathExist 判断文件夹是否存在
func PathExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
