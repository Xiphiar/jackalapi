package http

import (
	"net/http"

	"github.com/JackalLabs/jackalgo/handlers/file_io_handler"
	"github.com/julienschmidt/httprouter"
)

type Handlers map[string]*func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, queue *Queue, fileIo *file_io_handler.FileIoHandler)

type UploadResponse struct {
	FID string `json:"fid"`
}
