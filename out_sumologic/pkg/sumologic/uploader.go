package sumologic

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
)

const (
	sourceNameHeader     = "X-Sumo-Name"
	sourceHostHeader     = "X-Sumo-Host"
	sourceCategoryHeader = "X-Sumo-Category"
)

type uploader struct {
	batches chan *Batch
	wg      sync.WaitGroup
}

func (u *uploader) upload(b *Batch) error {
	body := u.createBuffer(b)
	request, err := http.NewRequest(
		http.MethodPost,
		b.config.collectorURL,
		body,
	)
	if err != nil {
		return fmt.Errorf("unable to create the request, %w", err)
	}

	u.setHeaders(b, request)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send the request, %w", err)
	} else if response.StatusCode != 200 {
		return fmt.Errorf("request failed with error, %w", err)
	}
	return nil
}

func (u *uploader) createBuffer(b *Batch) *bytes.Reader {
	var buffer []byte
	for _, e := range b.entries {
		buffer = append(buffer, e.record...)
		buffer = append(buffer, "\n"...)
	}
	reader := bytes.NewReader(buffer)
	return reader
}

func (u *uploader) setHeaders(b *Batch, r *http.Request) {
	r.Header.Add(sourceNameHeader, b.config.sourceName)
	r.Header.Add(sourceHostHeader, b.config.sourceHost)
	r.Header.Add(sourceCategoryHeader, b.config.sourceCategory)
}
