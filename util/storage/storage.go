package storage

import (
	"context"
	"io"
	"template-go/util/logtrace"
)

type Storage interface {
	UploadAsRandom(
		ctx context.Context,
		trace *logtrace.LogTrace,
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
		trace *logtrace.LogTrace,
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
