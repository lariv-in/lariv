package p_courses_teachers

import (
	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
	"github.com/lariv-in/p_courses"
	"github.com/lariv-in/p_teachers"
)

const (
	teachersFieldName = "Teachers"
	coursesFieldName  = "Courses"
)

func courseTeachersInput() *components.ContainerError {
	return &components.ContainerError{
		Error: getters.GetterKey[error]("$error." + teachersFieldName),
		Children: []components.PageInterface{
			&components.InputManyToMany[p_teachers.Teacher]{
				Label: "Teachers",
				Name:  teachersFieldName,
				Getter: getters.GetterJoinAssociationList[CourseTeacher, p_teachers.Teacher](
					getters.IfOrElseGetter(getters.GetterKey[uint]("$in.ID"), getters.GetterStatic(uint(0))),
					"CourseID",
					"TeacherID",
					"code ASC",
					"User",
				),
				Url: lago.GetterRoutePath("teachers.MultiSelectRoute", nil),
				Display: getters.GetterFormat(
					"%s (%s)",
					getters.GetterAny(getters.GetterKey[string]("$in.User.Name")),
					getters.GetterAny(getters.GetterKey[string]("$in.Code")),
				),
				Placeholder: "Select teachers...",
			},
		},
	}
}

func teacherCoursesInput() components.ContainerError {
	return components.ContainerError{
		Error: getters.GetterKey[error]("$error." + coursesFieldName),
		Children: []components.PageInterface{
			&components.InputManyToMany[p_courses.Course]{
				Label: "Courses",
				Name:  coursesFieldName,
				Getter: getters.GetterJoinAssociationList[CourseTeacher, p_courses.Course](
					getters.IfOrElseGetter(getters.GetterKey[uint]("$in.ID"), getters.GetterStatic(uint(0))),
					"TeacherID",
					"CourseID",
					"name ASC",
				),
				Url: lago.GetterRoutePath("courses.MultiSelectRoute", nil),
				Display: getters.GetterFormat(
					"%s (%s)",
					getters.GetterAny(getters.GetterKey[string]("$in.Name")),
					getters.GetterAny(getters.GetterKey[string]("$in.Code")),
				),
				Placeholder: "Select courses...",
			},
		},
	}
}

func courseTeachersFilterInput() *components.InputManyToMany[p_teachers.Teacher] {
	return &components.InputManyToMany[p_teachers.Teacher]{
		Label: "Teachers",
		Name:  teachersFieldName,
		Getter: getters.GetterAssociationList[p_teachers.Teacher](
			getters.GetterContextAssociationIDs(getters.ContextKeyGet, teachersFieldName),
			"code ASC",
			"User",
		),
		Url: lago.GetterRoutePath("teachers.MultiSelectRoute", nil),
		Display: getters.GetterFormat(
			"%s (%s)",
			getters.GetterAny(getters.GetterKey[string]("$in.User.Name")),
			getters.GetterAny(getters.GetterKey[string]("$in.Code")),
		),
		Placeholder: "Filter by teachers...",
	}
}

func teacherCoursesFilterInput() *components.InputManyToMany[p_courses.Course] {
	return &components.InputManyToMany[p_courses.Course]{
		Label: "Courses",
		Name:  coursesFieldName,
		Getter: getters.GetterAssociationList[p_courses.Course](
			getters.GetterContextAssociationIDs(getters.ContextKeyGet, coursesFieldName),
			"name ASC",
		),
		Url: lago.GetterRoutePath("courses.MultiSelectRoute", nil),
		Display: getters.GetterFormat(
			"%s (%s)",
			getters.GetterAny(getters.GetterKey[string]("$in.Name")),
			getters.GetterAny(getters.GetterKey[string]("$in.Code")),
		),
		Placeholder: "Filter by courses...",
	}
}

func teachersDetailSection() *components.LabelInline {
	return &components.LabelInline{
		Title:   "Teachers",
		Classes: "items-start flex-col",
		Children: []components.PageInterface{
			&components.FieldList{
				Classes: "flex flex-col gap-1",
				Getter: getters.GetterAny(getters.GetterJoinAssociationList[CourseTeacher, p_teachers.Teacher](
					getters.IfOrElseGetter(getters.GetterKey[uint]("$in.ID"), getters.GetterStatic(uint(0))),
					"CourseID",
					"TeacherID",
					"code ASC",
					"User",
				)),
				Children: []components.PageInterface{
					&components.ButtonLink{
						GetterLabel: getters.GetterFormat(
							"%s (%s)",
							getters.GetterAny(getters.GetterKey[string]("$row.User.Name")),
							getters.GetterAny(getters.GetterKey[string]("$row.Code")),
						),
						Link: lago.GetterRoutePath("teachers.DetailRoute", map[string]getters.Getter[any]{
							"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
						}),
						Classes: "btn-link btn-sm justify-start px-0 min-h-0 h-auto",
					},
				},
			},
		},
	}
}

func coursesDetailSection() components.LabelInline {
	return components.LabelInline{
		Title:   "Courses",
		Classes: "items-start flex-col",
		Children: []components.PageInterface{
			&components.FieldList{
				Classes: "flex flex-col gap-1",
				Getter: getters.GetterAny(getters.GetterJoinAssociationList[CourseTeacher, p_courses.Course](
					getters.IfOrElseGetter(getters.GetterKey[uint]("$in.ID"), getters.GetterStatic(uint(0))),
					"TeacherID",
					"CourseID",
					"name ASC",
				)),
				Children: []components.PageInterface{
					&components.ButtonLink{
						GetterLabel: getters.GetterFormat(
							"%s (%s)",
							getters.GetterAny(getters.GetterKey[string]("$row.Name")),
							getters.GetterAny(getters.GetterKey[string]("$row.Code")),
						),
						Link: lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
							"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
						}),
						Classes: "btn-link btn-sm justify-start px-0 min-h-0 h-auto",
					},
				},
			},
		},
	}
}

func patchCoursePages() {
	lago.RegistryPage.Patch("courses.CourseFilter", func(page components.PageInterface) components.PageInterface {
		form, ok := page.(*components.FormComponent[p_courses.Course])
		if !ok {
			panic("Base page for courses.CourseFilter was not FormComponent[p_courses.Course]")
		}
		form.ChildrenInput = append(form.ChildrenInput, courseTeachersFilterInput())
		return form
	})

	lago.RegistryPage.Patch("courses.CourseCreateForm", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for courses.CourseCreateForm was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "courses.CourseFormFieldsBody", func(column *components.ContainerColumn) *components.ContainerColumn {
			column.Children = append(column.Children, courseTeachersInput())
			return column
		})
		return scaffold
	})

	lago.RegistryPage.Patch("courses.CourseUpdateForm", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for courses.CourseUpdateForm was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "courses.CourseFormFieldsBody", func(column *components.ContainerColumn) *components.ContainerColumn {
			column.Children = append(column.Children, courseTeachersInput())
			return column
		})
		return scaffold
	})

	lago.RegistryPage.Patch("courses.CourseDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for courses.CourseDetail was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "courses.CourseDetailContent", func(column *components.ContainerColumn) *components.ContainerColumn {
			column.Children = append(column.Children, teachersDetailSection())
			return column
		})
		return scaffold
	})
}

func patchTeacherPages() {
	lago.RegistryPage.Patch("teachers.TeacherFilter", func(page components.PageInterface) components.PageInterface {
		form, ok := page.(*components.FormComponent[p_teachers.Teacher])
		if !ok {
			panic("Base page for teachers.TeacherFilter was not FormComponent[p_teachers.Teacher]")
		}
		form.ChildrenInput = append(form.ChildrenInput, teacherCoursesFilterInput())
		return form
	})

	lago.RegistryPage.Patch("teachers.TeacherCreateForm", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for teachers.TeacherCreateForm was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "teachers.TeacherFormFieldsBody", func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children, teacherCoursesInput())
			return column
		})
		return scaffold
	})

	lago.RegistryPage.Patch("teachers.TeacherUpdateForm", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for teachers.TeacherUpdateForm was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "teachers.TeacherFormFieldsBody", func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children, teacherCoursesInput())
			return column
		})
		return scaffold
	})

	lago.RegistryPage.Patch("teachers.TeacherDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for teachers.TeacherDetail was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "teachers.TeacherDetailContent", func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children, coursesDetailSection())
			return column
		})
		return scaffold
	})
}

func init() {
	patchCoursePages()
	patchTeacherPages()
}
