package storage

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"io"
	"path"
	"strings"
	"template-go/util/logger"
	"template-go/util/trace"
)

var _ Storage = (*AwsStorage)(nil)

type AwsStorage struct {
	logger   logger.Logger
	bucket   string
	basePath string
	baseUrl  string
	client   *s3.Client
}

func NewAwsStorage(region string, bucket string, basePath string, baseUrl string) *AwsStorage {
	// Initialize a session using AWS SDK
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))

	if err != nil {
		panic(err)
	}

	basePath = strings.Trim(basePath, "/")

	baseUrl = strings.Trim(baseUrl, "/")

	// Create S3 service client
	client := s3.NewFromConfig(cfg)

	return &AwsStorage{
		logger:   logger.NewZerologLogger("AwsStorage"),
		bucket:   bucket,
		basePath: basePath,
		baseUrl:  baseUrl,
		client:   client,
	}
}

func (s *AwsStorage) UploadAsRandom(
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
) {
	fileName := uuid.New().String() + "." + ext
	return s.Upload(ctx, trace, src, fileName, filePath, contentType)
}

func (s *AwsStorage) Upload(
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
) {
	objectKey := path.Join(s.basePath, filePath, fileName)

	out, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
		Body:   src,

		// If we don't set content-type, AWS will set it to application/octet-stream, which will make browsers visiting
		// the URL automatically downloads the file.
		ContentType: aws.String(contentType),
	})
	if err != nil {
		s.logger.ErrorErr(trace, err).MarshalJson("out", out).Msg("Upload File")
		return "", "", err
	}

	relativePath = strings.Trim(path.Join(filePath, fileName), "/")
	absoluteUrl = s.baseUrl + "/" + relativePath

	return absoluteUrl, relativePath, nil
}
