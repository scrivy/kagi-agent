package models

// POST /api/upload

type PostUploadResponse struct {
	UUID             string `json:"uuid"`
	Filename         string `json:"filename"`
	OriginalFilename string `json:"original_filename"`
	MimeType         string `json:"mime_type"`
	Size             int    `json:"size"`
	URL              string `json:"url"`
}
