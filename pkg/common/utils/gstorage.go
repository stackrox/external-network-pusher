package utils

import (
	"context"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

const (
	// gCloudClientTimeout is the timeout value we use for
	// clients connected to Google Cloud
	gCloudClientTimeout = 3 * time.Minute
)

// WriteToBucket writes the specified content to the specified bucket.
//  - bucketName: Name of the bucket this fn writes to
//  - objectPrefix: Prefix of the file. EX: folder1
//  - objectName: Name for the file that this fn creates for writing out data.
//  - data: Content of the file.
func WriteToBucket(
	bucketName string,
	objectPrefix string,
	objectName string,
	data []byte,
) error {
	objectPath := filepath.Join(objectPrefix, objectName)

	ctx, cancel := context.WithTimeout(context.Background(), gCloudClientTimeout)
	defer cancel()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	writer := client.Bucket(bucketName).Object(objectPath).NewWriter(ctx)
	if _, err = writer.Write(data); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}

// DeleteObjectWithPrefix deletes any object in the specified bucket that
// starts with the specified prefix.
func DeleteObjectWithPrefix(bucketName, prefix string) error {
	ctx, cancel := context.WithTimeout(context.Background(), gCloudClientTimeout)
	defer cancel()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	bucket := client.Bucket(bucketName)
	query := &storage.Query{Prefix: prefix}

	var names []string
	it := bucket.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			errors.Wrapf(err, "failed while trying to list and traverse objects in bucket %s", bucketName)
			return err
		}
		names = append(names, attrs.Name)
	}

	for _, name := range names {
		err = bucket.Object(name).Delete(ctx)
		if err != nil {
			errors.Wrapf(
				err,
				"failed to delete object with name %s under bucket %s. Please clean up manually",
				name,
				bucketName)
			// Return early
			return err
		}
	}

	return nil
}