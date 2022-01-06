package models

type ApiConfig struct {
	AdminHTTPAddr  string `json:"adminHttpAddr"`
	PublicHttpAddr string `json:"publicHttpAddr"`
	APIKey         string `json:"apiKey"`
	APIKeyHeader   string `json:"apiKeyHeader"`
	FilesPath      string `json:"filesPath"`
	ConvertToRes   string `json:"convertToRes"`
	MaxFileSize    int    `json:"maxFileSize"`
}

type ResElement struct {
	Width  int
	Height int
}

type ConvertCommand struct {
	Path       string
	WebP       bool
	ConvertRes bool
	TargetRes  ResElement
}
