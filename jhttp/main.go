package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/JackalLabs/jackalapi/jhttp/http"
	"github.com/JackalLabs/jackalgo/handlers/file_io_handler"
	"github.com/JackalLabs/jackalgo/handlers/file_upload_handler"
	"github.com/julienschmidt/httprouter"

	http2 "net/http"
)

const MaxFileSize = 32 << 30
const ParentFolder = "s/jhttp"

func main() {

	Gets := make(http.Handlers, 0)
	Posts := make(http.Handlers, 0)

	Gets["/version"] = func(w http2.ResponseWriter, r *http2.Request, ps httprouter.Params, q *http.Queue, fileIo *file_io_handler.FileIoHandler) {
		_, err := w.Write([]byte("v0.0.0"))
		if err != nil {
			panic(err)
		}
	}

	Posts["/upload"] = func(w http2.ResponseWriter, r *http2.Request, ps httprouter.Params, q *http.Queue, fileIo *file_io_handler.FileIoHandler) {
		// ParseMultipartForm parses a request body as multipart/form-data
		err := r.ParseMultipartForm(MaxFileSize) // MAX file size lives here
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http2.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		file, h, err := r.FormFile("file") // Retrieve the file from form data
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http2.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		var b bytes.Buffer

		_, err = io.Copy(&b, file)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http2.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		fileUpload, err := file_upload_handler.TrackVirtualFile(b.Bytes(), h.Filename, ParentFolder)
		if err != nil {
			fmt.Println("fileupload failed", err)
			w.WriteHeader(http2.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		folder, err := fileIo.DownloadFolder(ParentFolder)
		if err != nil {
			fmt.Println("download folder failed", err)
			w.WriteHeader(http2.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		var wg sync.WaitGroup
		wg.Add(1)

		m := q.Push(fileUpload, folder, fileIo, &wg)

		wg.Wait()

		if m.Error() != nil {
			fmt.Println("upload file failed", m.Error())
			w.WriteHeader(http2.StatusInternalServerError)
			_, _ = w.Write([]byte(m.Error().Error()))
			return
		}

		nv := http.UploadResponse{
			FID: m.Fid(),
		}
		err = json.NewEncoder(w).Encode(nv)
		if err != nil {
			panic(err)
		}
	}

	http.StartServer(Gets, Posts, []string{"jhttp"})

}
