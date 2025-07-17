package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/gen2brain/go-fitz"
	"github.com/heussd/pdftotext-go"
)

type PageContent struct {
	PageNumber int    `json:"page"`
	Content    string `json:"content"`
}

func ExtractTextFromPDF(pdfBytes []byte) ([]PageContent, error) {
	pages, err := pdftotext.ExtractOrError(pdfBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %v", err)
	}

	var result []PageContent
	for _, page := range pages {
		result = append(result, PageContent{
			PageNumber: page.Number,
			Content:    page.Content,
		})
	}

	return result, nil
}

func PDFPageToPNG(pdfBytes []byte, pageNumber int) (string, error) {
	doc, err := fitz.NewFromMemory(pdfBytes)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %v", err)
	}
	defer doc.Close()

	// Render page to image
	img, err := doc.Image(pageNumber - 1) // Page numbers are 0-indexed
	if err != nil {
		return "", fmt.Errorf("failed to render page %d: %v", pageNumber, err)
	}

	// Convert to PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return "", fmt.Errorf("failed to encode PNG: %v", err)
	}

	// Convert to base64
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
