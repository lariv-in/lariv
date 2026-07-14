package p_llm_assistant

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

type skillExportJSON struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Files       []string `json:"files"`
}

func sanitizeFilename(s string) string {
	var res []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			res = append(res, r)
		} else {
			res = append(res, '_')
		}
	}
	out := string(res)
	if out == "" {
		return "skill"
	}
	return out
}

func handleSkillExport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid skill ID", http.StatusBadRequest)
		return
	}

	db, err := getters.DBFromContext(ctx)
	if err != nil {
		http.Error(w, "No database connection", http.StatusInternalServerError)
		return
	}

	var skill Skill
	if err := db.WithContext(ctx).Preload("Files").First(&skill, uint(id)).Error; err != nil {
		http.Error(w, "Skill not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", sanitizeFilename(skill.Name)))

	zw := zip.NewWriter(w)
	defer zw.Close()

	exportData := skillExportJSON{
		Name:        skill.Name,
		Description: skill.Description,
		Content:     skill.Content,
		Files:       make([]string, 0, len(skill.Files)),
	}

	for _, file := range skill.Files {
		exportData.Files = append(exportData.Files, file.Name)
	}

	indexBytes, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal index.json", http.StatusInternalServerError)
		return
	}

	indexFile, err := zw.Create("index.json")
	if err != nil {
		http.Error(w, "Failed to create index.json in zip", http.StatusInternalServerError)
		return
	}
	if _, err := indexFile.Write(indexBytes); err != nil {
		http.Error(w, "Failed to write index.json to zip", http.StatusInternalServerError)
		return
	}

	for _, file := range skill.Files {
		dl, err := file.OpenDownload()
		if err != nil {
			continue
		}

		fWriter, err := zw.Create(file.Name)
		if err != nil {
			dl.Reader.Close()
			continue
		}

		_, _ = io.Copy(fWriter, dl.Reader)
		dl.Reader.Close()
	}
}

func handleSkillImportRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		handleSkillImportPost(w, r)
		return
	}
	// Delegate GET to the dynamic view
	lago.NewDynamicView("llm_assistant.SkillsImportView").ServeHTTP(w, r)
}

func handleSkillImportPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := r.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	fileHeader := r.MultipartForm.File["File"]
	if len(fileHeader) == 0 {
		http.Error(w, "Zip file is required", http.StatusBadRequest)
		return
	}
	fh := fileHeader[0]

	f, err := fh.Open()
	if err != nil {
		http.Error(w, "Failed to open zip file", http.StatusBadRequest)
		return
	}
	defer f.Close()

	zipBytes, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, "Failed to read zip file", http.StatusInternalServerError)
		return
	}

	readerAt := bytes.NewReader(zipBytes)
	zipReader, err := zip.NewReader(readerAt, int64(len(zipBytes)))
	if err != nil {
		http.Error(w, "Failed to parse zip file", http.StatusBadRequest)
		return
	}

	var indexFile *zip.File
	for _, file := range zipReader.File {
		if file.Name == "index.json" {
			indexFile = file
			break
		}
	}
	if indexFile == nil {
		http.Error(w, "index.json is missing in the zip archive", http.StatusBadRequest)
		return
	}

	indexReader, err := indexFile.Open()
	if err != nil {
		http.Error(w, "Failed to open index.json", http.StatusInternalServerError)
		return
	}
	defer indexReader.Close()

	var exportData skillExportJSON
	if err := json.NewDecoder(indexReader).Decode(&exportData); err != nil {
		http.Error(w, "Failed to decode index.json", http.StatusBadRequest)
		return
	}

	if exportData.Name == "" {
		http.Error(w, "Skill name is required in index.json", http.StatusBadRequest)
		return
	}

	db, err := getters.DBFromContext(ctx)
	if err != nil {
		http.Error(w, "No database connection", http.StatusInternalServerError)
		return
	}

	var createdNodes []p_filesystem.VNode
	cleanup := func() {
		for _, node := range createdNodes {
			_ = p_filesystem.Store.Delete(node.FilePath)
		}
	}

	findFileInZip := func(name string) *zip.File {
		for _, file := range zipReader.File {
			if file.Name == name {
				return file
			}
		}
		return nil
	}

	for _, filename := range exportData.Files {
		zf := findFileInZip(filename)
		if zf == nil {
			cleanup()
			http.Error(w, fmt.Sprintf("File %q specified in index.json is missing from the zip archive", filename), http.StatusBadRequest)
			return
		}

		zfReader, err := zf.Open()
		if err != nil {
			cleanup()
			http.Error(w, fmt.Sprintf("Failed to open file %q in zip", filename), http.StatusInternalServerError)
			return
		}

		storedPath, err := p_filesystem.Store.SaveFromReader(zfReader, filepath.Ext(filename))
		zfReader.Close()
		if err != nil {
			cleanup()
			http.Error(w, fmt.Sprintf("Failed to save file %q to storage", filename), http.StatusInternalServerError)
			return
		}

		node := p_filesystem.VNode{
			Name:        filename,
			IsDirectory: false,
			FilePath:    storedPath,
		}
		if err := gorm.G[p_filesystem.VNode](db).Create(ctx, &node); err != nil {
			_ = p_filesystem.Store.Delete(storedPath)
			cleanup()
			http.Error(w, fmt.Sprintf("Failed to create file node %q in database", filename), http.StatusInternalServerError)
			return
		}
		createdNodes = append(createdNodes, node)
	}

	skill := Skill{
		Name:        exportData.Name,
		Description: exportData.Description,
		Content:     exportData.Content,
		Files:       createdNodes,
	}

	if err := gorm.G[Skill](db).Create(ctx, &skill); err != nil {
		cleanup()
		for _, node := range createdNodes {
			_, _ = gorm.G[p_filesystem.VNode](db).Where("id = ?", node.ID).Delete(ctx)
		}
		http.Error(w, fmt.Sprintf("Failed to save Skill to database: %v", err), http.StatusInternalServerError)
		return
	}

	redirectURLGetter := lago.RoutePath("llm_assistant.SkillsDetailRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static(skill.ID)),
	})
	redirectURL, err := redirectURLGetter(ctx)
	if err != nil {
		http.Error(w, "Failed to resolve redirect URL", http.StatusInternalServerError)
		return
	}

	views.HtmxRedirect(w, r, redirectURL, http.StatusSeeOther)
}
