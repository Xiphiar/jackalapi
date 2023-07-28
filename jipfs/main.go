package main

import (
	"bytes"
	"fmt"
	"io"
	http2 "net/http"
	url2 "net/url"
	"strings"
	"sync"

	"github.com/JackalLabs/jackalapi/jhttp/http"
	"github.com/JackalLabs/jackalapi/jipfs/ipfs"
	"github.com/JackalLabs/jackalgo/handlers/file_io_handler"
	"github.com/JackalLabs/jackalgo/handlers/file_upload_handler"
	"github.com/julienschmidt/httprouter"
)

const ParentFolder = "s/jipfs"

func main() {

	h, err := ipfs.MakeHost(0, 0)
	if err != nil {
		panic(err)
	}
	defer h.Close()

	Gets := make(http.Handlers, 0)
	Posts := make(http.Handlers, 0)

	getIpfs := func(w http2.ResponseWriter, r *http2.Request, ps httprouter.Params, q *http.Queue, fileIo *file_io_handler.FileIoHandler) {

		cid := ps.ByName("cid")
		if len(cid) == 0 {
			w.WriteHeader(500)
			return
		}
		cid = cid[1:]
		fileName := strings.ReplaceAll(cid, "/", "_")

		var allBytes []byte

		handler, err := fileIo.DownloadFile(fmt.Sprintf("%s/%s", ParentFolder, fileName))
		if err != nil {
			url, err := url2.Parse("https://ipfs.io/ipfs/")
			if err != nil {
				fmt.Println("urlparse failed", err)
				w.WriteHeader(http2.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			url = url.JoinPath(cid)

			req, err := http2.NewRequest("GET", url.String(), nil)
			if err != nil {
				fmt.Println("fileupload failed", err)
				w.WriteHeader(http2.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			res, err := http2.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("fileupload failed", err)
				w.WriteHeader(http2.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			var b bytes.Buffer
			size, err := io.Copy(&b, res.Body)
			if err != nil {
				fmt.Println("fileupload failed", err)
				w.WriteHeader(http2.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			fmt.Printf("saving file with size of %d.\n", size)
			err = res.Body.Close()
			if err != nil {
				fmt.Println("fileupload failed", err)
				w.WriteHeader(http2.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			allBytes = b.Bytes()
			bs := make([]byte, len(allBytes))
			copy(bs, allBytes)

			fileUpload, err := file_upload_handler.TrackVirtualFile(bs, fileName, ParentFolder)
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

		} else {
			allBytes = handler.GetFile().Buffer().Bytes()
		}

		_, err = w.Write(allBytes)
		if err != nil {
			panic(err)
		}
	}

	Gets["/ipfs/*cid"] = &getIpfs

	getVersion := func(w http2.ResponseWriter, r *http2.Request, ps httprouter.Params, q *http.Queue, fileIo *file_io_handler.FileIoHandler) {
		_, _ = w.Write([]byte("v0.0.0"))
	}

	Gets["/version"] = &getVersion

	http.StartServer(Gets, Posts, []string{"jipfs"})
}
