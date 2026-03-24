package p_announcements_semesters

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_announcements"
	"github.com/lariv-in/lago/p_semesters"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

const formFieldsBodyKey = "announcements.AnnouncementFormFieldsBody"

func announcementSemesterGetter(announcementIDGetter getters.Getter[uint]) getters.Getter[p_semesters.Semester] {
	return func(ctx context.Context) (p_semesters.Semester, error) {
		aid, err := announcementIDGetter(ctx)
		if err != nil || aid == 0 {
			return p_semesters.Semester{}, nil
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok || db == nil {
			return p_semesters.Semester{}, fmt.Errorf("announcementSemesterGetter: missing $db in context")
		}

		var details AnnouncementSemesterDetails
		err = db.Preload("Semester").Where("announcement_id = ?", aid).Take(&details).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return p_semesters.Semester{}, nil
			}
			return p_semesters.Semester{}, err
		}

		return details.Semester, nil
	}
}

func announcementSemesterNameFromRow() getters.Getter[string] {
	g := announcementSemesterGetter(getters.GetterKey[uint]("$row.ID"))
	return func(ctx context.Context) (string, error) {
		s, err := g(ctx)
		if err != nil || s.ID == 0 {
			return "", nil
		}
		return s.Name, nil
	}
}

func announcementSemesterNameFromIn() getters.Getter[string] {
	g := announcementSemesterGetter(getters.GetterKey[uint]("$in.ID"))
	return func(ctx context.Context) (string, error) {
		s, err := g(ctx)
		if err != nil || s.ID == 0 {
			return "", nil
		}
		return s.Name, nil
	}
}

// semesterEnvironmentDefaultGetter selects the semester whose [Start, End] contains time.Now(),
// matching announcementsListSemesterEnvQueryPatcher when the environment cookie has no semester.
func semesterEnvironmentDefaultGetter(ctx context.Context) (uint, error) {
	db, ok := ctx.Value("$db").(*gorm.DB)
	if !ok || db == nil {
		return 0, nil
	}
	id, ok := semesterEnvironmentDefault(db, time.Now())
	if !ok {
		return 0, nil
	}
	return id, nil
}

func semestersEnvOptionsGetterForEnvironment(ctx context.Context) ([]registry.Pair[uint, string], error) {
	db, ok := ctx.Value("$db").(*gorm.DB)
	if !ok || db == nil {
		return nil, fmt.Errorf("semestersEnvOptionsGetterForEnvironment: missing $db in context")
	}

	var semesters []p_semesters.Semester
	if err := db.Order(`"start" ASC`).Find(&semesters).Error; err != nil {
		return nil, err
	}

	options := make([]registry.Pair[uint, string], 0, len(semesters))
	for _, s := range semesters {
		options = append(options, registry.Pair[uint, string]{
			Key:   s.ID,
			Value: s.Name,
		})
	}
	return options, nil
}

func appendSemesterFormField(col components.ContainerColumn) components.ContainerColumn {
	col.Children = append(col.Children,
		&components.ContainerError{
			Error: getters.GetterKey[error]("$error.SemesterID"),
			Children: []components.PageInterface{
				&components.InputForeignKey[p_semesters.Semester]{
					Label:       "Semester",
					Name:        "SemesterID",
					Required:    true,
					Getter:      announcementSemesterGetter(getters.GetterKey[uint]("$in.ID")),
					Url:         lago.GetterRoutePath("semesters.SelectRoute", nil),
					Display:     getters.GetterKey[string]("$in.Name"),
					Placeholder: "Select a semester...",
				},
			},
		},
	)
	return col
}

func patchAnnouncementFormPages() {
	patchScaffold := func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("announcements form page was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, formFieldsBodyKey, appendSemesterFormField)
		return scaffold
	}

	lago.RegistryPage.Patch("announcements.AnnouncementCreateForm", patchScaffold)
	lago.RegistryPage.Patch("announcements.AnnouncementUpdateForm", patchScaffold)
}

func patchAnnouncementTablePage() {
	lago.RegistryPage.Patch("announcements.AnnouncementTable", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("announcements.AnnouncementTable was not ShellScaffold")
		}

		components.InsertChildBefore(scaffold, "announcements.AnnouncementTableBody", func(_ *components.DataTable[p_announcements.Announcement]) *components.Environment[uint] {
			return &components.Environment[uint]{
				Label:   "Semester",
				Key:     getters.GetterStatic("semester"),
				Options: semestersEnvOptionsGetterForEnvironment,
				Default: semesterEnvironmentDefaultGetter,
			}
		})

		return scaffold
	})
}

func patchAnnouncementDetailPage() {
	lago.RegistryPage.Patch("announcements.AnnouncementDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("announcements.AnnouncementDetail was not ShellScaffold")
		}

		components.InsertChildAfter(scaffold, "announcements.AnnouncementDetailTitle", func(_ *components.FieldTitle) *components.FieldSubtitle {
			return &components.FieldSubtitle{Getter: announcementSemesterNameFromIn()}
		})

		return scaffold
	})
}

func patchAnnouncementSelectionTable() {
	lago.RegistryPage.Patch("announcements.AnnouncementSelectionTable", func(page components.PageInterface) components.PageInterface {
		modal, ok := page.(*components.Modal)
		if !ok {
			panic("announcements.AnnouncementSelectionTable was not *components.Modal")
		}

		components.ReplaceChild(modal, "announcements.AnnouncementSelectionTableBody", func(table *components.DataTable[p_announcements.Announcement]) *components.DataTable[p_announcements.Announcement] {
			semCol := components.TableColumn{
				Label: "Semester",
				Name:  "Semester",
				Children: []components.PageInterface{
					&components.FieldText{Getter: announcementSemesterNameFromRow()},
				},
			}
			newCols := make([]components.TableColumn, 0, len(table.Columns)+1)
			newCols = append(newCols, table.Columns[0])
			newCols = append(newCols, semCol)
			newCols = append(newCols, table.Columns[1:]...)
			table.Columns = newCols
			return table
		})

		return modal
	})
}

func init() {
	patchAnnouncementFormPages()
	patchAnnouncementTablePage()
	patchAnnouncementDetailPage()
	patchAnnouncementSelectionTable()
	patchAnnouncementViews()
}
