package sumologic

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"sync"
)

const (
	sourceNameHeader     = "X-Sumo-Name"
	sourceHostHeader     = "X-Sumo-Host"
	sourceCategoryHeader = "X-Sumo-Category"
	compressionHeader    = "Content-Encoding"
	compressionType      = "gzip"
)

type uploader struct {
	batches chan *Batch
	wg      sync.WaitGroup
}

func (u *uploader) upload(b *Batch) error {
	body, err := u.createBuffer(b)
	if err != nil {
		return err
	}
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

func (u *uploader) createBuffer(b *Batch) (*bytes.Buffer, error) {
	var buffer []byte
	for _, e := range b.entries {
		buffer = append(buffer, e.record...)
		buffer = append(buffer, "\n"...)
	}
	return u.compressBuffer(buffer)
}

func (u *uploader) compressBuffer(b []byte) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	c, err := gzip.NewWriterLevel(
		buffer,
		gzip.BestCompression,
	)
	c.Write(b)
	c.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to compress the data, %w", err)

	}
	return buffer, nil
}

func (u *uploader) setHeaders(b *Batch, r *http.Request) {
	if b.config.sourceName != "" {
		r.Header.Add(sourceNameHeader, b.config.sourceName)
	}
	r.Header.Add(sourceHostHeader, b.config.sourceHost)
	r.Header.Add(sourceCategoryHeader, b.config.sourceCategory)
	r.Header.Add(compressionHeader, compressionType)
}
