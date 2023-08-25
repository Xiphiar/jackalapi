package japicore

import (
	"fmt"
	"sync"
	"time"

	"github.com/JackalLabs/jackalgo/handlers/file_io_handler"
	"github.com/JackalLabs/jackalgo/handlers/file_upload_handler"
	"github.com/JackalLabs/jackalgo/handlers/folder_handler"
)

type Queue struct {
	messages []*message
}

func NewQueue() *Queue {
	q := Queue{
		messages: make([]*message, 0),
	}
	return &q
}

type message struct {
	fileIo *file_io_handler.FileIoHandler
	upload *file_upload_handler.FileUploadHandler
	folder *folder_handler.FolderHandler
	wg     *sync.WaitGroup
	err    error
	fid    string
}

func (m *message) Error() error {
	return m.err
}

func (m *message) Fid() string {
	return m.fid
}

func (q *Queue) size() int {
	return len(q.messages)
}

func (q *Queue) isEmpty() bool {
	return len(q.messages) == 0
}

func (q *Queue) Push(upload *file_upload_handler.FileUploadHandler, folder *folder_handler.FolderHandler, fileIo *file_io_handler.FileIoHandler, wg *sync.WaitGroup) *message {
	m := message{
		fileIo: fileIo,
		upload: upload,
		folder: folder,
		wg:     wg,
	}

	q.messages = append(q.messages, &m)
	return &m
}

func (q *Queue) pop() *message {
	m := q.messages[0]
	q.messages = q.messages[1:]
	return m
}

func (q *Queue) listenOnce() {

	if q.isEmpty() {
		return
	}

	message := q.pop()

	_, fids, _, err := message.fileIo.StaggeredUploadFiles([]*file_upload_handler.FileUploadHandler{message.upload}, message.folder, false)

	fmt.Println(fids)

	message.err = err
	if len(fids) > 0 {
		message.fid = fids[0]
	}

	message.wg.Done()
}

func (q *Queue) Listen() {
	for {
		q.listenOnce()
		time.Sleep(time.Second * 5)
	}
}
