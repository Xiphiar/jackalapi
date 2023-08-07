package http

import (
	"github.com/uptrace/bunrouter"
	"net/http"

	"github.com/JackalLabs/jackalgo/handlers/file_io_handler"
)

type Handlers map[string]*func(w http.ResponseWriter, r bunrouter.Request, queue *Queue, fileIo *file_io_handler.FileIoHandler) error

type UploadResponse struct {
	FID string `json:"fid"`
}
