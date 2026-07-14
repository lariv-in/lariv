package p_llm_assistant

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lariv-in/lago"
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
)

func assistantHistoryRowLabel(ctx context.Context) (string, error) {
	id, err := getters.Key[uint]("$row.ID")(ctx)
	if err != nil {
		return "", err
	}
	title, _ := getters.Key[string]("$row.Title")(ctx)
	title = strings.TrimSpace(title)
	if title == "" {
		title = "(untitled)"
	}
	updated, err := getters.Key[time.Time]("$row.UpdatedAt")(ctx)
	if err != nil {
		return fmt.Sprintf("#%d · %s", id, title), nil
	}
	return fmt.Sprintf("#%d · %s · %s", id, title, updated.UTC().Format(time.RFC3339)), nil
}

func registerAssistantHistoryPage() {
	registerPluginPage("llm_assistant.HistoryPage", &components.ShellScaffold{
		Page: components.Page{Key: "llm_assistant.HistoryPage"},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "llm_assistant.AssistantMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[LlmAssistantSession]{
				Page:    components.Page{Key: "llm_assistant.HistoryTableBody"},
				UID:     "llm-assistant-history-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[LlmAssistantSession]]("assistantSessions"),
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("llm_assistant.ChatSessionRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Chat",
						Children: []components.PageInterface{
							&components.FieldText{Getter: assistantHistoryRowLabel},
						},
					},
				},
			},
		},
	})
}

func init() {
	registerAssistantHistoryPage()
}
