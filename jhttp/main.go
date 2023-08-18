package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	jhttp "github.com/JackalLabs/jackalapi/jhttp/http"
	"github.com/JackalLabs/jackalgo/handlers/file_upload_handler"
	"github.com/rs/cors"
	"github.com/uptrace/bunrouter"

	"net/http"
)

const MaxFileSize = 32 << 30
const ParentFolder = "s/jhttp"

func main() {

	router := bunrouter.New()

	port := os.Getenv("JHTTP_PORT")
	if len(port) == 0 {
		port = "3535"
	}

	portNum, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		panic(err)
	}

	q, fileIo := jhttp.InitServer([]string{"jhttp"})

	router.GET("/version", func(w http.ResponseWriter, r bunrouter.Request) error {
		_, err := w.Write([]byte("v0.0.0"))
		if err != nil {
			return err
		}

		return nil
	})

	router.GET("/download/:name", func(w http.ResponseWriter, r bunrouter.Request) error {

		name := r.Param("name")

		var allBytes []byte

		handler, err := fileIo.DownloadFile(fmt.Sprintf("%s/%s", ParentFolder, name))
		if err != nil {
			fmt.Println("download file failed", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return err
		}

		allBytes = handler.GetFile().Buffer().Bytes()

		_, err = w.Write(allBytes)
		if err != nil {
			return err
		}

		return nil
	})

	router.POST("/upload", func(w http.ResponseWriter, r bunrouter.Request) error {
		// ParseMultipartForm parses a request body as multipart/form-data
		err := r.ParseMultipartForm(MaxFileSize) // MAX file size lives here
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return err
		}

		file, h, err := r.FormFile("file") // Retrieve the file from form data
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return err
		}

		var b bytes.Buffer

		_, err = io.Copy(&b, file)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return err
		}

		fileUpload, err := file_upload_handler.TrackVirtualFile(b.Bytes(), h.Filename, ParentFolder)
		if err != nil {
			fmt.Println("fileupload failed", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return err
		}

		folder, err := fileIo.DownloadFolder(ParentFolder)
		if err != nil {
			fmt.Println("download folder failed", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return err
		}

		var wg sync.WaitGroup
		wg.Add(1)

		m := q.Push(fileUpload, folder, fileIo, &wg)

		wg.Wait()

		if m.Error() != nil {
			fmt.Println("upload file failed", m.Error())
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(m.Error().Error()))
			return err
		}

		nv := jhttp.UploadResponse{
			FID: m.Fid(),
		}
		err = json.NewEncoder(w).Encode(nv)
		if err != nil {
			panic(err)
		}

		return nil
	})

	handler := cors.Default().Handler(router)

	fmt.Printf("🌍 Started jHTTP: http://0.0.0.0:%d\n", portNum)
	err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", portNum), handler)
	if err != nil {
		panic(err)
	}

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("Server Closed\n")
		return
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}

}
