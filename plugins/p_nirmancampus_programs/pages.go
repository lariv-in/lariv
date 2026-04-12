package p_nirmancampus_programs

import (
	"context"
	"fmt"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

// ProgramDisplayLabel returns "Name (University label)" using [UniversityChoices]; empty university key → name only.
func ProgramDisplayLabel(nameGetter, universityKeyGetter getters.Getter[string]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		name, err := nameGetter(ctx)
		if err != nil {
			return "", err
		}
		ukey, errU := universityKeyGetter(ctx)
		if errU != nil || ukey == "" {
			return name, nil
		}
		if p, ok := registry.PairFromPairs(ukey, UniversityChoices); ok {
			return fmt.Sprintf("%s (%s)", name, p.Value), nil
		}
		return fmt.Sprintf("%s (%s)", name, ukey), nil
	}
}

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerDeletePages()
	registerSelectionPages()
	registerProgramMediaMultiSelectPages()
	registerStructurePages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("programs.ProgramMenu", &components.SidebarMenu{
		Title: getters.Static("Programs"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Programs"),
				Url:   lago.RoutePath("programs.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Courses"),
				Url:   lago.RoutePath("courses.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("programs.ProgramDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Program: %s", getters.Any(getters.Key[string]("program.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all Programs"),
			Url:   lago.RoutePath("programs.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Program Detail"),
				Url: lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("program.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit Program"),
				Url: lago.RoutePath("programs.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("program.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit program structure"),
				Url: lago.RoutePath("programs.StructureEditRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("program.ID")),
				}),
			},
		},
	})
}
