package file

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/hoangdv99/morgana/internal/configs"
	"github.com/hoangdv99/morgana/internal/utils"
	"github.com/minio/minio-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client interface {
	Write(ctx context.Context, filePath string) (io.WriteCloser, error)
	Read(ctx context.Context, filePath string) (io.ReadCloser, error)
}

func NewClient(
	downloadConfig configs.Download,
	logger *zap.Logger,
) (Client, error) {
	switch downloadConfig.Mode {
	case configs.DownloadModeLocal:
		return NewLocalClient(downloadConfig, logger)
	case configs.DownloadModeS3:
		return NewS3Client(downloadConfig, logger)
	default:
		return nil, fmt.Errorf("unsupported download mode: %s", downloadConfig.Mode)
	}
}

type bufferedFileReader struct {
	file           *os.File
	bufferedReader io.Reader
}

func newBufferedFileReader(file *os.File) io.ReadCloser {
	return &bufferedFileReader{
		file:           file,
		bufferedReader: bufio.NewReader(file),
	}
}

func (b bufferedFileReader) Close() error {
	return b.file.Close()
}

func (b bufferedFileReader) Read(p []byte) (n int, err error) {
	return b.bufferedReader.Read(p)
}

type LocalClient struct {
	downloadDirectory string
	logger            *zap.Logger
}

func NewLocalClient(downloadConfig configs.Download, logger *zap.Logger) (Client, error) {
	err := os.Mkdir(downloadConfig.DownloadDirectory, os.ModeDir)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("failed to create download directory: %w", err)
		}
	}

	return &LocalClient{
		downloadDirectory: downloadConfig.DownloadDirectory,
		logger:            logger,
	}, nil
}

func (l *LocalClient) Read(ctx context.Context, filePath string) (io.ReadCloser, error) {
	logger := utils.LoggerWithContext(ctx, l.logger).With(zap.String("file_path", filePath))

	absolutePath := path.Join(l.downloadDirectory, filePath)
	file, err := os.Open(absolutePath)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to open file")
		return nil, status.Error(codes.Internal, "failed to open file")
	}

	return newBufferedFileReader(file), nil
}

func (l *LocalClient) Write(ctx context.Context, filePath string) (io.WriteCloser, error) {
	logger := utils.LoggerWithContext(ctx, l.logger).With(zap.String("file_path", filePath))

	absolutePath := path.Join(l.downloadDirectory, filePath)
	file, err := os.Create(absolutePath)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to open file")
		return nil, status.Error(codes.Internal, "failed to open file")
	}

	return file, nil
}

type s3ClientReadWriteCloser struct {
	writtenData []byte
	isClosed    bool
}

func newS3ClientReadWriteCloser(
	ctx context.Context,
	minioClient *minio.Client,
	logger *zap.Logger,
	bucketName,
	ObjectName string,
) io.ReadWriteCloser {
	logger = utils.LoggerWithContext(ctx, logger)
	readWriteCloser := &s3ClientReadWriteCloser{
		writtenData: make([]byte, 0),
		isClosed:    false,
	}

	go func() {
		_, err := minioClient.PutObjectWithContext(ctx, bucketName, ObjectName, readWriteCloser, -1, minio.PutObjectOptions{})
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to put object")
		}
	}()

	return readWriteCloser
}

func (s *s3ClientReadWriteCloser) Close() error {
	s.isClosed = true
	return nil
}

func (s *s3ClientReadWriteCloser) Read(p []byte) (n int, err error) {
	if len(s.writtenData) > 0 {
		writtenLength := copy(p, s.writtenData)
		s.writtenData = s.writtenData[writtenLength:]
		return writtenLength, nil
	}

	if s.isClosed {
		return 0, io.EOF
	}

	return 0, nil
}

func (s *s3ClientReadWriteCloser) Write(p []byte) (n int, err error) {
	s.writtenData = append(s.writtenData, p...)
	return len(p), nil
}

type S3Client struct {
	minioClient *minio.Client
	bucket      string
	logger      *zap.Logger
}

func NewS3Client(
	downloadConfig configs.Download,
	logger *zap.Logger,
) (Client, error) {
	minioClient, err := minio.New(downloadConfig.Address, downloadConfig.Username, downloadConfig.Password, false)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create minio client")
		return nil, err
	}

	return &S3Client{
		minioClient: minioClient,
		bucket:      downloadConfig.Bucket,
		logger:      logger,
	}, nil
}

func (s S3Client) Read(ctx context.Context, filePath string) (io.ReadCloser, error) {
	logger := utils.LoggerWithContext(ctx, s.logger).With(zap.String("file_path", filePath))

	object, err := s.minioClient.GetObjectWithContext(ctx, s.bucket, filePath, minio.GetObjectOptions{})
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get s3 object")
		return nil, status.Error(codes.Internal, "failed to get s3 object")
	}

	return object, nil
}

func (s S3Client) Write(ctx context.Context, filePath string) (io.WriteCloser, error) {
	return newS3ClientReadWriteCloser(ctx, s.minioClient, s.logger, s.bucket, filePath), nil
}
