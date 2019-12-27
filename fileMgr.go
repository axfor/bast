//Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/aixiaoxiang/bast/guid"
	"github.com/aixiaoxiang/bast/logs"
)

//FileInfo struct
type FileInfo struct {
	FileName string `json:"fileName"`
	RawName  string `json:"rawName"`
	Chunks   int    `json:"chunks"`
}

//FileDefault default file router
func FileDefault(dir string) {
	//get
	FileServer("/f/", dir)
	//upload
	Post("/files/upload", FileUploadHandle(dir))
	//merge
	Post("/files/merge", MergeHandle(dir))
}

//FileUploadHandle return a upload handle
func FileUploadHandle(dir string) func(ctx *Context) {
	return func(ctx *Context) {
		FileUpload(ctx, dir)
	}
}

//FileUpload real upload handle
func FileUpload(ctx *Context, dir string) {
	realFiles, err := FileHandleUpload(ctx, dir, false)
	if err != nil {
		ctx.JSONWithCode(err.Error(), SerError)
	} else {
		ctx.JSONWithCodeMsg(realFiles, SerOK, "upload sucess")
	}
}

//FileHandleUpload real upload handle and returns file info
func FileHandleUpload(ctx *Context, dir string, returnRealFile bool) ([]FileInfo, error) {
	//ctx.ParseForm()
	err := ctx.ParseMultipartForm(32 << 40) //maximum 64M
	if err != nil {
		logs.Errors("parseMultipartForm error", err)
		return nil, errors.New("Invalid file format")
	}
	mp := ctx.In.MultipartForm
	if mp == nil {
		return nil, errors.New("Invalid file format")
	}
	if mp.File == nil || len(mp.File) == 0 {
		return nil, errors.New("Not to upload files")
	}
	//m5 := md5.New()
	var realFiles []FileInfo
	for _, v := range mp.File {
		for _, f := range v {
			fn := ctx.GetTrimString("fn") //fn
			if fn == "" {
				fn = ctx.GetTrimString("filename")
			}
			if fn == "" {
				fn = ctx.GetTrimString("fileName")
			}
			id := ctx.GetTrimString("id") //id
			fn += id
			chunk, err := ctx.GetInt("chunk")   //chunk
			chunks, err := ctx.GetInt("chunks") //chunks
			// fmt.Printf("chunk=%d,chunks=%d", chunk, chunks)
			fileName := f.Filename
			//m5.Write([]byte(fileName))
			//m5FileName := hex.EncodeToString(m5.Sum(nil))
			if fn == "" {
				fn = guid.GUID()
			}
			file, err := f.Open()
			if err != nil {
				return nil, errors.New("upload error")
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
					return nil, errors.New("Directory does not exist")
				}
			}
			writer, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				return nil, errors.New("Failed to create file")
			}
			defer writer.Close()
			io.Copy(writer, file)
		}
	}
	return realFiles, nil
}

//MergeHandle return merge handle
func MergeHandle(dir string) func(ctx *Context) {
	return func(ctx *Context) {
		Merge(ctx, dir)
	}
}

//Merge files
func Merge(ctx *Context, dir string) {
	var data []FileInfo
	err := ctx.JSONObj(&data)
	if err != nil {
		ctx.Failed("Invalid merge files info")
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

//real merge files and delete invalid files
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
			return errors.New("Failed to create file")
		}
		defer writer.Close()
		for i := 0; i < chunks; i++ {
			fp := f + "." + strconv.Itoa(i)
			file, err := os.Open(fp)
			if err != nil {
				return errors.New("Failed to Merge files")
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err == nil {
				os.Remove(fp)
			}
		}
		return nil
	}
	return errors.New("Invalid files info")
}

// PathExist returns a boolean indicating whether the error is known to
// report that a file or directory does not exist. It is satisfied by
// ErrNotExist as well as some syscall errors.
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
