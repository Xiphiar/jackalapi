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
		toClone := false

		cloneHeader := req.Header.Get("J-Clone-Ipfs")
		if strings.ToLower(cloneHeader) == "true" {
			toClone = true
		}

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

func UploadHandler(fileIo *file_io_handler.FileIoHandler, queue *Queue) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		var byteBuffer bytes.Buffer
		var wg sync.WaitGroup
		wg.Add(1)
		WorkingFileSize := 32 << 30

		envSize := os.Getenv("JHTTP_MAX_FILE")
		if len(envSize) > 0 {
			envParse, err := strconv.Atoi(envSize)
			if err != nil {
				return err
			}
			WorkingFileSize = envParse
		}
		MaxFileSize := int64(WorkingFileSize)

		operatingRoot := os.Getenv("JHTTP_OP_ROOT")
		if len(operatingRoot) == 0 {
			operatingRoot = "s/JAPI"
		}

		// ParseMultipartForm parses a request body as multipart/form-data
		err := req.ParseMultipartForm(MaxFileSize) // MAX file size lives here
		if err != nil {
			processHttpPostError("ParseMultipartForm", err, w)
			return nil
		}

		// Retrieve the file from form data
		file, head, err := req.FormFile("file")
		if err != nil {
			processHttpPostError("FormFileFile", err, w)
			return nil
		}

		_, err = io.Copy(&byteBuffer, file)
		if err != nil {
			processHttpPostError("MakeByteBuffer", err, w)
			return nil
		}

		fileUpload, err := file_upload_handler.TrackVirtualFile(byteBuffer.Bytes(), head.Filename, operatingRoot)
		if err != nil {
			processHttpPostError("TrackVirtualFile", err, w)
			return nil
		}

		folder, err := fileIo.DownloadFolder(operatingRoot)
		if err != nil {
			processHttpPostError("DownloadFolder", err, w)
			return nil
		}

		m := queue.Push(fileUpload, folder, fileIo, &wg)

		wg.Wait()

		if m.Error() != nil {
			processHttpPostError("UploadFailed", m.Error(), w)
			return nil
		}

		successfulUpload := UploadResponse{
			FID: m.Fid(),
		}
		err = json.NewEncoder(w).Encode(successfulUpload)
		if err != nil {
			processHttpPostError("JSONSuccessEncode", err, w)
			return nil
		}

		_, err = w.Write([]byte("uploadHandler"))
		if err != nil {
			processError("UploadHandlerWrite", err)
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
