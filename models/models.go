package models

type ApiConfig struct {
	AdminHTTPAddr  string
	PublicHttpAddr string
	APIKey         string
	APIKeyHeader   string
	FilesPath      string
	MaxFileSize    int
	Resolutions    []ResElement
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
