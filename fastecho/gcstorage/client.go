package gcstorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"

	"github.com/ingka-group-digital/ocp-go-utils/stringutils"
)

var (
	StorageObjectNotFound = errors.New("storage: object not found")
)

// Bucket is the interface defining the interaction with GCS buckets
type Bucket interface {
	Read(ctx context.Context, bucketName, filePath string) ([]byte, error)
	ReadStream(ctx context.Context, bucketName, filePath string) (*storage.Reader, error)
	Write(ctx context.Context, bucketName, filePath string, data []byte) error
	WriteStream(ctx context.Context, bucketName, filePath string) *storage.Writer
	Delete(ctx context.Context, bucketName, filePath string) error
}

// Client is the struct that implements the Bucket interface, holding the connection to the GCS
type Client struct {
	client *storage.Client
}

// New creates a new GCS Client
func New(ctx context.Context) (*Client, error) {
	sc, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	return &Client{
		client: sc,
	}, nil
}

// ReadStream returns a reader for the data present on a cloud bucket
func (c *Client) ReadStream(ctx context.Context, bucketName, filePath string) (*storage.Reader, error) {
	reader, err := c.client.Bucket(bucketName).Object(filePath).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create a reader: %w", err)
	}

	return reader, nil
}

// Read returns data present on a cloud bucket
func (c *Client) Read(ctx context.Context, bucketName, filePath string) ([]byte, error) {
	reader, err := c.ReadStream(ctx, bucketName, filePath)
	if err != nil {
		return nil, fmt.Errorf("could not create a reader: %w", err)
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("could not read file from bucket: %w", err)
	}

	if err := reader.Close(); err != nil {
		return nil, fmt.Errorf("failed to close the reader: %w", err)
	}

	return bytes, nil
}

// WriteStream returns a *storage.Writer for the data present on a cloud bucket
func (c *Client) WriteStream(ctx context.Context, bucketName, filePath string) *storage.Writer {
	return c.client.Bucket(bucketName).Object(filePath).NewWriter(ctx)
}

// Write stores the given data to a cloud bucket
func (c *Client) Write(ctx context.Context, bucketName, filePath string, data []byte) error {
	writer := c.WriteStream(ctx, bucketName, filePath)
	writer.ContentType = "application/json"

	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("could not write to bucket: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close the writer: %w", err)
	}

	return nil
}

// Delete removes the data present on a cloud bucket
func (c *Client) Delete(ctx context.Context, bucketName, filePath string) error {
	err := c.client.Bucket(bucketName).Object(filePath).Delete(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return StorageObjectNotFound
		}

		return fmt.Errorf("could not delete from bucket: %w", err)
	}
	return nil
}

// DeconstructGsURI receives a gs:// URI and returns the bucket name and the filepath
//
// Split the string like the one below based on /. slice[2] contains the bucket name and the rest is the file path
// gs://ocp-mlflow-dev/xfer/weekly-forecasts/output/scheduled__2022-01-25T22:00:00+00:00/fcp-baseline-latest-NL.json
func DeconstructGsURI(gsURI string) (string, string, error) {
	r, err := regexp.Compile("^(gs://)(.*)")
	if err != nil {
		return "", "", fmt.Errorf("failed to compile regex")
	}

	results := r.FindAllStringSubmatch(gsURI, -1)
	if results == nil {
		return "", "", fmt.Errorf("invalid gs:// URI string: %s", gsURI)
	}

	// results[0][0] contains the original gsURI
	// results[0][1] contains only the gs://
	// results[0][2] contains the part after gs://
	gsUriParts := strings.Split(results[0][2], "/")
	if len(gsUriParts) == 1 {
		return "", "", fmt.Errorf("invalid gs:// URI, cannot split to parts: %s", gsURI)
	}

	// First split contains the bucket name
	bucketName := gsUriParts[0]

	// The rest of the split is the filepath
	filePath := strings.Join(gsUriParts[1:], "/")

	if stringutils.IsEmpty(gsUriParts[0]) || stringutils.IsEmpty(filePath) {
		return "", "", fmt.Errorf("invalid gs uri, cannot identify bucket name and filepath: %s", gsURI)
	}

	return bucketName, filePath, nil
}

// DeconstructPath separates the directories from the filename
func DeconstructPath(filePath string) (string, string, error) {
	parts := strings.Split(filePath, "/")
	if len(parts) == 1 {
		return "", "", fmt.Errorf("invalid file path, cannot split to parts: %s", filePath)
	}

	// The rest of the split is the location
	location := strings.Join(parts[0:len(parts)-1], "/")

	// Last split contains the file name
	fileName := parts[len(parts)-1]

	return location, fileName, nil
}

// ConstructGsURI generates a GS URI out of the bucket name and the filepath (directories + filename)
func ConstructGsURI(bucketName, filePath string) string {
	return fmt.Sprintf("gs://%s/%s", bucketName, filePath)
}
