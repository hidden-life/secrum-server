package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	baseDirectory string
}

func NewLocalStorage(baseDirectory string) (*LocalStorage, error) {
	if baseDirectory == "" {
		return nil, fmt.Errorf("baseDirectory cannot be empty")
	}

	if err := os.MkdirAll(baseDirectory, 0755); err != nil {
		return nil, fmt.Errorf("error creating local directory: %w", err)
	}

	return &LocalStorage{baseDirectory: baseDirectory}, nil
}

func (l *LocalStorage) fullPath(path string) string {
	return filepath.Join(l.baseDirectory, filepath.Clean(path))
}

func (l *LocalStorage) Save(_ context.Context, path string, r io.Reader) error {
	fp := l.fullPath(path)
	if err := os.MkdirAll(filepath.Dir(fp), 0755); err != nil {
		return fmt.Errorf("error creating local directory: %w", err)
	}

	f, err := os.Create(fp)
	if err != nil {
		return fmt.Errorf("error creating local file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func (l *LocalStorage) Open(_ context.Context, path string) (io.ReadCloser, error) {
	fp := l.fullPath(path)
	f, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	return f, nil
}

func (l *LocalStorage) Delete(_ context.Context, path string) error {
	fp := l.fullPath(path)
	if err := os.Remove(fp); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error deleting file: %w", err)
	}

	return nil
}
