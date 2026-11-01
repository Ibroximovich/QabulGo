package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrFileTooLarge       = errors.New("fayl hajmi 5MB dan oshmasligi kerak")
	ErrInvalidFileType    = errors.New("faqat rasm fayllari (JPEG, PNG, WebP) qabul qilinadi")
	ErrEmptyFile          = errors.New("yuklanayotgan fayl bo‘sh bo‘lishi mumkin emas")
	ErrInvalidSubFolder   = errors.New("noto‘g‘ri papka nomi")
)

const MaxUploadSize = 5 * 1024 * 1024 // 5 MB

type StorageService interface {
	SaveUploadedFile(fileHeader *multipart.FileHeader, subFolder string) (string, error)
	DeleteFile(fileURL string) error
}

type localStorageService struct {
	baseDir string // e.g. "uploads"
	baseURL string // e.g. "/uploads"
}

func NewLocalStorageService(baseDir, baseURL string) StorageService {
	if baseDir == "" {
		baseDir = "uploads"
	}
	if baseURL == "" {
		baseURL = "/uploads"
	}
	_ = os.MkdirAll(baseDir, 0755)
	return &localStorageService{
		baseDir: baseDir,
		baseURL: baseURL,
	}
}

func (s *localStorageService) SaveUploadedFile(fileHeader *multipart.FileHeader, subFolder string) (string, error) {
	if fileHeader == nil || fileHeader.Size == 0 {
		return "", ErrEmptyFile
	}

	if fileHeader.Size > MaxUploadSize {
		return "", ErrFileTooLarge
	}

	// Xavfsiz subfolder tekshiruvi (path traversal oldini olish)
	cleanFolder := filepath.Clean(subFolder)
	if strings.Contains(cleanFolder, "..") || strings.HasPrefix(cleanFolder, "/") || strings.HasPrefix(cleanFolder, "\\") {
		return "", ErrInvalidSubFolder
	}

	// Faylni ochish
	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("faylni ochishda xatolik: %w", err)
	}
	defer src.Close()

	// Sniff MIME type from the first 512 bytes
	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("faylni o‘qishda xatolik: %w", err)
	}

	contentType := http.DetectContentType(buffer[:n])
	var ext string
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		return "", ErrInvalidFileType
	}

	// Fayl ko'rsatkichini boshiga qaytarish
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("fayl o‘rnini qaytarishda xatolik: %w", err)
	}

	// Maqsadli katalog
	targetDir := filepath.Join(s.baseDir, cleanFolder)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("katalog yaratishda xatolik: %w", err)
	}

	// Xavfsiz unikal fayl nomi
	safeFilename := uuid.New().String() + ext
	targetFilePath := filepath.Join(targetDir, safeFilename)

	dst, err := os.Create(targetFilePath)
	if err != nil {
		return "", fmt.Errorf("fayl yaratishda xatolik: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("faylni saqlashda xatolik: %w", err)
	}

	// Return URL (masalan "/uploads/clinics/uuid.jpg")
	return fmt.Sprintf("%s/%s/%s", s.baseURL, cleanFolder, safeFilename), nil
}

func (s *localStorageService) DeleteFile(fileURL string) error {
	if !strings.HasPrefix(fileURL, s.baseURL) {
		return nil
	}
	relPath := strings.TrimPrefix(fileURL, s.baseURL)
	cleanRel := filepath.Clean(relPath)
	if strings.Contains(cleanRel, "..") {
		return errors.New("xavfsiz bo‘lmagan fayl yo‘li")
	}

	fullPath := filepath.Join(s.baseDir, cleanRel)
	return os.Remove(fullPath)
}

