package nethttplibrary

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"

	abstractions "github.com/microsoft/kiota/abstractions/go"
)

// CompressionHandler represents a compression middleware
type CompressionHandler struct {
	options CompressionOptions
}

// CompressionOptions is a configuration object for the CompressionHandler middleware
type CompressionOptions struct {
	enableCompression bool
}

type compression interface {
	abstractions.RequestOption
	ShouldCompress() bool
}

var compressKey = abstractions.RequestOptionKey{Key: "CompressionHandler"}

// NewCompressionHandler creates an instance of a compression middleware
func NewCompressionHandler() *CompressionHandler {
	options := NewCompressionOptions(true)
	return NewCompressionHandlerWithOptions(options)
}

// NewCompressionHandlerWithOptions creates an instance of the compression middlerware with
// specified configurations.
func NewCompressionHandlerWithOptions(option CompressionOptions) *CompressionHandler {
	return &CompressionHandler{options: option}
}

// NewCompressionOptions creates a configuration object for the CompressionHandler
func NewCompressionOptions(enableCompression bool) CompressionOptions {
	return CompressionOptions{enableCompression: enableCompression}
}

// GetKey returns CompressionOptions unique name in context object
func (o CompressionOptions) GetKey() abstractions.RequestOptionKey {
	return compressKey
}

// ShouldCompress reads compression setting form CompressionOptions
func (o CompressionOptions) ShouldCompress() bool {
	return o.enableCompression
}

// Intercept is invoked by the middleware pipeline to either move the request/response
// to the next middleware in the pipeline
func (c *CompressionHandler) Intercept(pipeline Pipeline, middlewareIndex int, req *http.Request) (*http.Response, error) {
	reqOption, ok := req.Context().Value(compressKey).(compression)
	if !ok {
		reqOption = c.options
	}

	if !reqOption.ShouldCompress() || req.Body == nil {
		return pipeline.Next(req, middlewareIndex)
	}

	unCompressedBody, err := ioutil.ReadAll(req.Body)
	unCompressedContentLength := req.ContentLength
	if err != nil {
		return nil, err
	}

	compressedBody, size, err := compressReqBody(unCompressedBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Encoding", "gzip")
	req.Body = compressedBody
	req.ContentLength = int64(size)

	// Sending request with compressed body
	resp, err := pipeline.Next(req, middlewareIndex)
	if err != nil {
		return nil, err
	}

	// If response has status 415 retry request with uncompressed body
	if resp.StatusCode == 415 {
		delete(req.Header, "Content-Encoding")
		req.Body = ioutil.NopCloser(bytes.NewBuffer(unCompressedBody))
		req.ContentLength = unCompressedContentLength

		return pipeline.Next(req, middlewareIndex)
	}

	return resp, nil
}

func compressReqBody(reqBody []byte) (io.ReadCloser, int, error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	if _, err := gzipWriter.Write(reqBody); err != nil {
		return nil, 0, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, 0, err
	}

	return ioutil.NopCloser(&buffer), buffer.Len(), nil
}