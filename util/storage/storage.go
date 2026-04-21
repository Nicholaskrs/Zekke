package storage

import (
	"context"
	"io"
	"template-go/util/trace"
)

type Storage interface {
	UploadAsRandom(
		ctx context.Context,
		trace *trace.Trace,
		src io.Reader,
		filePath string,
		ext string,
		contentType string,
	) (
		absoluteUrl string,
		relativePath string,
		err error,
	)

	Upload(
		ctx context.Context,
		trace *trace.Trace,
		src io.Reader,
		fileName string,
		filePath string,
		contentType string,
	) (
		absoluteUrl string,
		relativePath string,
		err error,
	)
}
