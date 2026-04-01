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

// queryPatcherStructureUnitForContextProgram restricts ProgramStructureUnit loads to the program in context.
var queryPatcherStructureUnitForContextProgram views.QueryPatcher = func(_ *views.View, r *http.Request, query *gorm.DB) *gorm.DB {
	stmt := &gorm.Statement{DB: query}
	if err := stmt.Parse(&ProgramStructureUnit{}); err != nil {
		return query
	}
	if stmt.Schema == nil || stmt.Schema.Table != "program_structure_units" {
		return query
	}
	p, ok := r.Context().Value("program").(Program)
	if !ok || p.ID == 0 {
		return query.Where("1 = 0")
	}
	return query.Where("program_id = ?", p.ID)
}

// middlewareProgramsStructureLoadProgram loads the program from {id} with role scope and preloads structure units.
func middlewareProgramsStructureLoadProgram(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		db := r.Context().Value("$db").(*gorm.DB)
		q := db.Model(&Program{}).
			Preload("ProgramStructureUnits", preloadProgramStructureUnitCourseAssociations)
		q = ProgramScopeByRole(nil, r, q)
		var p Program
		if err := q.First(&p, id).Error; err != nil {
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

// formPatcherStructureUnitProgramIDFromPath forces ProgramID from the program path param (do not trust the client).
func formPatcherStructureUnitProgramIDFromPath(_ *views.View, r *http.Request, m map[string]any, formErrors map[string]error) (map[string]any, map[string]error) {
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
		if v.HasErrors(fieldErrors) {
			v.RenderWithErrors(w, r, fieldErrors, values)
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
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}
		if record.ProgramID != parentID {
			fieldErrors["ProgramID"] = fmt.Errorf("invalid program")
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}
		db := r.Context().Value("$db").(*gorm.DB)
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(record).Error; err != nil {
				return err
			}
			return replaceStructureUnitCourseAssociations(tx, record, assoc)
		}); err != nil {
			slog.Error("structure unit create failed", "err", err)
			fieldErrors["_form"] = err
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}
		loc, err := structureEditRedirectURL(r.Context())
		if err != nil {
			fieldErrors["_form"] = err
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}
		lago.Redirect(w, r, loc)
	})
}

func handleStructureUnitUpdate(v *views.View) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values, fieldErrors, err := v.ParseForm(w, r)
		if err != nil {
			return
		}
		if v.HasErrors(fieldErrors) {
			v.RenderWithErrors(w, r, fieldErrors, values)
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
		db := r.Context().Value("$db").(*gorm.DB)
		var existing ProgramStructureUnit
		if err := db.Where("id = ? AND program_id = ?", unitID, parentID).First(&existing).Error; err != nil {
			http.NotFound(w, r)
			return
		}
		regular, assoc := splitStructureUnitFormValues(values)
		vals := maps.Clone(regular)

		var patch ProgramStructureUnit
		if err := views.PopulateFromMap(&patch, vals); err != nil {
			fieldErrors["_form"] = err
			v.RenderWithErrors(w, r, fieldErrors, values)
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
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}
		loc, err := structureEditRedirectURL(r.Context())
		if err != nil {
			fieldErrors["_form"] = err
			v.RenderWithErrors(w, r, fieldErrors, values)
			return
		}
		lago.Redirect(w, r, loc)
	})
}
