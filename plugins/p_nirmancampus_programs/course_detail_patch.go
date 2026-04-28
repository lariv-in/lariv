package p_nirmancampus_programs

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/plugins/p_users"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/views"
	"gorm.io/gorm"
)

const courseDetailProgramPlacementsContextKey = "course_program_placements_table"

// CourseProgramPlacement is one row: a program structure unit where this course appears,
// with whether it is compulsory or in the optional pool.
type CourseProgramPlacement struct {
	Program    Program
	TermNumber uint
	Kind       string
}

func init() {
	registerCourseDetailProgramPlacementsPatch()
}

func appendPlacement(rows *[]CourseProgramPlacement, units []ProgramStructureUnit, kind string) {
	for i := range units {
		u := units[i]
		*rows = append(*rows, CourseProgramPlacement{
			Program:    u.Program,
			TermNumber: u.TermNumber,
			Kind:       kind,
		})
	}
}

func sortCourseProgramPlacements(rows []CourseProgramPlacement) {
	sort.Slice(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.Program.Name != b.Program.Name {
			return a.Program.Name < b.Program.Name
		}
		if a.TermNumber != b.TermNumber {
			return a.TermNumber < b.TermNumber
		}
		return a.Kind < b.Kind
	})
}

func queryStructureUnitsForCourse(
	db *gorm.DB,
	courseID uint,
	kindJoinTable string,
	programIDSubquery *gorm.DB,
) ([]ProgramStructureUnit, error) {
	q := db.Model(&ProgramStructureUnit{}).
		Preload("Program").
		Joins("INNER JOIN "+kindJoinTable+" AS j ON j.program_structure_unit_id = program_structure_units.id AND j.course_id = ?", courseID).
		Where("program_structure_units.deleted_at IS NULL")
	if programIDSubquery != nil {
		q = q.Where("program_structure_units.program_id IN (?)", programIDSubquery)
	}
	var units []ProgramStructureUnit
	if err := q.Find(&units).Error; err != nil {
		return nil, err
	}
	return units, nil
}

// attachCourseProgramPlacementsContext loads ProgramStructureUnits that reference the
// course (compulsory or optional pool), scoped like ProgramScopeByRole for students.
type attachCourseProgramPlacementsContext struct{}

func (attachCourseProgramPlacementsContext) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		course, ok := r.Context().Value("course").(courses.Course)
		if !ok || course.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}

		db, dberr := getters.DBFromContext(r.Context())
		if dberr != nil {
			slog.Error("attachCourseProgramPlacementsContext: db from context", "error", dberr)
			next.ServeHTTP(w, r)
			return
		}

		user, roleName := p_users.UserAndRoleFromContext(r.Context(), "CourseProgramPlacements")

		var programIDSubquery *gorm.DB
		switch roleName {
		case "superuser", "admin", "unassigned":
			programIDSubquery = nil
		case "student":
			email := strings.TrimSpace(user.Email)
			if email == "" {
				programIDSubquery = db.Table("academic_records").Select("program_id").Where("1 = 0")
				break
			}
			studentSub := db.Model(&p_nirmancampus_students.Student{}).
				Select("id").
				Where("email = ?", email)
			programIDSubquery = db.Table("academic_records").
				Select("program_id").
				Where("student_id IN (?)", studentSub).
				Where("deleted_at IS NULL")
		default:
			programIDSubquery = db.Table("academic_records").Select("program_id").Where("1 = 0")
		}

		var rows []CourseProgramPlacement

		compulsory, err := queryStructureUnitsForCourse(db, course.ID, "program_structure_unit_compulsory_courses", programIDSubquery)
		if err != nil {
			slog.Error("attachCourseProgramPlacementsContext: compulsory query", "error", err)
			next.ServeHTTP(w, r)
			return
		}
		appendPlacement(&rows, compulsory, "Compulsory")

		optional, err := queryStructureUnitsForCourse(db, course.ID, "program_structure_unit_optional_courses", programIDSubquery)
		if err != nil {
			slog.Error("attachCourseProgramPlacementsContext: optional query", "error", err)
			next.ServeHTTP(w, r)
			return
		}
		appendPlacement(&rows, optional, "Optional pool")

		sortCourseProgramPlacements(rows)

		ol := components.ObjectList[CourseProgramPlacement]{
			Items:    rows,
			Number:   1,
			NumPages: 1,
			Total:    uint64(len(rows)),
		}
		ctx := context.WithValue(r.Context(), courseDetailProgramPlacementsContextKey, ol)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func courseDetailProgramPlacementsSection() components.PageInterface {
	return &components.DataTable[CourseProgramPlacement]{
		Page:        components.Page{Key: "programs.CourseDetailProgramPlacements"},
		UID:         "course-detail-program-placements",
		Title:       "Programs",
		Classes:     "w-full mt-4",
		Data:        getters.Key[components.ObjectList[CourseProgramPlacement]](courseDetailProgramPlacementsContextKey),
		DefaultView: "Grid",
		RowAttr: getters.RowAttrNavigate(lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
			"id": getters.Any(getters.Key[uint]("$row.Program.ID")),
		})),
		Columns: []components.TableColumn{
			{
				Label: "Program",
				Name:  "Program.Name",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: ProgramDisplayLabel(
							getters.Key[string]("$row.Program.Name"),
							getters.Key[string]("$row.Program.University"),
						),
					},
				},
			},
			{
				Label: "Term",
				Name:  "TermNumber",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.TermNumber"))),
					},
				},
			},
			{
				Label: "Role",
				Name:  "Kind",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.Key[string]("$row.Kind")},
				},
			},
		},
	}
}

func registerCourseDetailProgramPlacementsPatch() {
	lago.RegistryView.Patch("courses.DetailView", func(v *views.View) *views.View {
		return v.InsertLayerAfter("courses.detail", "programs.course_detail_placements", attachCourseProgramPlacementsContext{})
	})

	lago.RegistryPage.Patch("courses.CourseDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			log.Panic("courses.CourseDetail was not ShellScaffold")
		}
		// courses/pages_detail.go uses *ContainerColumn; ReplaceChild must use pointer type or patch never matches.
		components.ReplaceChild[*components.ContainerColumn](scaffold, "courses.CourseDetailContent", func(column *components.ContainerColumn) *components.ContainerColumn {
			column.Children = append(column.Children, courseDetailProgramPlacementsSection())
			return column
		})
		return scaffold
	})
}
