package p_lacerate

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"path"
	"strings"

	epub "github.com/mathieu-keller/epub-parser/v2"
	"github.com/mathieu-keller/epub-parser/v2/model"
	"github.com/PuerkitoBio/goquery"
)

const directMediaEpubMaxSpineItems = 48

// directMediaOPFSpine mirrors OPF package spine/manifest (local-name match; same pattern as epub-parser).
type directMediaOPFSpine struct {
	Manifest struct {
		Items []struct {
			ID        string `xml:"id,attr"`
			Href      string `xml:"href,attr"`
			MediaType string `xml:"media-type,attr"`
		} `xml:"item"`
	} `xml:"manifest"`
	Spine struct {
		Itemrefs []struct {
			IDRef  string `xml:"idref,attr"`
			Linear string `xml:"linear,attr,omitempty"`
		} `xml:"itemref"`
	} `xml:"spine"`
}

func directMediaEpubMetadataMarkdown(book *model.Book) string {
	if book == nil {
		return ""
	}
	meta := book.Metadata
	var b strings.Builder
	b.WriteString("## EPUB metadata\n\n")
	if meta.MainId.Id != "" || meta.MainId.Scheme != "" {
		fmt.Fprintf(&b, "- **Main identifier:** %s (%s)\n", strings.TrimSpace(meta.MainId.Id), strings.TrimSpace(meta.MainId.Scheme))
	}
	if meta.Titles != nil {
		for _, t := range *meta.Titles {
			ti := strings.TrimSpace(t.Title)
			if ti == "" {
				continue
			}
			fmt.Fprintf(&b, "- **Title:** %s\n", ti)
		}
	}
	if meta.Languages != nil {
		for _, lang := range *meta.Languages {
			lang = strings.TrimSpace(lang)
			if lang == "" {
				continue
			}
			fmt.Fprintf(&b, "- **Language:** %s\n", lang)
		}
	}
	if meta.Creators != nil {
		for _, c := range *meta.Creators {
			n := strings.TrimSpace(c.Name)
			if n == "" {
				continue
			}
			role := strings.TrimSpace(c.Role)
			if role != "" && role != "unknown" {
				fmt.Fprintf(&b, "- **Creator (%s):** %s\n", role, n)
			} else {
				fmt.Fprintf(&b, "- **Creator:** %s\n", n)
			}
		}
	}
	if meta.Publishers != nil {
		for _, p := range *meta.Publishers {
			t := strings.TrimSpace(p.Text)
			if t == "" {
				continue
			}
			fmt.Fprintf(&b, "- **Publisher:** %s\n", t)
		}
	}
	if meta.Subjects != nil {
		for _, s := range *meta.Subjects {
			t := strings.TrimSpace(s.Text)
			if t == "" {
				continue
			}
			fmt.Fprintf(&b, "- **Subject:** %s\n", t)
		}
	}
	if meta.Descriptions != nil {
		for _, d := range *meta.Descriptions {
			t := strings.TrimSpace(d.Text)
			if t == "" {
				continue
			}
			t = intelSanitizePostgresText(t)
			if len(t) > 4000 {
				t = t[:4000] + "…"
			}
			fmt.Fprintf(&b, "- **Description:** %s\n", t)
		}
	}
	if meta.Dates != nil {
		for _, d := range *meta.Dates {
			d = strings.TrimSpace(d)
			if d == "" {
				continue
			}
			fmt.Fprintf(&b, "- **Date:** %s\n", d)
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "## EPUB metadata" {
		return ""
	}
	return out
}

func directMediaEpubReadSpine(book *model.Book) (manifest map[string]struct {
	href      string
	mediaType string
}, spine []string, err error) {
	if book == nil {
		return nil, nil, fmt.Errorf("nil book")
	}
	var opf directMediaOPFSpine
	if err := book.ReadXML(book.Container.Rootfile.Path, &opf); err != nil {
		return nil, nil, err
	}
	manifest = make(map[string]struct {
		href      string
		mediaType string
	})
	for _, it := range opf.Manifest.Items {
		id := strings.TrimSpace(it.ID)
		if id == "" {
			continue
		}
		manifest[id] = struct {
			href      string
			mediaType string
		}{href: strings.TrimSpace(it.Href), mediaType: strings.TrimSpace(it.MediaType)}
	}
	for _, ref := range opf.Spine.Itemrefs {
		if strings.EqualFold(strings.TrimSpace(ref.Linear), "no") {
			continue
		}
		id := strings.TrimSpace(ref.IDRef)
		if id == "" {
			continue
		}
		spine = append(spine, id)
	}
	if len(manifest) == 0 || len(spine) == 0 {
		return manifest, spine, fmt.Errorf("empty manifest or spine")
	}
	return manifest, spine, nil
}

func directMediaEpubChapterToMarkdown(book *model.Book, href string, maxBytes int64) (string, string) {
	href = strings.TrimSpace(href)
	if href == "" {
		return "", ""
	}
	rc, err := book.Open(href)
	if err != nil {
		return "", fmt.Sprintf("open %s: %v", href, err)
	}
	defer rc.Close()
	raw, err := io.ReadAll(io.LimitReader(rc, maxBytes+1))
	if err != nil {
		return "", fmt.Sprintf("read %s: %v", href, err)
	}
	if int64(len(raw)) > maxBytes {
		return "", fmt.Sprintf("chapter %s exceeds size cap (%d)", href, maxBytes)
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Sprintf("parse html %s: %v", href, err)
	}
	body, err := doc.Find("body").Html()
	if err != nil {
		return "", fmt.Sprintf("body html %s: %v", href, err)
	}
	if strings.TrimSpace(body) == "" {
		html, _ := doc.Selection.Html()
		body = html
	}
	if linkExtractHTMLConv == nil {
		return "", "html converter not initialized"
	}
	md, err := linkExtractHTMLConv.ConvertString(body)
	if err != nil {
		return "", fmt.Sprintf("convert %s: %v", href, err)
	}
	md = strings.TrimSpace(md)
	return md, ""
}

func directMediaEpubSpineMarkdown(book *model.Book) (string, error) {
	if book == nil {
		return "", fmt.Errorf("nil book")
	}
	maxEntry := Config.DirectMedia.MaxArchiveEntryBytes
	if maxEntry <= 0 {
		maxEntry = 32 << 20
	}
	manifest, spine, err := directMediaEpubReadSpine(book)
	if err != nil {
		return "", err
	}
	var parts []string
	n := 0
	for _, id := range spine {
		if n >= directMediaEpubMaxSpineItems {
			parts = append(parts, fmt.Sprintf("\n*(spine truncated after %d items)*", directMediaEpubMaxSpineItems))
			break
		}
		it, ok := manifest[id]
		if !ok || strings.TrimSpace(it.href) == "" {
			continue
		}
		mt := strings.ToLower(it.mediaType)
		switch {
		case strings.Contains(mt, "html") || mt == "application/xhtml+xml":
			// ok
		case mt == "application/x-dtbncx+xml", strings.HasPrefix(mt, "image/"), strings.HasPrefix(mt, "font/"),
			strings.HasPrefix(mt, "audio/"), strings.HasPrefix(mt, "video/"):
			continue
		default:
			continue
		}
		md, note := directMediaEpubChapterToMarkdown(book, it.href, maxEntry)
		if note != "" {
			slog.Warn("lacerate: direct media epub chapter", "note", note, "href", it.href)
			continue
		}
		if md == "" {
			continue
		}
		parts = append(parts, "### "+path.Base(it.href)+"\n\n"+md)
		n++
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("no html spine items extracted")
	}
	return strings.Join(parts, "\n\n---\n\n"), nil
}

func directMediaEpubSections(asset directMediaAsset) []string {
	zr, err := zip.NewReader(bytes.NewReader(asset.Bytes), int64(len(asset.Bytes)))
	if err != nil {
		return []string{"## EPUB\n\n*(invalid zip/epub container: " + err.Error() + ")*"}
	}
	book, err := epub.OpenBook(zr)
	if err != nil {
		return []string{"## EPUB\n\n*(parse failed: " + err.Error() + ")*"}
	}
	var sections []string
	if meta := directMediaEpubMetadataMarkdown(book); meta != "" {
		sections = append(sections, meta)
	}
	body, err := directMediaEpubSpineMarkdown(book)
	if err != nil {
		sections = append(sections, "## Spine text\n\n*(could not extract: "+err.Error()+")*")
		return sections
	}
	body = strings.TrimSpace(body)
	if body == "" {
		sections = append(sections, "## Spine text\n\n*(empty)*")
		return sections
	}
	body = directMediaTruncateTextBodyForIntel(body)
	sections = append(sections, "## Spine text\n\n"+body)
	return sections
}
