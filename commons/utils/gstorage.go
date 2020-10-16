package utils

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

const (
	GCloudClientTimeout = 3 * time.Minute
)

func WriteToBucket(
	bucketName string,
	objectPrefix string,
	objectName string,
	data []byte,
) error {
	objectPath := filepath.Join(objectPrefix, objectName)

	ctx, cancel := context.WithTimeout(context.Background(), GCloudClientTimeout)
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

func DeleteObjectWithPrefix(bucketName string, prefix string) error {
	ctx, cancel := context.WithTimeout(context.Background(), GCloudClientTimeout)
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
			log.Fatal(err)
		}
		names = append(names, attrs.Name)
	}

	for _, name := range names {
		err = bucket.Object(name).Delete(ctx)
		if err != nil {
			log.Printf(
				"Failed to delete object with name %s under bucket %s. Please clean up manually",
				name,
				bucketName)
			// Return early
			return err
		}
	}

	return nil
}

