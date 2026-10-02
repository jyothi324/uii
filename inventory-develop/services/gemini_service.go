package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type GeminiImageResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				InlineData *struct {
					Data     string `json:"data"`
					MimeType string `json:"mimeType"`
				} `json:"inlineData,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func fileToBase64(fileHeader *multipart.FileHeader) (string, string, error) {
	if fileHeader == nil {
		return "", "", nil
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", "", err
	}

	return base64.StdEncoding.EncodeToString(data), fileHeader.Header.Get("Content-Type"), nil
}

func ProcessGeminiImage(image, texture *multipart.FileHeader, target, notes string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-image:generateContent?key=%s",
		apiKey,
	)

	// Convert files
	base64Image, mimeType, err := fileToBase64(image)
	if err != nil {
		return "", err
	}

	textureBase64, textureMime, _ := fileToBase64(texture)

	// Build prompt
	prompt := "You are a professional interior designer. Retain layout."

	if textureBase64 != "" {
		prompt += fmt.Sprintf("\nReplace %s using texture reference.", target)
	}

	if notes != "" {
		prompt += fmt.Sprintf("\nApply changes: %s", notes)
	}

	// Build parts
	parts := []map[string]interface{}{
		{"text": prompt},
		{
			"inlineData": map[string]string{
				"data":     base64Image,
				"mimeType": mimeType,
			},
		},
	}

	if textureBase64 != "" {
		parts = append(parts, map[string]interface{}{
			"inlineData": map[string]string{
				"data":     textureBase64,
				"mimeType": textureMime,
			},
		})
	}

	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": parts},
		},
	}

	jsonData, _ := json.Marshal(body)

	client := &http.Client{Timeout: 0}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result GeminiImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Candidates) == 0 {
		return "", fmt.Errorf("no candidates returned")
	}

	for _, part := range result.Candidates[0].Content.Parts {
		if part.InlineData != nil {
			return fmt.Sprintf(
				"data:%s;base64,%s",
				part.InlineData.MimeType,
				part.InlineData.Data,
			), nil
		}
	}

	return "", fmt.Errorf("no image generated")
}