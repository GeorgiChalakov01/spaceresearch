package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pgvector/pgvector-go"

	"github.com/GeorgiChalakov01/spaceresearch/backend/app/core/db"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/core/ollama"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/core/common"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/core/minio"
)


func ProcessDocument(details []byte) {
	var document common.Document
	err := json.Unmarshal(details, &document)
	if err != nil {
		fmt.Printf("\nCould not unmarshal the record. \nError: %v\n", err)
		return
	}

	// Initialize MinIO client
	minioClient, err := minio.NewMinioClient()
	if err != nil {
		fmt.Printf("Failed to initialize MinIO client. \nError: %v\n", err)
		return
	}

	ctx := context.Background()

	// Get file from MinIO
	fileBytes, err := minio.GetFileContent(ctx, minioClient, document.BucketName, document.ObjectName)
	if err != nil {
		fmt.Printf("Failed to get file from MinIO. \nError: %v\n", err)
		return
	}

	// Extract text from PDF
	fmt.Printf("Extracting text from %s/%s...\n", document.BucketName, document.ObjectName)
	pageContents, err := ExtractTextFromPDF(fileBytes)
	if err != nil {
		fmt.Printf("Failed to extract text from PDF. \nError: %v\n", err)
		return
	}

	// Get model from environment or use default
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "gemma3:4b-it-qat" // Default model
	}

	// Process each page
	for _, page := range pageContents {
		fmt.Printf("Generating the visual explanation of page %d...\n", page.PageNumber)
		
		// Convert PDF page to PNG
		fmt.Printf("Converting PDF page to PNG...\n")
		imageInBase64, err := PDFPageToPNG(fileBytes, page.PageNumber)
		if err != nil {
			fmt.Printf("Failed to convert PDF page [%d] to PNG. \nError: %v\n", page.PageNumber, err)
			return
		}

		// Prepare the prompt
		prompt := fmt.Sprintf(
			"Please explain in detail what you see in the provided page of a NASA document. "+
				"Your explanation will be stored as an official document so don't ask questions or talk to me. "+
				"Start with 'The document ...' and continue with a description of what it looks like. "+
				"Explain any graphics, drawings, formulas, tables or other non textual elements you see.\n"+
				"Here is a transcription of the text in the document:\n%s",
			page.Content,
		)

		// Call Ollama with streaming
		fmt.Printf("Doing inference on Ollama with model %s...\n", model)
		tokenChan, errChan := ollama.CallOllama(model, prompt, imageInBase64)

		// Collect tokens and build full response
		var fullResponse strings.Builder
		for {
			select {
			case token, ok := <-tokenChan:
				if !ok {
					// Token channel closed, processing complete
					goto ProcessChunk
				}
				fullResponse.WriteString(token)
			case err := <-errChan:
				if err != nil {
					fmt.Printf("Ollama inference failed: %v\n", err)
					return
				}
			}
		}

	ProcessChunk:
		explanationOfVisuals := fullResponse.String()

		// Prepare the chunks
		chunk := "********************\n"
		chunk += "Document: " + document.ObjectName + " Page: " + strconv.Itoa(page.PageNumber) + "\n"
		chunk += "====================\n"
		chunk += "Page Textual Content:\n"
		chunk += page.Content + "\n"
		chunk += "====================\n"
		chunk += "Page Visual Explanation:\n"
		chunk += explanationOfVisuals + "\n"
		chunk += "********************\n"
		fmt.Println(chunk)
		

		// Generate Embedding
		fmt.Println("Generating embedding...")
		embeddingFA, err := ollama.GetEmbedding(chunk)
		if err != nil {
			fmt.Printf("Could not generate embedding. Error:\n%v\n", err)
			return
		}
		embedding := pgvector.NewVector(embeddingFA)

		// Connect to db
		conn, err := db.Connect()
		if err != nil {
			fmt.Printf("Could not connect to DB. Error:\n%v\n", err)
			return
		}
		defer conn.Close(context.Background())
		
		// Build the record
		record := db.ChunkRecord {
			Document: document,
			PageNumber: page.PageNumber,
			OriginalText: chunk,
			Embedding: embedding,
		}

		// Insert record in the Chunks table
		err = db.InsertChunk(conn, record)
		if err != nil {
			fmt.Printf("Error storing the chunk in the database.Error:\n%v\n", err)
		}
	}
}
