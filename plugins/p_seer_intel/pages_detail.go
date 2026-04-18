package p_seer_intel

import (
	"context"
	"fmt"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/pgvector/pgvector-go"
)

func formatIntelEmbeddingLabel(v *pgvector.Vector) string {
	if v == nil {
		return "—"
	}
	vec := v.Slice()
	if len(vec) == 0 {
		return "—"
	}
	return fmt.Sprintf("%d dimensions", len(vec))
}

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
								Title: "Embedding",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.Map(getters.Key[Intel]("intel"), func(_ context.Context, in Intel) (string, error) {
											return formatIntelEmbeddingLabel(in.Embedding), nil
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
