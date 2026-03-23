package p_academicrecords_courses

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_academicrecords"
	"github.com/lariv-in/lago/p_courses"
)

const coursesFieldName = "Courses"

func academicRecordCoursesInput() *components.ContainerError {
	return &components.ContainerError{
		Error: getters.GetterKey[error]("$error." + coursesFieldName),
		Children: []components.PageInterface{
			&components.InputManyToMany[p_courses.Course]{
				Label: "Courses",
				Name:  coursesFieldName,
				Getter: getters.GetterJoinAssociationList[AcademicRecordCourse, p_courses.Course](
					getters.IfOrElseGetter(getters.GetterKey[uint]("$in.ID"), getters.GetterStatic(uint(0))),
					"AcademicRecordID",
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

func academicRecordCoursesFilterInput() *components.InputManyToMany[p_courses.Course] {
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

func academicRecordCoursesDetailSection() components.LabelInline {
	return components.LabelInline{
		Title:   "Courses",
		Classes: "items-start flex-col",
		Children: []components.PageInterface{
			&components.FieldManyToMany[p_courses.Course]{
				Getter: getters.GetterJoinAssociationList[AcademicRecordCourse, p_courses.Course](
					getters.IfOrElseGetter(getters.GetterKey[uint]("$in.ID"), getters.GetterStatic(uint(0))),
					"AcademicRecordID",
					"CourseID",
					"name ASC",
				),
				Display: getters.GetterFormat(
					"%s (%s)",
					getters.GetterAny(getters.GetterKey[string]("$in.Name")),
					getters.GetterAny(getters.GetterKey[string]("$in.Code")),
				),
				Link: lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Classes: "w-full",
			},
		},
	}
}

func patchAcademicRecordPages() {
	lago.RegistryPage.Patch("academicrecords.AcademicRecordFilter", func(page components.PageInterface) components.PageInterface {
		form, ok := page.(*components.FormComponent[p_academicrecords.AcademicRecord])
		if !ok {
			panic("Base page for academicrecords.AcademicRecordFilter was not FormComponent[p_academicrecords.AcademicRecord]")
		}
		form.ChildrenInput = append(form.ChildrenInput, academicRecordCoursesFilterInput())
		return form
	})

	lago.RegistryPage.Patch("academicrecords.AcademicRecordCreateForm", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for academicrecords.AcademicRecordCreateForm was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "academicrecords.AcademicRecordFormFieldsBody", func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children, academicRecordCoursesInput())
			return column
		})
		return scaffold
	})

	lago.RegistryPage.Patch("academicrecords.AcademicRecordUpdateForm", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for academicrecords.AcademicRecordUpdateForm was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "academicrecords.AcademicRecordFormFieldsBody", func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children, academicRecordCoursesInput())
			return column
		})
		return scaffold
	})

	lago.RegistryPage.Patch("academicrecords.AcademicRecordDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			panic("Base page for academicrecords.AcademicRecordDetail was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "academicrecords.AcademicRecordDetailContent", func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children, academicRecordCoursesDetailSection())
			return column
		})
		return scaffold
	})
}

func init() {
	patchAcademicRecordPages()
}
