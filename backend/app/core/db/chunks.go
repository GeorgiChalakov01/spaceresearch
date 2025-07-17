package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/core/common"
	"github.com/pgvector/pgvector-go"
)

type ChunkRecord struct {
	Document	common.Document
	PageNumber	int
	OriginalText	string
	Embedding	pgvector.Vector
}

func InsertChunk(conn *pgx.Conn, record ChunkRecord) error {
	// Start a transaction
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return err
	}
	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "INSERT INTO chunks (user_id, object_name, page_number, bucket_name, chunk, embedding) VALUES ($1, $2, $3, $4, $5, $6)", record.Document.UserID, record.Document.ObjectName, record.PageNumber, record.Document.BucketName, record.OriginalText, record.Embedding)

	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}
