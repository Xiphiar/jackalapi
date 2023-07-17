package http

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Handlers map[string]func(w http.ResponseWriter, r *http.Request, ps httprouter.Params)

type UploadResponse struct {
	FID string `json:"fid"`
}
