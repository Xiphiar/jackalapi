package japicore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JackalLabs/jackalgo/handlers/file_io_handler"
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
			processError("WWriteError for MethodNotAllowedHandler", err)
		}
		return nil
	}
}

func VersionHandler() bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		version := "v0.0.0"
		_, err := w.Write([]byte(version))
		if err != nil {
			processError("WWriteError for VersionHandler", err)
		}
		return nil
	}
}

func ImportHandler(fileIo *file_io_handler.FileIoHandler, queue *ScrapeQueue) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		var data fileScrape
		source := req.Header.Get("J-Source-Path")

		err := json.NewDecoder(req.Body).Decode(&data)
		if err != nil {
			processHttpPostError("JSONDecoder", err, w)
			return nil
		}

		var wg sync.WaitGroup

		for _, target := range data.targets {
			wg.Add(1)
			queue.Push(fileIo, w, &wg, "bulk", target, source)
		}

		wg.Wait()

		_, err = w.Write([]byte("Import complete"))
		if err != nil {
			processError("WWriteError for ImportHandler", err)
		}
		return nil
	}
}

func IpfsHandler(fileIo *file_io_handler.FileIoHandler, queue *FileIoQueue) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		var allBytes []byte

		operatingRoot := os.Getenv("JAPI_IPFS_ROOT")
		if len(operatingRoot) == 0 {
			operatingRoot = "s/JAPI/IPFS"
		}
		gateway := os.Getenv("JAPI_IPFS_GATEWAY")
		if len(gateway) == 0 {
			gateway = "https://ipfs.io/ipfs/"
		}
		toClone := false
		cloneHeader := req.Header.Get("J-Clone-Ipfs")
		if strings.ToLower(cloneHeader) == "true" {
			toClone = true
		}

		id := req.Param("id")
		if len(id) == 0 {
			warning := "Failed to get IPFS CID"
			asError := errors.New(strings.ToLower(warning))
			processHttpPostError("processUpload", asError, w)
			return asError
		}

		cid := strings.ReplaceAll(id, "/", "_")

		handler, err := fileIo.DownloadFile(fmt.Sprintf("%s/%s", operatingRoot, cid))
		if err != nil {
			if !toClone {
				warning := "IPFS CID Not Found"
				w.WriteHeader(404)
				_, err := w.Write([]byte(warning))
				if err != nil {
					return err
				}
				return errors.New(strings.ToLower(warning))
			}

			byteBuffer, err := httpGetFileRequest(w, gateway, cid)
			if err != nil {
				processHttpPostError("httpGetFileRequest", err, w)
				return nil
			}

			byteReader := bytes.NewReader(byteBuffer.Bytes())
			workingBytes := cloneBytes(byteReader)
			allBytes = cloneBytes(byteReader)

			fid := processUpload(w, fileIo, workingBytes, cid, operatingRoot, queue)
			if len(fid) == 0 {
				warning := "Failed to get FID post-upload"
				asError := errors.New(strings.ToLower(warning))
				processHttpPostError("IpfsHandler", asError, w)
				return asError
			}
		} else {
			allBytes = handler.GetFile().Buffer().Bytes()
		}
		_, err = w.Write(allBytes)
		if err != nil {
			processError("WWriteError for IpfsHandler", err)
			return err
		}
		return nil
	}
}

func DownloadHandler(fileIo *file_io_handler.FileIoHandler) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		id := req.Param("id")
		if len(id) == 0 {
			warning := "Failed to get FileName"
			asError := errors.New(strings.ToLower(warning))
			processHttpPostError("processUpload", asError, w)
			return asError
		}
		fid := strings.ReplaceAll(id, "/", "_")

		handler, err := fileIo.DownloadFileFromFid(fid)
		if err != nil {
			return err
		}

		fileBytes := handler.GetFile().Buffer().Bytes()
		_, err = w.Write(fileBytes)
		if err != nil {
			processError("WWriteError for DownloadHandler", err)
		}
		return nil
	}
}

func UploadHandler(fileIo *file_io_handler.FileIoHandler, queue *FileIoQueue) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		var byteBuffer bytes.Buffer
		var wg sync.WaitGroup
		wg.Add(1)
		WorkingFileSize := 32 << 30

		envSize := os.Getenv("JAPI_MAX_FILE")
		if len(envSize) > 0 {
			envParse, err := strconv.Atoi(envSize)
			if err != nil {
				return err
			}
			WorkingFileSize = envParse
		}
		MaxFileSize := int64(WorkingFileSize)

		operatingRoot := os.Getenv("JAPI_OP_ROOT")
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

		fid := processUpload(w, fileIo, byteBuffer.Bytes(), head.Filename, operatingRoot, queue)
		if len(fid) == 0 {
			warning := "Failed to get FID"
			asError := errors.New(strings.ToLower(warning))
			processHttpPostError("processUpload", asError, w)
			return asError
		}

		successfulUpload := UploadResponse{
			FID: fid,
		}
		err = json.NewEncoder(w).Encode(successfulUpload)
		if err != nil {
			processHttpPostError("JSONSuccessEncode", err, w)
			return nil
		}

		_, err = w.Write([]byte("uploadHandler"))
		if err != nil {
			processError("WWriteError for UploadHandler", err)
		}
		return nil
	}
}

func DeleteHandler(fileIo *file_io_handler.FileIoHandler) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		id := req.Param("id")
		if len(id) == 0 {
			warning := "Failed to get FileName"
			asError := errors.New(strings.ToLower(warning))
			processHttpPostError("processUpload", asError, w)
			return asError
		}

		fid := strings.ReplaceAll(id, "/", "_")
		fmt.Println(fid)

		// TODO - add file deletion to fileIo
		//fileIo.deleteFile

		return nil
	}
}
