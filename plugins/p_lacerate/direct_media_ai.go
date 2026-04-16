package p_lacerate

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/genai"
)

const (
	directMediaAIFilePollInterval = 2 * time.Second
	directMediaAIFileWaitTimeout  = 2 * time.Minute
)

func directMediaGenAIClient(ctx context.Context) (*genai.Client, string, error) {
	key := strings.TrimSpace(Config.GeminiEmbedding.APIKey)
	if key == "" {
		return nil, "", nil
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  key,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		slog.Error("lacerate: direct media genai client", "error", err)
		return nil, "", err
	}
	model := strings.TrimSpace(Config.DirectMedia.AIModel)
	if model == "" {
		model = defaultDirectMediaAIModel
	}
	return client, model, nil
}

func directMediaAnalyzeImage(ctx context.Context, data []byte, mimeType, label string) (string, error) {
	client, model, err := directMediaGenAIClient(ctx)
	if err != nil || client == nil || len(data) == 0 {
		return "", err
	}
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}
	resp, err := client.Models.GenerateContent(ctx, model, []*genai.Content{
		genai.NewContentFromParts([]*genai.Part{
			genai.NewPartFromText(fmt.Sprintf("Analyze this image for OSINT ingest. Return concise markdown with sections: Summary, Visible text, Notable entities, Notes. File label: %s", strings.TrimSpace(label))),
			genai.NewPartFromBytes(data, mimeType),
		}, genai.RoleUser),
	}, nil)
	if err != nil {
		slog.Error("lacerate: direct media image analysis", "error", err, "label", label)
		return "", err
	}
	return strings.TrimSpace(resp.Text()), nil
}

func directMediaAnalyzeUploadedFile(ctx context.Context, data []byte, mimeType, label, prompt string) (string, error) {
	client, model, err := directMediaGenAIClient(ctx)
	if err != nil || client == nil || len(data) == 0 {
		return "", err
	}

	ext := filepath.Ext(strings.TrimSpace(label))
	tmp, err := os.CreateTemp("", "lacerate-direct-media-*"+ext)
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	file, err := client.Files.UploadFromPath(ctx, tmpPath, &genai.UploadFileConfig{
		DisplayName: sanitizePreviewFileName(label),
		MIMEType:    mimeType,
	})
	if err != nil {
		slog.Error("lacerate: direct media file upload", "error", err, "label", label, "mime", mimeType)
		return "", err
	}
	defer func() {
		if file != nil && strings.TrimSpace(file.Name) != "" {
			if _, err := client.Files.Delete(context.Background(), file.Name, nil); err != nil {
				slog.Warn("lacerate: direct media file delete", "error", err, "file", file.Name)
			}
		}
	}()

	file, err = directMediaWaitForUploadedFile(ctx, client, file)
	if err != nil {
		return "", err
	}
	resp, err := client.Models.GenerateContent(ctx, model, []*genai.Content{
		genai.NewContentFromParts([]*genai.Part{
			genai.NewPartFromText(prompt),
			genai.NewPartFromFile(*file),
		}, genai.RoleUser),
	}, nil)
	if err != nil {
		slog.Error("lacerate: direct media file analysis", "error", err, "label", label, "mime", mimeType)
		return "", err
	}
	return strings.TrimSpace(resp.Text()), nil
}

func directMediaWaitForUploadedFile(ctx context.Context, client *genai.Client, file *genai.File) (*genai.File, error) {
	if client == nil || file == nil {
		return nil, fmt.Errorf("direct media uploaded file is nil")
	}
	waitCtx, cancel := context.WithTimeout(ctx, directMediaAIFileWaitTimeout)
	defer cancel()
	current := file
	for {
		switch current.State {
		case genai.FileStateActive, genai.FileStateUnspecified:
			return current, nil
		case genai.FileStateFailed:
			return nil, fmt.Errorf("direct media file processing failed for %q", current.Name)
		}
		select {
		case <-waitCtx.Done():
			return nil, waitCtx.Err()
		case <-time.After(directMediaAIFilePollInterval):
		}
		next, err := client.Files.Get(waitCtx, current.Name, nil)
		if err != nil {
			slog.Error("lacerate: direct media file get", "error", err, "file", current.Name)
			return nil, err
		}
		current = next
	}
}
