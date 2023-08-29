package japicore

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func httpGetFileRequest(w http.ResponseWriter, host string, path string) (bytes.Buffer, error) {
	var byteBuffer bytes.Buffer

	bytes.NewReader(byteBuffer.Bytes())

	url, err := url.Parse(host)
	if err != nil {
		processHttpPostError("UrlParse", err, w)
		return byteBuffer, err
	}

	url = url.JoinPath(path)
	fmt.Println(url.String())

	innerReq, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		processHttpPostError("CreateGetRequest", err, w)
		return byteBuffer, err
	}

	res, err := http.DefaultClient.Do(innerReq)
	if err != nil {
		processHttpPostError("UseGetRequest", err, w)
		return byteBuffer, err
	}

	_, err = io.Copy(&byteBuffer, res.Body)
	if err != nil {
		processHttpPostError("BufferCopy", err, w)
		return byteBuffer, err
	}

	err = res.Body.Close()
	if err != nil {
		processHttpPostError("BodyClose", err, w)
		return byteBuffer, err
	}

	return byteBuffer, nil
}
