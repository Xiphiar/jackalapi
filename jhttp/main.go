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
	"github.com/JackalLabs/jackalgo/handlers/wallet_handler"
	"github.com/julienschmidt/httprouter"

	http2 "net/http"
)

const MaxFileSize = 32 << 30
const ParentFolder = "s/jhttp"

func main() {
	wallet, err := wallet_handler.NewWalletHandler(
		"slim odor fiscal swallow piece tide naive river inform shell dune crunch canyon ten time universe orchard roast horn ritual siren cactus upon forum",
		"https://jackal-testnet-rpc.polkachu.com:443",
		"lupulella-2")

	if err != nil {
		panic(err)
	}

	fileIo, err := file_io_handler.NewFileIoHandler(wallet.WithGas("500000"))
	if err != nil {
		panic(err)
	}

	res, err := fileIo.GenerateInitialDirs([]string{"jhttp"})
	if err != nil {
		panic(err)
	}

	fmt.Println(res.RawLog)

	fmt.Println(wallet.GetAddress())

	queue := http.NewQueue()
	go queue.Listen()

	Gets := make(http.Handlers, 0)
	Posts := make(http.Handlers, 0)

	Gets["/version"] = func(w http2.ResponseWriter, r *http2.Request, ps httprouter.Params) {
		_, err := w.Write([]byte("v0.0.0"))
		if err != nil {
			panic(err)
		}

		w.WriteHeader(200)
	}

	Posts["/upload"] = func(w http2.ResponseWriter, r *http2.Request, ps httprouter.Params) {
		// ParseMultipartForm parses a request body as multipart/form-data
		err = r.ParseMultipartForm(MaxFileSize) // MAX file size lives here
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

		m := queue.Push(fileUpload, folder, fileIo, &wg)

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

	http.Start(3535, Gets, Posts)
}
