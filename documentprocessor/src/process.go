package main

import (
	"fmt"
	"encoding/json"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/core/common"
)

func ProcessDocument(details []byte) {
	var document common.Document
	err := json.Unmarshal(details, &document)
	if err != nil {
		fmt.Printf("\nCould not unmarshal the record. \nError: %v\n")
		return
	}

	fmt.Printf("\n%s\n", document.ObjectName)
}
