package main

import (
	"filestore/handler"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update", handler.FileMetaUpdateHandler)
	http.HandleFunc("/file/delete", handler.Deletehanddler)

	err := http.ListenAndServe(":8888", nil)

	if err != nil {
		log.Fatalf("Failed to start server%+v\n ", err)
	}
	log.Println("Server listening on 8888")

}
