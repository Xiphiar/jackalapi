package japicore

import (
	"fmt"
	"sync"
	"time"

	"github.com/JackalLabs/jackalgo/handlers/file_upload_handler"
)

type ScrapeQueue struct {
	scrapees []*scrapee
}

func NewScrapeQueue() *ScrapeQueue {
	q := ScrapeQueue{
		scrapees: make([]*scrapee, 0),
	}
	return &q
}

type scrapee struct {
	filename string
	host     string
	wg       *sync.WaitGroup
	err      error
}

func (m *scrapee) Error() error {
	return m.err
}

func (m *scrapee) Filename() string {
	return m.filename
}

func (m *scrapee) Host() string {
	return m.host
}

func (q *ScrapeQueue) size() int {
	return len(q.scrapees)
}

func (q *ScrapeQueue) isEmpty() bool {
	return len(q.scrapees) == 0
}

func (q *ScrapeQueue) Push(filename string, host string, wg *sync.WaitGroup) *scrapee {
	m := scrapee{
		filename: filename,
		host:     host,
		wg:       wg,
	}

	q.scrapees = append(q.scrapees, &m)
	return &m
}

func (q *ScrapeQueue) pop() *scrapee {
	m := q.scrapees[0]
	q.scrapees = q.scrapees[1:]
	return m
}

func (q *ScrapeQueue) listenOnce() {

	if q.isEmpty() {
		return
	}

	scrapee := q.pop()

	_, fids, _, err := message.fileIo.StaggeredUploadFiles([]*file_upload_handler.FileUploadHandler{message.upload}, message.folder, false)

	fmt.Println(fids)

	message.err = err
	if len(fids) > 0 {
		message.fid = fids[0]
	}

	message.wg.Done()
}

func (q *ScrapeQueue) Listen() {
	for {
		q.listenOnce()
		time.Sleep(time.Second * 5)
	}
}
