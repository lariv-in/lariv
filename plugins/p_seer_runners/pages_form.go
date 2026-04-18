package p_seer_runners

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

// runnerCreateFormDefaults seeds the create form so [components.FormComponent] binds $in (non-zero struct).
var runnerCreateFormDefaults = Runner{
	Duration: 5 * time.Minute,
	Kind:     "",
}

func runnerCreateFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "seer_runners.RunnerCreateFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Duration"),
				Children: []components.PageInterface{
					&components.InputDuration{
						Label:    "Duration",
						Name:     "Duration",
						Required: true,
						Getter:   getters.Ref(getters.Key[time.Duration]("$in.Duration")),
						Classes:  "w-full max-w-md",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Kind"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Kind",
						Name:     "Kind",
						Required: true,
						Getter:   getters.Key[string]("$in.Kind"),
						Classes:  "w-full max-w-md",
					},
				},
			},
		},
	}
}

func registerFormPages() {
	createFormName := getters.Static("seer_runners.RunnerCreateForm")

	lago.RegistryPage.Register("seer_runners.RunnerCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_runners.RunnerMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createFormName,
				ActionURL: lago.RoutePath("seer_runners.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Runner]{
						Getter: getters.Static(runnerCreateFormDefaults),
						Attr:   getters.FormBubbling(createFormName),
						Title:  "Create runner",
						Subtitle: "Polling interval and kind label for orchestration (sources attach to runners separately).",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							runnerCreateFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save runner"},
						},
					},
				},
			},
		},
	})
}
