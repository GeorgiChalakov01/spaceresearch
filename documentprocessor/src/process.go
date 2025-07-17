package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/GeorgiChalakov01/spaceresearch/backend/app/core/common"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/core/minio"
)

// CallOllama performs streaming inference with Ollama API
// Returns a channel for streaming tokens and an error channel
func CallOllama(model, prompt string, imageBase64 string) (<-chan string, <-chan error) {
	tokenChan := make(chan string)
	errChan := make(chan error, 1) // Buffered to prevent blocking

	go func() {
		defer close(tokenChan)
		defer close(errChan)

		// Prepare the request body
		messages := []interface{}{
			map[string]interface{}{
				"role":	"user",
				"content": prompt,
			},
		}

		if imageBase64 != "" {
			messages[0].(map[string]interface{})["images"] = []string{imageBase64}
		}

		requestBody, err := json.Marshal(map[string]interface{}{
			"model":	model,
			"messages": messages,
			"stream":   true,
		})
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request body: %w", err)
			return
		}

		// Send the request to Ollama
		endpoint := os.Getenv("OLLAMA_CHAT_ENDPOINT")
		resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			errChan <- fmt.Errorf("request to Ollama failed: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("ollama API error: %s - %s", resp.Status, string(body))
			return
		}

		// Stream the response
		decoder := json.NewDecoder(resp.Body)
		for {
			var chunk struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				Response string `json:"response"`
				Done	 bool   `json:"done"`
			}

			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					break
				}
				errChan <- fmt.Errorf("failed to decode response chunk: %w", err)
				return
			}

			// Get content from either field
			content := chunk.Response
			if content == "" {
				content = chunk.Message.Content
			}

			if content != "" {
				tokenChan <- content
			}

			if chunk.Done {
				break
			}
		}
	}()

	return tokenChan, errChan
}

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
		tokenChan, errChan := CallOllama(model, prompt, imageInBase64)

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
	}
}
