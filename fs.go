//Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/axfor/bast/guid"
	"github.com/axfor/bast/logs"
)

//Fs struct
type Fs struct {
	FileName string `json:"fileName"`
	RawName  string `json:"rawName"`
	Chunks   int    `json:"chunks"`
}

//FileDefault default file router
func FileDefault(dir string) {
	FileRouter(dir, "/f/", "/files/upload", "/files/merge")
}

//FileRouter file router
func FileRouter(dir, access, upload, merge string) {
	access2 := access
	if access != "" && access[0] == '/' {
		access2 = access[1:]
	}
	//access
	FileServer(access, dir)
	//upload
	Post(upload, FileUploadHandle(dir, access2))
	if merge != "" {
		//merge
		Post(merge, FileMergeHandle(dir))
	}
}

//FileUploadHandle return a upload handle
func FileUploadHandle(dir, access string) func(ctx *Context) {
	return func(ctx *Context) {
		FileUpload(ctx, dir, access)
	}
}

//FileUpload real upload handle
func FileUpload(ctx *Context, dir, access string) {
	realFiles, err := doFileUpload(ctx, dir, access, false)
	if err != nil {
		ctx.JSONWithCode(err.Error(), SerError)
	} else {
		ctx.JSONWithCodeMsg(realFiles, SerOK, "upload sucess")
	}
}

//doFileUpload real upload handle and returns file info
func doFileUpload(ctx *Context, dir, access string, returnRealFile bool) ([]Fs, error) {
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
	var realFiles []Fs
	for _, v := range mp.File {
		for _, f := range v {
			err := fileUpload(ctx, dir, access, returnRealFile, f, &realFiles)
			if err != nil {
				return nil, errors.New("upload error")
			}
		}
	}
	return realFiles, nil
}

func fileUpload(ctx *Context, dir, access string, returnRealFile bool, f *multipart.FileHeader, realFiles *[]Fs) error {
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
	//m5 := md5.New()
	//m5.Write([]byte(fileName))
	//m5FileName := hex.EncodeToString(m5.Sum(nil))
	if fn == "" {
		fn = guid.GUID()
	}
	file, err := f.Open()
	if err != nil {
		return errors.New("upload error")
	}
	defer file.Close()
	fn += path.Ext(fileName)
	rfn := fn
	if chunks > 0 {
		fn += "." + strconv.Itoa(chunk)
	}
	fp := dir + fn
	err = ExistAndAutoMkdir(dir)
	if err != nil {
		return err
	}
	writer, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return errors.New("failed to create file")
	}
	defer writer.Close()
	_, err = io.Copy(writer, file)
	if err != nil {
		return errors.New("copy file error")
	}
	var fs Fs
	if returnRealFile {
		fs = Fs{FileName: fp, RawName: fileName, Chunks: chunks}
	} else {
		fs = Fs{FileName: access + rfn, RawName: fileName, Chunks: chunks}
	}
	if (chunk+1) == chunks && chunks > 0 {
		err = doFileMerge(&fs, dir)
		if err != nil {
			return errors.New("merge file error")
		}
	}
	*realFiles = append(*realFiles, fs)
	return nil
}

//FileMergeHandle return merge handle
func FileMergeHandle(dir string) func(ctx *Context) {
	return func(ctx *Context) {
		FileMerge(ctx, dir)
	}
}

//FileMerge files
func FileMerge(ctx *Context, dir string) {
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
		if err := doFileMerge(o, dir); err == nil {
			files = append(files, *o)
		}
	}
	ctx.JSON(files)
}

//doFileMerge real merge files and delete invalid files
func doFileMerge(fObj *Fs, dir string) error {
	fileName := fObj.FileName
	chunks := fObj.Chunks
	if chunks <= 0 {
		return nil
	}
	if fileName != "" {
		i := strings.Index(fileName, "/") + 1
		fileName = fileName[i:]
	}
	if fileName != "" {
		f := dir + fileName
		if Exist(f) {
			return nil
		}
		writer, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return errors.New("failed to create file")
		}
		defer writer.Close()
		for i := 0; i < chunks; i++ {
			fp := f + "." + strconv.Itoa(i)
			err = fileMerge(fp, writer)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return errors.New("invalid files info")
}

func fileMerge(fp string, writer *os.File) error {
	file, err := os.Open(fp)
	if err != nil {
		return errors.New("failed to merge files")
	}
	defer file.Close()
	_, err = io.Copy(writer, file)
	if err == nil {
		os.Remove(fp)
	}
	return err
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
