package models

import "time"

type BaseEvent struct {
	EventID   string      `json:"event_id"`
	EventType string      `json:"event_type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

type ImageReceivedEventData struct {
	ImageID   string                 `json:"image_id"`
	ImageData string                 `json:"image_data"` // base64 encoded
	FileName  string                 `json:"file_name"`
	FileSize  int64                  `json:"file_size"`
	MimeType  string                 `json:"mime_type"`
	UserID    string                 `json:"user_id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type FaceRecognitionEventData struct {
	ImageID      string                  `json:"image_id"`
	FacesFound   int                     `json:"faces_found"`
	ProcessingMs int64                   `json:"processing_ms"`
	Results      []FaceRecognitionResult `json:"results"`
}

type FaceRecognitionResult struct {
	FaceID      string                 `json:"face_id"`
	Confidence  float64                `json:"confidence"`
	BoundingBox BoundingBox            `json:"bounding_box"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

type BoundingBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type DataSavedEventData struct {
	ImageID    string    `json:"image_id"`
	SavedAt    time.Time `json:"saved_at"`
	StorageURL string    `json:"storage_url,omitempty"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
}
