package upload

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"encoding/json"

	"github.com/jackc/pgx/v5"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/GeorgiChalakov01/spaceresearch/core/minio"
	"github.com/GeorgiChalakov01/spaceresearch/core/common"
	"github.com/GeorgiChalakov01/spaceresearch/core/rabbitmq"
)

func ProcessUpload(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
	// Parse multipart form (max 100MB files)
	const maxMemory = 100 << 20 // 100MB
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get user from session
	cookie, err := r.Cookie("user_email")
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	email := cookie.Value

	user, err := common.GetUserData(conn, email)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Initialize MinIO client
	minioClient, err := minio.NewMinioClient()
	if err != nil {
		http.Error(w, "Failed to initialize MinIO client", http.StatusInternalServerError)
		return
	}

	// Get files from form
	files := r.MultipartForm.File["files"]
	bucketName := "documents"

	var uploadError bool
	for _, fileHeader := range files {
		var currentDocument common.Document
		// Open file
		file, err := fileHeader.Open()
		if err != nil {
			fmt.Printf("Error opening file: %v", err)
			continue
		}
		defer file.Close()

		// Create temp file
		tempDir := os.TempDir()
		tempFilePath := filepath.Join(tempDir, fileHeader.Filename)
		tempFile, err := os.Create(tempFilePath)
		if err != nil {
			fmt.Printf("Error creating temp file: %v", err)
			continue
		}
		defer os.Remove(tempFilePath)
		defer tempFile.Close()

		// Copy to temp file
		if _, err := io.Copy(tempFile, file); err != nil {
			fmt.Printf("Error copying to temp file: %v", err)
			continue
		}

		// Generate unique object name
		timestamp := time.Now().Unix()
		objectName := fmt.Sprintf("%d_%s", timestamp, fileHeader.Filename)

		// Upload to MinIO
		ctx := context.Background()
		if err := minio.UploadFile(ctx, minioClient, bucketName, objectName, tempFilePath, "application/pdf"); err != nil {
			fmt.Printf("Error uploading file [%s] to MinIO: %v", objectName, err)
			uploadError = true

			return
		}

		fmt.Printf("Successfully uploaded %s to bucket %s", objectName, bucketName)

		currentDocument.ObjectName = objectName
		currentDocument.BucketName = bucketName

		// Add record to database
		documentRecord := common.Document{
			UserID:		user.Id,
			FileName:	fileHeader.Filename,
			ObjectName:	objectName,
			BucketName:	bucketName,
			UploadedAt:	time.Now(),
		}
		id, err := common.InsertDocument(conn, documentRecord)
		if err != nil {
			fmt.Printf("Error storing the uploaded file's metadata in the DB: %v", err)
			uploadError = true
			// Delete the uploaded file from MinIO as metadata for it was not stored.
			minio.DeleteFile(ctx, minioClient, currentDocument.BucketName, currentDocument.ObjectName)
			return
		}

		currentDocument.ID = id

		// Connect to RabbitMQ
		conn, err := rabbitmq.Connect()
		if err != nil {
			fmt.Println("Could not connect to RabbitMQ. Error:\n%v", err)
			http.Redirect(w, r, "/err?error=rabbitmqConnectionError", http.StatusSeeOther)
			return
		}
		defer conn.Close()

		// Create a RabbitMQ channel
		ch, err := conn.Channel()
		if err != nil {
			fmt.Printf("\nCould not create a RabbitMQ channel. Error:\n%v\n", err)
			http.Redirect(w, r, "/err?error=rabbitmqCreateChannelError", http.StatusSeeOther)
			return
		}
		defer ch.Close()

		// Create a RabbitMQ queue
		q, err := ch.QueueDeclare(
			"documents",	// name
			true,		// durable
			false,		// delete when unused
			false,		// exclusive
			false,		// no-wait
			nil,		// arguments
		)
		if err != nil {
			fmt.Printf("\nCould not create a RabbitMQ queue. Error:\n%v\n", err)
			http.Redirect(w, r, "/err?error=rabbitmqCreateQueueError", http.StatusSeeOther)
			return
		}

		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// JSON-ify the current document object
		body, err := json.Marshal(currentDocument)
		if err != nil {
			fmt.Printf("\nCould not create a RabbitMQ queue. Error:\n%v\n", err)
			http.Redirect(w, r, "/err?error=rabbitmqCreateQueueError", http.StatusSeeOther)
			return
		}
		
		err = ch.PublishWithContext(ctx,
			"",		// exchange
			q.Name,		// routing key
			false,		// mandatory
			false,		// immediate
			amqp.Publishing {
				ContentType:	"application/json",
				Body:		body,
		})
		if err != nil {
			fmt.Printf("\nCould not publish message to RabbitMQ queue. Error:\n%v\n", err)
			http.Redirect(w, r, "/err?error=rabbitmqCreateQueueError", http.StatusSeeOther)
			return
		}

		fmt.Printf("\nPublished a message [%s] to queue [%s]\n", body, q.Name)
	}
	var msg string
	if uploadError {
		msg = "error=uploadFailedForOneOrMoreFiles"
	} else {
		msg = "success=uploadSuccessful"
	}

	http.Redirect(w, r, "/home?" + msg, http.StatusSeeOther)
}
