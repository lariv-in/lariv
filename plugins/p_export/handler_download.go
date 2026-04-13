package p_export

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func exportDB(r *http.Request, op string) *gorm.DB {
	db, err := getters.DBFromContext(r.Context())
	if err != nil {
		slog.Error("export: db from context", "operation", op, "error", err)
		return nil
	}
	return db
}

func downloadHandler(_ *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := exportDB(r, "download")
		if db == nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if err := r.ParseForm(); err != nil {
			slog.Error("export: parse form", "error", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		catalog, err := BuildExportCatalog(db)
		if err != nil {
			slog.Error("export: build catalog in download", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		selection, err := ExpandSelection(catalog, r.Form["models"])
		if err != nil {
			slog.Warn("export: invalid selection", "error", err, "models", r.Form["models"])
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		workbook, err := BuildWorkbook(db, catalog, selection)
		if err != nil {
			slog.Error("export: build workbook", "error", err, "models", selection.Tables)
			http.Error(w, "Failed to build workbook", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := workbook.Close(); err != nil {
				slog.Error("export: close workbook", "error", err)
			}
		}()

		var buffer bytes.Buffer
		if err := workbook.Write(&buffer); err != nil {
			slog.Error("export: write workbook", "error", err)
			http.Error(w, "Failed to write workbook", http.StatusInternalServerError)
			return
		}

		filename := fmt.Sprintf("export_%s.xlsx", time.Now().UTC().Format("20060102_150405"))
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buffer.Len()))
		if _, err := w.Write(buffer.Bytes()); err != nil {
			slog.Error("export: write response", "error", err)
		}
	})
}
