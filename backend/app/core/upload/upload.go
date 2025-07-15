package upload

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/jackc/pgx/v5"

    _ "github.com/GeorgiChalakov01/spaceresearch/core/common"
    "github.com/GeorgiChalakov01/spaceresearch/core/minio"
)

func ProcessUpload(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
    // Parse multipart form (max 100MB files)
    const maxMemory = 100 << 20 // 100MB
    if err := r.ParseMultipartForm(maxMemory); err != nil {
        http.Error(w, "Unable to parse form", http.StatusBadRequest)
        return
    }

    // Get user from session
    // cookie, err := r.Cookie("user_email")
    // if err != nil {
    //     http.Error(w, "User not authenticated", http.StatusUnauthorized)
    //     return
    // }
    // email := cookie.Value

    // user, err := common.GetUserData(conn, email)
    // if err != nil {
    //     http.Error(w, "User not found", http.StatusInternalServerError)
    //     return
    // }

    // Initialize MinIO client
    minioClient, err := minio.NewMinioClient()
    if err != nil {
        http.Error(w, "Failed to initialize MinIO client", http.StatusInternalServerError)
        return
    }

    // Get files from form
    files := r.MultipartForm.File["files"]
    bucketName := "documents"

    for _, fileHeader := range files {
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
            fmt.Printf("Error uploading file to MinIO: %v", err)
            http.Error(w, "Error uploading file", http.StatusInternalServerError)
            return
        }

        fmt.Printf("Successfully uploaded %s to bucket %s", objectName, bucketName)

        // Add record to database (placeholder)
        // fileRecord := common.File{
        // 	UserID:     user.Id,
        // 	FileName:   fileHeader.Filename,
        // 	ObjectName: objectName,
        // 	BucketName: bucketName,
        // 	UploadedAt: time.Now(),
        // }
        // if err := common.InsertFile(conn, fileRecord); err != nil {
        // 	fmt.Printf("Error inserting file record: %v", err)
        // }

        // Create RabbitMQ task (placeholder)
        // rabbitmq.PublishProcessingTask(objectName)
    }

    http.Redirect(w, r, "/home?success=uploadSuccess", http.StatusSeeOther)
}
