package p_nirmancampus_programs

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/plugins/p_users"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

func fieldDBName[T any](db *gorm.DB, fieldName string) (string, bool) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return "", false
	}
	if stmt.Schema == nil {
		return "", false
	}
	field := stmt.Schema.LookUpField(fieldName)
	if field == nil {
		return "", false
	}
	return field.DBName, true
}

// queryPatcherUniversity filters Program list by University from $get[param].
type queryPatcherUniversity struct {
	Param string
}

func (p queryPatcherUniversity) Patch(_ views.View, r *http.Request, q gorm.ChainInterface[Program]) gorm.ChainInterface[Program] {
	getMap, ok := r.Context().Value("$get").(map[string]any)
	if !ok {
		return q
	}

	raw, ok := getMap[p.Param]
	if !ok {
		return q
	}
	value, ok := raw.(string)
	if !ok {
		return q
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return q
	}

	db := r.Context().Value("$db").(*gorm.DB)
	col, ok := fieldDBName[Program](db, "University")
	if !ok {
		return q
	}

	return q.Where(col+" = ?", value)
}

// queryPatcherProgramType filters Program list by ProgramType from $get[param].
type queryPatcherProgramType struct {
	Param string
}

func (p queryPatcherProgramType) Patch(_ views.View, r *http.Request, q gorm.ChainInterface[Program]) gorm.ChainInterface[Program] {
	getMap, ok := r.Context().Value("$get").(map[string]any)
	if !ok {
		return q
	}

	raw, ok := getMap[p.Param]
	if !ok {
		return q
	}
	value, ok := raw.(string)
	if !ok {
		return q
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return q
	}

	db := r.Context().Value("$db").(*gorm.DB)
	col, ok := fieldDBName[Program](db, "ProgramType")
	if !ok {
		return q
	}

	return q.Where(col+" = ?", value)
}

type queryPatcherPreloadProgramStructureUnits struct{}

func (queryPatcherPreloadProgramStructureUnits) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[Program]) gorm.ChainInterface[Program] {
	return q.Preload("ProgramMedia", func(pb gorm.PreloadBuilder) error {
		pb.Order("language ASC")
		return nil
	}).Preload("ProgramStructureUnits", func(pb gorm.PreloadBuilder) error {
		pb.Order("term_number ASC")
		return nil
	}).Preload("ProgramStructureUnits.CompulsoryCourses", nil).
		Preload("ProgramStructureUnits.OptionalCourseSelectionPool", nil)
}

// queryPatcherProgramMediaOrder sorts languages for the program media multi-select list.
type queryPatcherProgramMediaOrder struct{}

func (queryPatcherProgramMediaOrder) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[ProgramMedia]) gorm.ChainInterface[ProgramMedia] {
	return q.Order("language ASC")
}

type queryPatcherPreloadStructureUnitCourseAssociations struct{}

func (queryPatcherPreloadStructureUnitCourseAssociations) Patch(_ views.View, _ *http.Request, q gorm.ChainInterface[ProgramStructureUnit]) gorm.ChainInterface[ProgramStructureUnit] {
	return q.Preload("CompulsoryCourses", nil).Preload("OptionalCourseSelectionPool", nil)
}

// ProgramScopeByRole restricts program queries by role (list/detail/update/delete).
type programScopeByRole struct{}

func (programScopeByRole) Patch(_ views.View, r *http.Request, q gorm.ChainInterface[Program]) gorm.ChainInterface[Program] {
	ctx := r.Context()

	rawUser := ctx.Value("$user")
	if rawUser == nil {
		slog.Error("ProgramScopeByRole: missing $user in context – auth layer not applied?")
		panic("ProgramScopeByRole: $user is nil in context")
	}
	user, ok := rawUser.(p_users.User)
	if !ok {
		slog.Error("ProgramScopeByRole: $user has unexpected type",
			"type", fmt.Sprintf("%T", rawUser),
		)
		panic("ProgramScopeByRole: $user has wrong type in context")
	}

	rawRole := ctx.Value("$role")
	if rawRole == nil {
		slog.Error("ProgramScopeByRole: missing $role in context – auth layer not applied?")
		panic("ProgramScopeByRole: $role is nil in context")
	}
	roleName, ok := rawRole.(string)
	if !ok {
		slog.Error("ProgramScopeByRole: $role has unexpected type",
			"type", fmt.Sprintf("%T", rawRole),
		)
		panic("ProgramScopeByRole: $role has wrong type in context")
	}

	dbVal := ctx.Value("$db")
	db, ok := dbVal.(*gorm.DB)
	if !ok || db == nil {
		slog.Error("ProgramScopeByRole: missing or invalid $db in context",
			"type", fmt.Sprintf("%T", dbVal),
		)
		panic("ProgramScopeByRole: $db is nil or wrong type in context")
	}

	switch roleName {
	case "superuser", "admin", "unassigned":
		return q
	case "student":
		email := strings.TrimSpace(user.Email)
		if email == "" {
			return q.Where("1 = 0")
		}
		studentSub := db.Model(&p_nirmancampus_students.Student{}).
			Select("id").
			Where("email = ?", email)
		programSub := db.Table("academic_records").
			Select("program_id").
			Where("student_id IN (?)", studentSub).
			Where("deleted_at IS NULL")
		return q.Where("programs.id IN (?)", programSub)
	default:
		return q.Where("1 = 0")
	}
}
