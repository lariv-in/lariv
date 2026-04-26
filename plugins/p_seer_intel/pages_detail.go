package p_seer_intel

import (
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("seer_intel.IntelDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_intel.IntelDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Intel]{
				Getter: getters.Key[Intel]("intel"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "seer_intel.IntelDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{
								Title: "Summary",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter:  getters.Key[string]("$in.Summary"),
										Classes: "whitespace-pre-wrap",
									},
								},
							},
							&components.LabelInline{
								Title: "Datetime",
								Children: []components.PageInterface{
									&components.FieldDatetime{Getter: getters.Key[time.Time]("$in.Datetime")},
								},
							},
							&components.LabelInline{
								Title: "Kind",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Kind")},
								},
							},
							&components.LabelInline{
								Title: "Source",
								Children: []components.PageInterface{
									&components.FieldLink{
										Page:    components.Page{Key: "seer_intel.IntelDetailSourceLink"},
										Href:    getters.Key[string](intelSourceDetailHrefKey),
										Label:   getters.Static("Open source"),
										Classes: "link link-primary",
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
