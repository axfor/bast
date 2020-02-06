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

//Fs struct
type Fs struct {
	FileName string `json:"fileName"`
	RawName  string `json:"rawName"`
	Chunks   int    `json:"chunks"`
}

//Default default file router
func Default(dir string) {
	Router(dir, "/f/", "/files/upload", "/files/merge")
}

//Router file router
func Router(dir, access, upload, merge string) {
	access2 := access
	if access != "" && access[0] == '/' {
		access2 = access[1:]
	}
	//get
	FileServer(access, dir)
	//upload
	Post(upload, UploadHandle(dir, access2))
	//merge
	Post(merge, MergeHandle(dir))
}

//UploadHandle return a upload handle
func UploadHandle(dir, access string) func(ctx *Context) {
	return func(ctx *Context) {
		Upload(ctx, dir, access)
	}
}

//Upload real upload handle
func Upload(ctx *Context, dir, access string) {
	realFiles, err := DoUpload(ctx, dir, access, false)
	if err != nil {
		ctx.JSONWithCode(err.Error(), SerError)
	} else {
		ctx.JSONWithCodeMsg(realFiles, SerOK, "upload sucess")
	}
}

//DoUpload real upload handle and returns file info
func DoUpload(ctx *Context, dir, access string, returnRealFile bool) ([]Fs, error) {
	err := ctx.ParseMultipartForm(32 << 40) //maximum 64M
	if err != nil {
		logs.Errors("parseMultipartForm error", err)
		return nil, errors.New("invalid file format")
	}
	mp := ctx.In.MultipartForm
	if mp == nil {
		return nil, errors.New("invalid file format")
	}
	if mp.File == nil || len(mp.File) == 0 {
		return nil, errors.New("not to upload files")
	}
	//m5 := md5.New()
	var realFiles []Fs
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
			fn += path.Ext(fileName)
			rfn := fn
			if chunks > 0 {
				fn += "." + strconv.Itoa(chunk)
			}
			fp := dir + fn
			if returnRealFile {
				realFiles = append(realFiles, Fs{FileName: fp, RawName: fileName, Chunks: chunks})
			} else {
				realFiles = append(realFiles, Fs{FileName: access + rfn, RawName: fileName, Chunks: chunks})
			}
			err = ExistAndAutoMkdir(dir)
			if err != nil {
				return nil, err
			}
			writer, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				return nil, errors.New("failed to create file")
			}
			defer writer.Close()
			io.Copy(writer, file)
			//
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
	var data []Fs
	err := ctx.JSONObj(&data)
	if err != nil {
		ctx.Failed("invalid merge files info")
		return
	}
	lg := len(data)
	files := make([]Fs, 0, lg)
	for i := 0; i < lg; i++ {
		o := &data[i]
		if err := DoMerge(o, dir); err == nil {
			files = append(files, *o)
		}
	}
	ctx.JSON(files)
}

//DoMerge real merge files and delete invalid files
func DoMerge(fObj *Fs, dir string) error {
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
			return errors.New("failed to create file")
		}
		defer writer.Close()
		for i := 0; i < chunks; i++ {
			fp := f + "." + strconv.Itoa(i)
			file, err := os.Open(fp)
			if err != nil {
				return errors.New("failed to Merge files")
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			if err == nil {
				os.Remove(fp)
			}
		}
		return nil
	}
	return errors.New("invalid files info")
}

// Exist returns a boolean indicating whether the error is known to
// report that a file or directory does not exist. It is satisfied by
// ErrNotExist as well as some syscall errors.
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

//ExistAndAutoMkdir Check that the file directory exists, there is no automatically created
func ExistAndAutoMkdir(filename string) (err error) {
	filename = path.Dir(filename)
	_, err = os.Stat(filename)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(filename, os.ModePerm)
		if err == nil {
			return nil
		}
	}
	return err
}
