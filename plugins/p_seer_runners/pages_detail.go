package p_seer_runners

import (
	"context"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("seer_runners.RunnerDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_runners.RunnerDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Runner]{
				Getter: getters.Key[Runner]("runner"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "seer_runners.RunnerDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.Format("Runner #%d", getters.Any(getters.Key[uint]("$in.ID"))),
							},
							&components.LabelInline{
								Title: "Kind",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Kind")},
								},
							},
							&components.LabelInline{
								Title: "Duration",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.Map(getters.Key[Runner]("runner"), func(_ context.Context, r Runner) (string, error) {
											return formatRunnerDuration(r.Duration), nil
										}),
									},
								},
							},
						},
					},
				},
			},
		},
	})
}
