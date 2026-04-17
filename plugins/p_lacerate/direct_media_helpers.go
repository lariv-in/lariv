package p_lacerate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type directMediaKind string

const (
	directMediaKindOther   directMediaKind = "other"
	directMediaKindPDF     directMediaKind = "pdf"
	directMediaKindImage   directMediaKind = "image"
	directMediaKindVideo   directMediaKind = "video"
	directMediaKindArchive directMediaKind = "archive"
	directMediaKindText    directMediaKind = "text"
	directMediaKindEPUB    directMediaKind = "epub"
)

type directMediaArchiveFormat string

const (
	directMediaArchiveFormatUnknown directMediaArchiveFormat = ""
	directMediaArchiveFormatZIP     directMediaArchiveFormat = "zip"
	directMediaArchiveFormatTAR     directMediaArchiveFormat = "tar"
	directMediaArchiveFormatTARGZ   directMediaArchiveFormat = "tar.gz"
	directMediaArchiveFormatGZIP    directMediaArchiveFormat = "gz"
)

// directMediaPDFMarkdownPrompt asks Gemini for clean markdown only (no low-level PDF text dump).
const directMediaPDFMarkdownPrompt = `Convert this PDF into archival markdown for OSINT records.

Rules:
- Output only readable UTF-8 markdown: headings (# / ##), bullet lists, short paragraphs. No binary data, hex dumps, or meaningless control characters.
- Do not transcribe garbled glyph runs or layout noise; summarize legible content instead.
- If the document is scanned, image-only, or mostly unreadable, say so briefly, then extract whatever useful text exists.

Use these section headings (## exactly as written):
## Summary
## Key entities
## Dates and places
## Contact info
## Notes
`

type directMediaAsset struct {
	SourceURL   string
	Path        string
	DisplayName string
	MIMEType    string
	SizeBytes   int64
	Bytes       []byte
	Kind        directMediaKind
	Note        string
}

type directMediaArchiveState struct {
	EntryCount    int
	ExpandedBytes int64
}

func (a directMediaAsset) dedupeSeed() string {
	if strings.TrimSpace(a.Path) == "" {
		return strings.TrimSpace(a.SourceURL)
	}
	return strings.TrimSpace(a.SourceURL) + "#" + strings.TrimSpace(a.Path)
}

func (a directMediaAsset) label() string {
	if s := strings.TrimSpace(a.Path); s != "" {
		return s
	}
	if s := strings.TrimSpace(a.DisplayName); s != "" {
		return s
	}
	return strings.TrimSpace(a.SourceURL)
}

func directMediaDedupHash(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	sum := sha256.Sum256([]byte("direct_media:" + raw))
	return hex.EncodeToString(sum[:])
}

func directMediaKindLabel(kind directMediaKind) string {
	switch kind {
	case directMediaKindPDF:
		return "PDF"
	case directMediaKindImage:
		return "Image"
	case directMediaKindVideo:
		return "Video"
	case directMediaKindArchive:
		return "Archive"
	case directMediaKindText:
		return "Text"
	case directMediaKindEPUB:
		return "EPUB"
	default:
		return "File"
	}
}

func directMediaKindFromNameAndMIME(name, mimeType string) directMediaKind {
	name = strings.ToLower(strings.TrimSpace(name))
	mimeType = strings.ToLower(strings.TrimSpace(strings.Split(mimeType, ";")[0]))
	switch {
	case mimeType == "text/html" || strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".htm"):
		return directMediaKindOther
	case (strings.HasPrefix(mimeType, "text/") && mimeType != "text/html" && mimeType != "text/javascript" && mimeType != "text/css") ||
		mimeType == "application/json" ||
		mimeType == "text/yaml" || mimeType == "application/x-yaml" ||
		mimeType == "application/xml" || mimeType == "text/xml" ||
		directMediaFilenameLooksLikePlainText(name):
		return directMediaKindText
	case mimeType == "application/pdf" || strings.HasSuffix(name, ".pdf"):
		return directMediaKindPDF
	case mimeType == "application/epub+zip" || strings.HasSuffix(name, ".epub"):
		return directMediaKindEPUB
	case strings.HasPrefix(mimeType, "image/"),
		strings.HasSuffix(name, ".jpg"),
		strings.HasSuffix(name, ".jpeg"),
		strings.HasSuffix(name, ".png"),
		strings.HasSuffix(name, ".gif"),
		strings.HasSuffix(name, ".webp"),
		strings.HasSuffix(name, ".bmp"),
		strings.HasSuffix(name, ".svg"),
		strings.HasSuffix(name, ".tif"),
		strings.HasSuffix(name, ".tiff"):
		return directMediaKindImage
	case strings.HasPrefix(mimeType, "video/"),
		strings.HasSuffix(name, ".mp4"),
		strings.HasSuffix(name, ".webm"),
		strings.HasSuffix(name, ".mov"),
		strings.HasSuffix(name, ".m4v"),
		strings.HasSuffix(name, ".avi"),
		strings.HasSuffix(name, ".mkv"):
		return directMediaKindVideo
	case directMediaArchiveFormatFromNameAndMIME(name, mimeType) != directMediaArchiveFormatUnknown,
		strings.HasSuffix(name, ".rar"),
		strings.HasSuffix(name, ".7z"),
		strings.HasSuffix(name, ".xz"),
		strings.HasSuffix(name, ".bz2"):
		return directMediaKindArchive
	default:
		return directMediaKindOther
	}
}

func directMediaFilenameLooksLikePlainText(name string) bool {
	switch strings.ToLower(strings.TrimSpace(filepath.Ext(name))) {
	case ".txt", ".md", ".markdown", ".log", ".csv", ".tsv", ".env", ".gitignore",
		".yml", ".yaml", ".json", ".xml", ".adoc", ".rst", ".sql",
		".sh", ".bash", ".zsh", ".ps1", ".bat", ".cmd", ".ini", ".cfg", ".toml", ".properties":
		return true
	default:
		return false
	}
}

func directMediaArchiveFormatFromNameAndMIME(name, mimeType string) directMediaArchiveFormat {
	name = strings.ToLower(strings.TrimSpace(name))
	mimeType = strings.ToLower(strings.TrimSpace(strings.Split(mimeType, ";")[0]))
	switch {
	case strings.HasSuffix(name, ".zip"),
		mimeType == "application/zip", mimeType == "application/x-zip-compressed":
		return directMediaArchiveFormatZIP
	case strings.HasSuffix(name, ".tar.gz"), strings.HasSuffix(name, ".tgz"):
		return directMediaArchiveFormatTARGZ
	case strings.HasSuffix(name, ".tar"),
		mimeType == "application/x-tar":
		return directMediaArchiveFormatTAR
	case strings.HasSuffix(name, ".gz"),
		mimeType == "application/gzip", mimeType == "application/x-gzip":
		return directMediaArchiveFormatGZIP
	default:
		return directMediaArchiveFormatUnknown
	}
}

func directMediaIsFetchableURL(ctx context.Context, raw string) (*url.URL, string, error) {
	normalized, err := normalizeWebsiteSeedURL(raw)
	if err != nil {
		return nil, "", err
	}
	parsed, err := url.Parse(normalized)
	if err != nil {
		return nil, "", err
	}
	if linkedURLFailsSSRF(ctx, parsed) {
		return nil, "", fmt.Errorf("url blocked by ssrf guard: %s", normalized)
	}
	return parsed, normalized, nil
}

func directMediaInferMIMEType(name, declared string, data []byte) string {
	declared = strings.TrimSpace(strings.ToLower(strings.Split(declared, ";")[0]))
	if extType := mime.TypeByExtension(strings.ToLower(filepath.Ext(name))); extType != "" && (declared == "" || declared == "application/octet-stream") {
		declared = extType
	}
	if len(data) > 0 {
		detected := strings.TrimSpace(strings.ToLower(strings.Split(http.DetectContentType(data), ";")[0]))
		if declared == "" || declared == "application/octet-stream" {
			declared = detected
		}
	}
	return declared
}

func directMediaFetchRoot(ctx context.Context, db *gorm.DB, raw string) (directMediaAsset, error) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, directMediaFSURLPrefix) {
		if db == nil {
			return directMediaAsset{}, fmt.Errorf("direct media filesystem url requires database access")
		}
		return directMediaFetchUploadedVNode(ctx, db, raw)
	}
	parsed, normalized, err := directMediaIsFetchableURL(ctx, raw)
	if err != nil {
		return directMediaAsset{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normalized, nil)
	if err != nil {
		return directMediaAsset{}, err
	}
	req.Header.Set("User-Agent", Config.IntelPreview.UserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	client := &http.Client{Timeout: linkedArticleTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return directMediaAsset{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return directMediaAsset{}, fmt.Errorf("direct media download status %s", resp.Status)
	}

	limit := Config.DirectMedia.MaxDownloadBytes
	if limit <= 0 {
		limit = defaultDirectMediaMaxDownloadBytes
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return directMediaAsset{}, err
	}
	if int64(len(data)) > limit {
		return directMediaAsset{}, fmt.Errorf("direct media file exceeds maxDownloadBytes (%d)", limit)
	}

	name := path.Base(parsed.Path)
	if name == "." || name == "/" || name == "" {
		name = parsed.Hostname()
	}
	mimeType := directMediaInferMIMEType(name, resp.Header.Get("Content-Type"), data)
	if strings.HasPrefix(mimeType, "text/html") {
		return directMediaAsset{}, fmt.Errorf("url resolved to html instead of direct media")
	}
	return directMediaAsset{
		SourceURL:   normalized,
		DisplayName: sanitizePreviewFileName(name),
		MIMEType:    mimeType,
		SizeBytes:   int64(len(data)),
		Bytes:       data,
	}, nil
}

func directMediaArchiveRegisterEntry(state *directMediaArchiveState, size int64) error {
	if state == nil {
		return nil
	}
	state.EntryCount++
	if state.EntryCount > Config.DirectMedia.MaxArchiveEntries {
		return fmt.Errorf("archive entry count exceeds limit (%d)", Config.DirectMedia.MaxArchiveEntries)
	}
	if size < 0 {
		size = 0
	}
	if size <= Config.DirectMedia.MaxArchiveEntryBytes {
		state.ExpandedBytes += size
		if state.ExpandedBytes > Config.DirectMedia.MaxArchiveExpandedBytes {
			return fmt.Errorf("archive expanded bytes exceeds limit (%d)", Config.DirectMedia.MaxArchiveExpandedBytes)
		}
	}
	return nil
}

func directMediaJoinPath(parentPath, child string) string {
	child = strings.TrimPrefix(strings.ReplaceAll(strings.TrimSpace(child), "\\", "/"), "/")
	if child == "" {
		return strings.TrimSpace(parentPath)
	}
	if strings.TrimSpace(parentPath) == "" {
		return path.Clean(child)
	}
	return path.Clean(path.Join(parentPath, child))
}

func directMediaReadLimited(rc io.ReadCloser, maxBytes int64) ([]byte, string) {
	defer rc.Close()
	data, err := io.ReadAll(io.LimitReader(rc, maxBytes+1))
	if err != nil {
		return nil, err.Error()
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Sprintf("entry exceeds maxArchiveEntryBytes (%d)", maxBytes)
	}
	return data, ""
}

func directMediaArchiveAssets(asset directMediaAsset, state *directMediaArchiveState) ([]directMediaAsset, error) {
	switch directMediaArchiveFormatFromNameAndMIME(asset.label(), asset.MIMEType) {
	case directMediaArchiveFormatZIP:
		return directMediaArchiveAssetsZIP(asset, state)
	case directMediaArchiveFormatTAR:
		return directMediaArchiveAssetsTAR(asset, state)
	case directMediaArchiveFormatTARGZ:
		return directMediaArchiveAssetsTARGZ(asset, state)
	case directMediaArchiveFormatGZIP:
		return directMediaArchiveAssetsGZIP(asset, state)
	default:
		return nil, nil
	}
}

func directMediaArchiveAssetsZIP(asset directMediaAsset, state *directMediaArchiveState) ([]directMediaAsset, error) {
	zr, err := zip.NewReader(bytes.NewReader(asset.Bytes), int64(len(asset.Bytes)))
	if err != nil {
		return nil, err
	}
	var out []directMediaAsset
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if err := directMediaArchiveRegisterEntry(state, int64(f.UncompressedSize64)); err != nil {
			return nil, err
		}
		childPath := directMediaJoinPath(asset.Path, f.Name)
		child := directMediaAsset{
			SourceURL:   asset.SourceURL,
			Path:        childPath,
			DisplayName: sanitizePreviewFileName(path.Base(childPath)),
			MIMEType:    directMediaInferMIMEType(childPath, "", nil),
			SizeBytes:   int64(f.UncompressedSize64),
		}
		rc, err := f.Open()
		if err != nil {
			child.Note = fmt.Sprintf("archive entry open failed: %v", err)
			out = append(out, child)
			continue
		}
		data, note := directMediaReadLimited(rc, Config.DirectMedia.MaxArchiveEntryBytes)
		child.Note = note
		child.Bytes = data
		child.MIMEType = directMediaInferMIMEType(childPath, child.MIMEType, data)
		out = append(out, child)
	}
	return out, nil
}

func directMediaArchiveAssetsTAR(asset directMediaAsset, state *directMediaArchiveState) ([]directMediaAsset, error) {
	tr := tar.NewReader(bytes.NewReader(asset.Bytes))
	return directMediaArchiveAssetsFromTarReader(asset, state, tr)
}

func directMediaArchiveAssetsTARGZ(asset directMediaAsset, state *directMediaArchiveState) ([]directMediaAsset, error) {
	gr, err := gzip.NewReader(bytes.NewReader(asset.Bytes))
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	return directMediaArchiveAssetsFromTarReader(asset, state, tar.NewReader(gr))
}

func directMediaArchiveAssetsFromTarReader(asset directMediaAsset, state *directMediaArchiveState, tr *tar.Reader) ([]directMediaAsset, error) {
	var out []directMediaAsset
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
		if hdr == nil || hdr.FileInfo().IsDir() {
			continue
		}
		if err := directMediaArchiveRegisterEntry(state, hdr.Size); err != nil {
			return nil, err
		}
		childPath := directMediaJoinPath(asset.Path, hdr.Name)
		child := directMediaAsset{
			SourceURL:   asset.SourceURL,
			Path:        childPath,
			DisplayName: sanitizePreviewFileName(path.Base(childPath)),
			MIMEType:    directMediaInferMIMEType(childPath, "", nil),
			SizeBytes:   hdr.Size,
		}
		data, err := io.ReadAll(io.LimitReader(tr, Config.DirectMedia.MaxArchiveEntryBytes+1))
		if err != nil {
			child.Note = fmt.Sprintf("archive entry read failed: %v", err)
			out = append(out, child)
			continue
		}
		if int64(len(data)) > Config.DirectMedia.MaxArchiveEntryBytes {
			child.Note = fmt.Sprintf("entry exceeds maxArchiveEntryBytes (%d)", Config.DirectMedia.MaxArchiveEntryBytes)
			out = append(out, child)
			continue
		}
		child.Bytes = data
		child.MIMEType = directMediaInferMIMEType(childPath, child.MIMEType, data)
		out = append(out, child)
	}
}

func directMediaArchiveAssetsGZIP(asset directMediaAsset, state *directMediaArchiveState) ([]directMediaAsset, error) {
	gr, err := gzip.NewReader(bytes.NewReader(asset.Bytes))
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	name := strings.TrimSuffix(asset.label(), ".gz")
	if name == asset.label() {
		name = asset.label() + ".ungz"
	}
	data, err := io.ReadAll(io.LimitReader(gr, Config.DirectMedia.MaxArchiveEntryBytes+1))
	if err != nil {
		return nil, err
	}
	size := int64(len(data))
	if err := directMediaArchiveRegisterEntry(state, size); err != nil {
		return nil, err
	}
	child := directMediaAsset{
		SourceURL:   asset.SourceURL,
		Path:        directMediaJoinPath(asset.Path, name),
		DisplayName: sanitizePreviewFileName(path.Base(name)),
		MIMEType:    directMediaInferMIMEType(name, "", data),
		SizeBytes:   size,
	}
	if size > Config.DirectMedia.MaxArchiveEntryBytes {
		child.Note = fmt.Sprintf("entry exceeds maxArchiveEntryBytes (%d)", Config.DirectMedia.MaxArchiveEntryBytes)
		return []directMediaAsset{child}, nil
	}
	child.Bytes = data
	return []directMediaAsset{child}, nil
}

func directMediaImageMetadata(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return ""
	}
	return fmt.Sprintf("- **Image format:** %s\n- **Dimensions:** %dx%d", format, cfg.Width, cfg.Height)
}

func directMediaTrimMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if markdownRuneLen(s) <= maxLinkedMarkdownRunes {
		return s
	}
	runes := []rune(s)
	return strings.TrimSpace(string(runes[:maxLinkedMarkdownRunes])) + "\n\n...(truncated)"
}

// directMediaTextBodyMaxRunes leaves room for metadata headers in the final intel markdown.
const directMediaTextBodyMaxRunes = maxLinkedMarkdownRunes - 8192

func directMediaTruncateTextBodyForIntel(s string) string {
	s = strings.TrimSpace(s)
	rs := []rune(s)
	if len(rs) <= directMediaTextBodyMaxRunes {
		return s
	}
	return strings.TrimSpace(string(rs[:directMediaTextBodyMaxRunes])) + "\n\n...(truncated)"
}

func directMediaIndentEachLineAsMarkdownCode(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		b.WriteString("    ")
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func directMediaTextIntelSections(data []byte) string {
	body := intelSanitizePostgresText(string(data))
	body = strings.TrimSpace(body)
	if body == "" {
		return "## Contents\n\n*(empty file)*"
	}
	body = directMediaTruncateTextBodyForIntel(body)
	return "## Contents\n\n" + directMediaIndentEachLineAsMarkdownCode(body)
}

func directMediaMarkdown(asset directMediaAsset, kind directMediaKind, sections ...string) string {
	var b strings.Builder
	title := asset.label()
	if title == "" {
		title = directMediaKindLabel(kind)
	}
	fmt.Fprintf(&b, "## %s\n\n", title)
	fmt.Fprintf(&b, "- **Kind:** %s\n", directMediaKindLabel(kind))
	if asset.SourceURL != "" {
		fmt.Fprintf(&b, "- **Source URL:** %s\n", asset.SourceURL)
	}
	if asset.Path != "" {
		fmt.Fprintf(&b, "- **Archive path:** %s\n", asset.Path)
	}
	if asset.MIMEType != "" {
		fmt.Fprintf(&b, "- **MIME type:** %s\n", asset.MIMEType)
	}
	if asset.SizeBytes > 0 {
		fmt.Fprintf(&b, "- **Size bytes:** %d\n", asset.SizeBytes)
	}
	if note := strings.TrimSpace(asset.Note); note != "" {
		fmt.Fprintf(&b, "- **Note:** %s\n", strings.ReplaceAll(note, "\n", " "))
	}
	for _, sec := range sections {
		sec = strings.TrimSpace(sec)
		if sec == "" {
			continue
		}
		b.WriteString("\n\n")
		b.WriteString(sec)
	}
	return directMediaTrimMarkdown(b.String())
}

func directMediaIntel(sourceID uint, asset directMediaAsset, content string, previewID *uint) (Intel, bool) {
	content = strings.TrimSpace(content)
	dedup := directMediaDedupHash(asset.dedupeSeed())
	if content == "" || dedup == "" {
		return Intel{}, false
	}
	sourceIDCopy := sourceID
	dedupCopy := dedup
	return Intel{
		SourceID:       &sourceIDCopy,
		DedupHash:      &dedupCopy,
		Content:        content,
		Datetime:       time.Now().UTC(),
		PreviewImageID: previewID,
	}, true
}

func directMediaMaybeAppendIntel(out *[]Intel, existingDedup map[string]struct{}, intel Intel, ok bool) {
	if !ok || intel.DedupHash == nil || *intel.DedupHash == "" {
		return
	}
	if _, dup := existingDedup[*intel.DedupHash]; dup {
		return
	}
	existingDedup[*intel.DedupHash] = struct{}{}
	*out = append(*out, intel)
}

func directMediaArchiveSummarySection(children []directMediaAsset) string {
	if len(children) == 0 {
		return "## Archive contents\n\nNo files extracted."
	}
	var b strings.Builder
	b.WriteString("## Archive contents\n\n")
	fmt.Fprintf(&b, "- **Files extracted:** %d\n", len(children))
	limit := len(children)
	if limit > 12 {
		limit = 12
	}
	for i := 0; i < limit; i++ {
		fmt.Fprintf(&b, "- `%s`\n", children[i].label())
	}
	if len(children) > limit {
		fmt.Fprintf(&b, "- ... and %d more\n", len(children)-limit)
	}
	return b.String()
}

func directMediaExtractAsset(ctx context.Context, db *gorm.DB, sourceID uint, existingDedup map[string]struct{}, archiveState *directMediaArchiveState, asset directMediaAsset, depth int) ([]Intel, error) {
	kind := asset.Kind
	if kind == "" {
		kind = directMediaKindFromNameAndMIME(asset.label(), asset.MIMEType)
	}
	switch kind {
	case directMediaKindPDF:
		// PDF body comes from Gemini file analysis only; local rsc.io/pdf text was dropped (NULs / garbage in Postgres, low signal).
		aiText, err := directMediaAnalyzeUploadedFile(ctx, asset.Bytes, asset.MIMEType, asset.label(), directMediaPDFMarkdownPrompt)
		if err != nil {
			slog.Warn("lacerate: direct media pdf ai analysis", "error", err, "label", asset.label())
		}
		var sections []string
		if strings.TrimSpace(aiText) != "" {
			sections = append(sections, strings.TrimSpace(aiText))
		} else {
			sections = append(sections, "## PDF ingest\n\nGemini returned no markdown (check API key and `directMedia` model). Raw PDF text extraction is not stored to avoid unusable payloads.")
		}
		intel, ok := directMediaIntel(sourceID, asset, directMediaMarkdown(asset, kind, sections...), nil)
		var out []Intel
		directMediaMaybeAppendIntel(&out, existingDedup, intel, ok)
		return out, nil
	case directMediaKindImage:
		var sections []string
		if meta := directMediaImageMetadata(asset.Bytes); meta != "" {
			sections = append(sections, "## Image metadata\n\n"+meta)
		}
		aiText, err := directMediaAnalyzeImage(ctx, asset.Bytes, asset.MIMEType, asset.label())
		if err != nil {
			slog.Warn("lacerate: direct media image ai analysis", "error", err, "label", asset.label())
		}
		if aiText != "" {
			sections = append(sections, "## AI analysis\n\n"+aiText)
		}
		previewID := persistIntelPreviewBytes(ctx, db, asset.label(), asset.Bytes, asset.MIMEType)
		intel, ok := directMediaIntel(sourceID, asset, directMediaMarkdown(asset, kind, sections...), previewID)
		var out []Intel
		directMediaMaybeAppendIntel(&out, existingDedup, intel, ok)
		return out, nil
	case directMediaKindVideo:
		var sections []string
		aiText, err := directMediaAnalyzeUploadedFile(ctx, asset.Bytes, asset.MIMEType, asset.label(), "Analyze this video for OSINT ingest. Return concise markdown with sections: Summary, Spoken or visible text, Notable entities, Dates and locations, Notes.")
		if err != nil {
			slog.Warn("lacerate: direct media video ai analysis", "error", err, "label", asset.label())
		}
		if aiText != "" {
			sections = append(sections, "## AI analysis\n\n"+aiText)
		}
		intel, ok := directMediaIntel(sourceID, asset, directMediaMarkdown(asset, kind, sections...), nil)
		var out []Intel
		directMediaMaybeAppendIntel(&out, existingDedup, intel, ok)
		return out, nil
	case directMediaKindArchive:
		summary := []string{}
		if directMediaArchiveFormatFromNameAndMIME(asset.label(), asset.MIMEType) == directMediaArchiveFormatUnknown {
			asset.Note = strings.TrimSpace(strings.TrimSpace(asset.Note) + "\nUnsupported archive format; metadata only.")
			intel, ok := directMediaIntel(sourceID, asset, directMediaMarkdown(asset, kind), nil)
			var out []Intel
			directMediaMaybeAppendIntel(&out, existingDedup, intel, ok)
			return out, nil
		}
		children, err := directMediaArchiveAssets(asset, archiveState)
		if err != nil {
			return nil, err
		}
		summary = append(summary, directMediaArchiveSummarySection(children))
		var out []Intel
		intel, ok := directMediaIntel(sourceID, asset, directMediaMarkdown(asset, kind, summary...), nil)
		directMediaMaybeAppendIntel(&out, existingDedup, intel, ok)
		if depth <= 0 {
			return out, nil
		}
		for _, child := range children {
			childIntels, err := directMediaExtractAsset(ctx, db, sourceID, existingDedup, archiveState, child, depth-1)
			if err != nil {
				slog.Error("lacerate: direct media archive child", "error", err, "label", child.label())
				continue
			}
			out = append(out, childIntels...)
		}
		return out, nil
	case directMediaKindText:
		section := directMediaTextIntelSections(asset.Bytes)
		intel, ok := directMediaIntel(sourceID, asset, directMediaMarkdown(asset, kind, section), nil)
		var out []Intel
		directMediaMaybeAppendIntel(&out, existingDedup, intel, ok)
		return out, nil
	case directMediaKindEPUB:
		sections := directMediaEpubSections(asset)
		intel, ok := directMediaIntel(sourceID, asset, directMediaMarkdown(asset, kind, sections...), nil)
		var out []Intel
		directMediaMaybeAppendIntel(&out, existingDedup, intel, ok)
		return out, nil
	default:
		intel, ok := directMediaIntel(sourceID, asset, directMediaMarkdown(asset, kind), nil)
		var out []Intel
		directMediaMaybeAppendIntel(&out, existingDedup, intel, ok)
		return out, nil
	}
}
