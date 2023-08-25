package japicore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JackalLabs/jackalgo/handlers/file_io_handler"
	"github.com/JackalLabs/jackalgo/handlers/file_upload_handler"
	"github.com/uptrace/bunrouter"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

func handler() bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		return nil
	}
}

func MethodNotAllowedHandler() bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		w.WriteHeader(http.StatusMethodNotAllowed)
		warning := fmt.Sprintf("%s method not availble for \"%s\"", req.URL.Path, req.Method)

		_, err := w.Write([]byte(warning))
		if err != nil {
			return err
		}
		return nil
	}
}

func VersionHandler() bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		version := "v0.0.0"
		_, err := w.Write([]byte(version))
		if err != nil {
			return err
		}
		return nil
	}
}

func ImportHandler(fileIo *file_io_handler.FileIoHandler) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		var list fileScrape
		source := req.Header.Get("J-Source-Path")
		//TODO
		err := json.NewDecoder(req.Body).Decode(&list)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil
		}
	}
}

func IpfsHandler(fileIo *file_io_handler.FileIoHandler) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		id := req.Param("id")
		if len(id) == 0 {
			w.WriteHeader(500)
			return errors.New("failed to get fileName")
		}
		fid := strings.ReplaceAll(id, "/", "_")

		handler, err := fileIo.DownloadFileFromFid(fid)
		if err != nil {
			return err
		}

		fileBytes := handler.GetFile().Buffer().Bytes()
		_, err = w.Write(fileBytes)
		if err != nil {
			return err
		}
		return nil
	}
}

func DownloadHandler(fileIo *file_io_handler.FileIoHandler) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		id := req.Param("id")
		if len(id) == 0 {
			w.WriteHeader(500)
			return errors.New("failed to get fileName")
		}
		fid := strings.ReplaceAll(id, "/", "_")

		handler, err := fileIo.DownloadFileFromFid(fid)
		if err != nil {
			return err
		}

		fileBytes := handler.GetFile().Buffer().Bytes()
		_, err = w.Write(fileBytes)
		if err != nil {
			return err
		}
		return nil
	}
}

func UploadHandler(fileIo *file_io_handler.FileIoHandler) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {

		MaxFileSize := 32 << 30
		envSize := os.Getenv("JHTTP_MAX_FILE")
		if len(envSize) > 0 {
			envParse, err := strconv.Atoi(envSize)
			if err != nil {
				return err
			}
			MaxFileSize = envParse
		}

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

		_, err := w.Write([]byte("uploadHandler"))
		if err != nil {
			return err
		}
		return nil
	}
}

func DeleteHandler(fileIo *file_io_handler.FileIoHandler) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		id := req.Param("id")
		if len(id) == 0 {
			w.WriteHeader(500)
			return errors.New("failed to get fileName")
		}
		fid := strings.ReplaceAll(id, "/", "_")

		// TODO - add file deletion to fileIo
		//fileIo.deleteFile

		_, err := w.Write([]byte("deleteHandler"))
		if err != nil {
			return err
		}
		return nil
	}
}
