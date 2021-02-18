package handler

import (
	"encoding/json"
	"filestore/meta"
	"filestore/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// UploadHandler deal with upload files
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 返回上传页面
		content, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		}
		w.Write(content)
	} else if r.Method == "POST" {
		//接受文件流并保存到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Println("failed to get data\n", err.Error())
			return
		}

		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "./tmp/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:00"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Println("failed to create a new file:", err.Error())
			return
		}

		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Println("failed to save data into file:", err.Error())
			return
		}
		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		meta.UploadFileMeta(fileMeta)

		w.Write([]byte("Upload success"))

	}
}

func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var filehash []string
	var ok bool
	if filehash, ok = r.Form["filehash"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
	}

	fMeta := meta.GetFileMeta(filehash[0])
	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	fmt.Println(filehash)
	//TODO: get file according to the hash value,and return to the client
	fmeta := meta.GetFileMeta(filehash) // should detect err
	file, err := os.Open(fmeta.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	// read file from location
	data, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-disposition", "attachment; filename=\""+fmeta.FileName+"\"")
	w.Write(data)
}

func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	opType := r.Form.Get("op")
	filehash := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	curMeta := meta.GetFileMeta(filehash)
	curMeta.FileName = newFileName
	meta.UploadFileMeta(curMeta)
	data, err := json.Marshal(curMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
	w.WriteHeader(http.StatusOK)

}

func Deletehanddler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	fmeta := meta.GetFileMeta(filehash)
	os.Remove(fmeta.Location)
	meta.DeleteFileMeta(filehash)

	w.WriteHeader(http.StatusNoContent)

}
