package japicore

import (
	"bytes"
	"fmt"
	"github.com/JackalLabs/jackalgo/handlers/file_io_handler"
	"github.com/JackalLabs/jackalgo/handlers/file_upload_handler"
	"net/http"
	"sync"
)

func processError(block string, caughtError error) {
	fmt.Printf("***** Error in block: %s *****\n", block)
	fmt.Println(caughtError)
	fmt.Println("***** End Error Report *****")
}

func processHttpPostError(block string, caughtError error, w http.ResponseWriter) {
	fmt.Printf("***** Error in block: %s *****\n", block)
	fmt.Println(caughtError)
	fmt.Println("***** End Error Report *****")
	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte(caughtError.Error()))
	if err != nil {
		processError(fmt.Sprintf("processHttpPostError for %s", block), err)
	}
}

func cloneBytes(reader *bytes.Reader) []byte {
	var allBytes []byte
	reader.Read(allBytes)
	reader.Seek(0, 0)
	return allBytes
}

func processUpload(w http.ResponseWriter, fileIo *file_io_handler.FileIoHandler, bytes []byte, cid string, path string, queue *FileIoQueue) string {
	fileUpload, err := file_upload_handler.TrackVirtualFile(bytes, cid, path)
	if err != nil {
		processHttpPostError("TrackVirtualFile", err, w)
		return ""
	}

	folder, err := fileIo.DownloadFolder(path)
	if err != nil {
		processHttpPostError("DownloadFolder", err, w)
		return ""
	}

	var wg sync.WaitGroup
	wg.Add(1)

	m := queue.Push(fileUpload, folder, fileIo, &wg)

	wg.Wait()

	if m.Error() != nil {
		processHttpPostError("UploadFailed", m.Error(), w)
		return ""
	}

	return m.Fid()
}
