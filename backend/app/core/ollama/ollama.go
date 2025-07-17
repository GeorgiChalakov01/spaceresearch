package ollama

import (
	"bytes"
	"time"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)


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

func GetEmbedding(text string) ([]float32, error) {
	model := os.Getenv("OLLAMA_EMBEDDING_MODEL")
	if model == "" {
		model = "nomic-embed-text"
	}
	
	apiURL := os.Getenv("OLLAMA_EMBEDDING_ENDPOINT")
	if apiURL == "" {
		apiURL = "http://localhost:11434/api/embeddings"
	}

	// Create request payload
	payload := map[string]string{
		"model":  model,
		"prompt": text,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request with timeout
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error [%d]: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var embeddingResp struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return embeddingResp.Embedding, nil
}
