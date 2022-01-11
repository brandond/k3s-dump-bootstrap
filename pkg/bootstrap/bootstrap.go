package bootstrap

import "time"

type File struct {
	Timestamp time.Time
	Content   []byte
}

type PathsDataformat map[string]File
