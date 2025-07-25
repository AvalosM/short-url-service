package shorturl

import "errors"

var (
	ErrShortURLNotFound = errors.New("short URL not found")
	ErrShortURLExists   = errors.New("short URL already exists")
	ErrInvalidLongURL   = errors.New("invalid long URL")
)
