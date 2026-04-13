package p_nirmancampus_programs

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"strconv"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

// ctxKeyStructureParentProgram marks the scoped program ID for structure routes (POST redirect, validation).
type ctxKeyStructureParentProgram struct{}

// structureUnitScopeForContextProgram restricts ProgramStructureUnit loads to the program in context.
type structureUnitScopeForContextProgram struct{}

func (structureUnitScopeForContextProgram) Patch(_ views.View, r *http.Request, q gorm.ChainInterface[ProgramStructureUnit]) gorm.ChainInterface[ProgramStructureUnit] {
	p, ok := r.Context().Value("program").(Program)
	if !ok || p.ID == 0 {
		return q.Where("1 = 0")
	}
	return q.Where("program_id = ?", p.ID)
}

// programsStructureLoadProgramLayer loads the program from {id} with role scope and preloads structure units.
type programsStructureLoadProgramLayer struct{}

func (programsStructureLoadProgramLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("structure load program: db from context", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		query := programScopeByRole{}.Patch(views.View{}, r, gorm.G[Program](db).Scopes())
		query = queryPatcherPreloadProgramStructureUnits{}.Patch(views.View{}, r, query)
		p, err := query.Where("id = ?", uint(id)).First(r.Context())
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), "program", p)
		ctx = context.WithValue(ctx, ctxKeyStructureParentProgram{}, p.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func structureParentProgramID(ctx context.Context) (uint, bool) {
	v, ok := ctx.Value(ctxKeyStructureParentProgram{}).(uint)
	return v, ok && v != 0
}

func structureEditRedirectURL(ctx context.Context) (string, error) {
	pid, ok := structureParentProgramID(ctx)
	if !ok {
		return "", fmt.Errorf("missing program id")
	}
	return lago.RoutePath("programs.StructureEditRoute", map[string]getters.Getter[any]{
		"id": getters.Any(getters.Static[uint](pid)),
	})(ctx)
}

// structureUnitProgramIDFromPathPatcher forces ProgramID from the program path param (do not trust the client).
type structureUnitProgramIDFromPathPatcher struct{}

func (structureUnitProgramIDFromPathPatcher) Patch(_ views.View, r *http.Request, m map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err == nil {
		m["ProgramID"] = uint(id)
	}
	return m, formErrors
}

func splitStructureUnitFormValues(values map[string]any) (map[string]any, map[string]components.AssociationIDs) {
	regular := make(map[string]any)
	assoc := make(map[string]components.AssociationIDs)
	for k, v := range values {
		switch typed := v.(type) {
		case components.AssociationIDs:
			if typed.Field == "" {
				typed.Field = k
			}
			assoc[k] = typed
		case *components.AssociationIDs:
			if typed == nil {
				continue
			}
			if typed.Field == "" {
				typed.Field = k
			}
			assoc[k] = *typed
		default:
			regular[k] = v
		}
	}
	return regular, assoc
}

func replaceStructureUnitCourseAssociations(db *gorm.DB, unit *ProgramStructureUnit, assoc map[string]components.AssociationIDs) error {
	for _, key := range []string{"CompulsoryCourses", "OptionalCourseSelectionPool"} {
		a, ok := assoc[key]
		if !ok {
			continue
		}
		field := a.Field
		if field == "" {
			field = key
		}
		if err := replaceCourseAssociationByIDs(db, unit, field, a.IDs); err != nil {
			return err
		}
	}
	return nil
}

func replaceCourseAssociationByIDs(db *gorm.DB, unit *ProgramStructureUnit, field string, ids []uint) error {
	association := db.Model(unit).Association(field)
	if len(ids) == 0 {
		return association.Clear()
	}
	rows := make([]courses.Course, len(ids))
	for i, id := range ids {
		rows[i] = courses.Course{Model: gorm.Model{ID: id}}
	}
	return association.Replace(rows)
}

func handleStructureUnitCreate(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			return
		}
		var progPatch structureUnitProgramIDFromPathPatcher
		values, fieldErrors = progPatch.Patch(*v, r, values, fieldErrors)
		if len(fieldErrors) != 0 {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		parentID, ok := structureParentProgramID(r.Context())
		if !ok {
			http.Error(w, "missing program", http.StatusBadRequest)
			return
		}
		regular, assoc := splitStructureUnitFormValues(values)
		vals := maps.Clone(regular)

		record := new(ProgramStructureUnit)
		if err := views.PopulateFromMap(record, vals); err != nil {
			fieldErrors["_form"] = err
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		if record.ProgramID != parentID {
			fieldErrors["ProgramID"] = fmt.Errorf("invalid program")
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("structure unit create: db from context", "error", dberr)
			fieldErrors["_form"] = dberr
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := gorm.G[ProgramStructureUnit](tx).Create(r.Context(), record); err != nil {
				return err
			}
			return replaceStructureUnitCourseAssociations(tx, record, assoc)
		}); err != nil {
			slog.Error("structure unit create failed", "err", err)
			fieldErrors["_form"] = err
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		loc, err := structureEditRedirectURL(r.Context())
		if err != nil {
			fieldErrors["_form"] = err
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, loc, http.StatusMovedPermanently)
	})
}

func handleStructureUnitUpdate(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			return
		}
		var progPatch structureUnitProgramIDFromPathPatcher
		values, fieldErrors = progPatch.Patch(*v, r, values, fieldErrors)
		if len(fieldErrors) != 0 {
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		parentID, ok := structureParentProgramID(r.Context())
		if !ok {
			http.Error(w, "missing program", http.StatusBadRequest)
			return
		}
		unitIDStr := r.PathValue("unitId")
		unitID, err := strconv.ParseUint(unitIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid unit", http.StatusBadRequest)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("structure unit update: db from context", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		existing, err := gorm.G[ProgramStructureUnit](db).Where("id = ? AND program_id = ?", unitID, parentID).First(r.Context())
		if err != nil {
			http.NotFound(w, r)
			return
		}
		regular, assoc := splitStructureUnitFormValues(values)
		vals := maps.Clone(regular)

		var patch ProgramStructureUnit
		if err := views.PopulateFromMap(&patch, vals); err != nil {
			fieldErrors["_form"] = err
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		existing.TermNumber = patch.TermNumber
		existing.OptionalCourseCount = patch.OptionalCourseCount
		existing.ProgramID = parentID
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			return replaceStructureUnitCourseAssociations(tx, &existing, assoc)
		}); err != nil {
			slog.Error("structure unit update failed", "err", err)
			fieldErrors["_form"] = err
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		loc, err := structureEditRedirectURL(r.Context())
		if err != nil {
			fieldErrors["_form"] = err
			ctx := views.ContextWithErrorsAndValues(r.Context(), values, fieldErrors)
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, loc, http.StatusMovedPermanently)
	})
}

func handleStructureUnitDelete(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parentID, ok := structureParentProgramID(r.Context())
		if !ok {
			http.Error(w, "missing program", http.StatusBadRequest)
			return
		}
		unitIDStr := r.PathValue("unitId")
		unitID, err := strconv.ParseUint(unitIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid unit", http.StatusBadRequest)
			return
		}
		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("structure unit delete: db from context", "error", dberr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		existing, err := gorm.G[ProgramStructureUnit](db).Where("id = ? AND program_id = ?", unitID, parentID).First(r.Context())
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&existing).Association("CompulsoryCourses").Clear(); err != nil {
				return err
			}
			if err := tx.Model(&existing).Association("OptionalCourseSelectionPool").Clear(); err != nil {
				return err
			}
			return tx.Delete(&existing).Error
		}); err != nil {
			slog.Error("structure unit delete failed", "err", err)
			ctx := views.ContextWithErrorsAndValues(r.Context(), nil, map[string]error{"_form": err})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		loc, err := structureEditRedirectURL(r.Context())
		if err != nil {
			ctx := views.ContextWithErrorsAndValues(r.Context(), nil, map[string]error{"_form": err})
			v.RenderPage(w, r.WithContext(ctx))
			return
		}
		views.HtmxRedirect(w, r, loc, http.StatusSeeOther)
	})
}
