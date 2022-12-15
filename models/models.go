package models

type ApiConfig struct {
	APIKey       string
	APIKeyHeader string
	FilesPath    string
	MaxFileSize  int
	CacheTime    int
	Resolutions  []ResElement
}

type ResElement struct {
	Width  int
	Height int
}

type ConvertCommand struct {
	Path       string
	WebP       bool
	Heic       bool
	ConvertRes bool
	TargetRes  ResElement
}
