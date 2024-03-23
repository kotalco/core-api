package security

import (
	"fmt"
	"io"
	"mime/multipart"
)

func ValidateFile(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("couldn't open file: %w", err)
	}
	defer file.Close()

	certBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("couldn't read cert file: %w", err)
	}
	return certBytes, nil
}
