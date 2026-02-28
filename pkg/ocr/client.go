package ocr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: time.Minute * 2, // OCR can take some time
		},
	}
}

type OCRResponse struct {
	Text string `json:"text"`
}

// ExtractText takes an absolute path to an image file and sends it to the OCR service
func (c *Client) ExtractText(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("could not create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("could not copy file content: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("could not close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/extract", body)
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to OCR service failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OCR service returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var ocrResp OCRResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocrResp); err != nil {
		return "", fmt.Errorf("could not decode OCR response: %w", err)
	}

	return ocrResp.Text, nil
}
