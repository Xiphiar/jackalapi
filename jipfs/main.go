package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	jhttp "github.com/JackalLabs/jackalapi/jhttp/http"
	"github.com/JackalLabs/jackalapi/jipfs/ipfs"
	"github.com/JackalLabs/jackalgo/handlers/file_upload_handler"
	"github.com/rs/cors"
	"github.com/uptrace/bunrouter"
	"io"
	"net/http"
	url2 "net/url"
	"os"
	"strconv"
	"strings"
	"sync"
)

const ParentFolder = "s/jipfs"

func main() {

	h, err := ipfs.MakeHost(0, 0)
	if err != nil {
		panic(err)
	}
	defer h.Close()

	fmt.Println("Starting jIPFS...")

	port := os.Getenv("JHTTP_PORT")
	if len(port) == 0 {
		port = "3535"
	}

	portNum, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		panic(err)
	}

	router := bunrouter.New()

	q, fileIo := jhttp.InitServer([]string{"stratus"})

	router.GET("/version", func(w http.ResponseWriter, r bunrouter.Request) error {
		_, err := w.Write([]byte("v0.0.0"))
		if err != nil {
			return err
		}
		return nil
	})

	router.GET("/ls", func(w http.ResponseWriter, r bunrouter.Request) error {

		folder, err := fileIo.DownloadFolder(ParentFolder)
		if err != nil {
			return err
		}

		children := folder.GetChildFiles()

		childrenJson, err := json.Marshal(children)
		if err != nil {
			return err
		}

		w.Write(childrenJson)

		return nil
	})

	router.GET("/ipfs/*cid", func(w http.ResponseWriter, r bunrouter.Request) error {

		cid := r.Param("cid")
		if len(cid) == 0 {
			w.WriteHeader(500)
			return errors.New("failed to get cid")
		}
		fileName := strings.ReplaceAll(cid, "/", "_")

		var allBytes []byte

		handler, err := fileIo.DownloadFile(fmt.Sprintf("%s/%s", ParentFolder, fileName))
		if err != nil {
			url, err := url2.Parse("https://ipfs.io/ipfs/")
			if err != nil {
				fmt.Println("urlparse failed", err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return err
			}

			url = url.JoinPath(cid)
			fmt.Println(url.String())

			req, err := http.NewRequest("GET", url.String(), nil)
			if err != nil {
				fmt.Println("fileupload failed", err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return err
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println("fileupload failed", err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return err
			}

			var b bytes.Buffer
			size, err := io.Copy(&b, res.Body)
			if err != nil {
				fmt.Println("fileupload failed", err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return err
			}
			fmt.Printf("saving file with size of %d.\n", size)
			err = res.Body.Close()
			if err != nil {
				fmt.Println("fileupload failed", err)
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return err
			}

			allBytes = b.Bytes()
			bs := make([]byte, len(allBytes))
			copy(bs, allBytes)

			fileUpload, err := file_upload_handler.TrackVirtualFile(bs, fileName, ParentFolder)
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

		} else {
			allBytes = handler.GetFile().Buffer().Bytes()
		}

		_, err = w.Write(allBytes)
		if err != nil {
			return err
		}

		return nil
	})

	handler := cors.Default().Handler(router)

	fmt.Printf("🌍 Started jIPFS: http://0.0.0.0:%d\n", portNum)
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
