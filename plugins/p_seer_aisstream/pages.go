package p_seer_aisstream

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"gorm.io/datatypes"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func init() {
	registerAISStreamMenuPages()
	registerAISStreamTableAndDetail()
	registerAISStreamMapPages()
}

func registerAISStreamMenuPages() {
	lago.RegistryPage.Register("seer_aisstream.AppMenu", &components.SidebarMenu{
		Title: getters.Static("AISstream"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Messages"),
				Url:   lago.RoutePath("seer_aisstream.DefaultRoute", nil),
			},
			&aisStreamMapMenuLink{Page: components.Page{Key: "seer_aisstream.AppMenuMapLink"}},
		},
	})
}

func registerAISStreamTableAndDetail() {
	lago.RegistryPage.Register("seer_aisstream.MessageTablePage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_aisstream.AppMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[AISStreamMessage]{
				Page:    components.Page{Key: "seer_aisstream.MessageTableBody"},
				UID:     "seer-aisstream-messages",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[AISStreamMessage]]("aisStreamMessages"),
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_aisstream.MessageDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{Label: "Type", Name: "MessageType", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.MessageType")},
					}},
					{Label: "MMSI", Name: "MMSI", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.MMSI")},
					}},
					{Label: "Ship", Name: "ShipName", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.ShipName")},
					}},
					{Label: "Lon", Name: "Longitude", Children: []components.PageInterface{
						&components.FieldText{Getter: aisPointAxisString(getters.Key[lago.PGPoint]("$row.Position"), true)},
					}},
					{Label: "Lat", Name: "Latitude", Children: []components.PageInterface{
						&components.FieldText{Getter: aisPointAxisString(getters.Key[lago.PGPoint]("$row.Position"), false)},
					}},
					{Label: "Received", Name: "ReceivedAt", Children: []components.PageInterface{
						&components.FieldText{Getter: aisTimeString(getters.Key[time.Time]("$row.ReceivedAt"))},
					}},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_aisstream.MessageDetailMenu", &components.SidebarMenu{
		Title: getters.Static("AIS Message"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("All messages"),
			Url:   lago.RoutePath("seer_aisstream.MessageListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("View"),
				Url: lago.RoutePath("seer_aisstream.MessageDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("aisStreamMessage.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("seer_aisstream.MessageDetailPage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_aisstream.MessageDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[AISStreamMessage]{
				Getter: getters.Key[AISStreamMessage]("aisStreamMessage"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page:    components.Page{Key: "seer_aisstream.MessageDetailBody"},
						Classes: "gap-3 w-full max-w-4xl",
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Format("AIS %s", getters.Any(getters.Key[string]("$in.MessageType")))},
							aisLbl("ID", getters.Format("%d", getters.Any(getters.Key[uint]("$in.ID")))),
							aisLbl("Message type", getters.Key[string]("$in.MessageType")),
							aisLbl("MMSI", getters.Key[string]("$in.MMSI")),
							aisLbl("Ship name", getters.Key[string]("$in.ShipName")),
							aisLbl("Received (UTC)", aisTimeString(getters.Key[time.Time]("$in.ReceivedAt"))),
							aisLbl("AIS time (UTC)", aisTimePtrString(getters.Key[*time.Time]("$in.TimeUTC"))),
							aisLbl("Longitude", aisPointAxisString(getters.Key[lago.PGPoint]("$in.Position"), true)),
							aisLbl("Latitude", aisPointAxisString(getters.Key[lago.PGPoint]("$in.Position"), false)),
							aisLbl("SOG (kn)", aisFloatPtrString(getters.Key[*float64]("$in.SOG"))),
							aisLbl("COG (deg)", aisFloatPtrString(getters.Key[*float64]("$in.COG"))),
							aisLbl("Heading (deg)", aisFloatPtrString(getters.Key[*float64]("$in.Heading"))),
							&aisTypedPayloadDisplay{Page: components.Page{Key: "seer_aisstream.TypedPayload"}},
							aisLbl("Raw metadata", aisJSONGetter(getters.Key[datatypes.JSON]("$in.RawMetadata"))),
							aisLbl("Raw message", aisJSONGetter(getters.Key[datatypes.JSON]("$in.RawMessage"))),
						},
					},
				},
			},
		},
	})
}

func aisLbl(title string, g getters.Getter[string]) *components.LabelInline {
	return &components.LabelInline{Title: title, Children: []components.PageInterface{
		&components.FieldText{Getter: g, Classes: "min-w-0 whitespace-pre-wrap break-all"},
	}}
}

func aisFloatPtrString(g getters.Getter[*float64]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		p, err := g(ctx)
		if err != nil || p == nil {
			return "", err
		}
		return strconv.FormatFloat(*p, 'f', 5, 64), nil
	}
}

func aisTimeString(g getters.Getter[time.Time]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		t, err := g(ctx)
		if err != nil || t.IsZero() {
			return "", err
		}
		return t.UTC().Format(time.RFC3339), nil
	}
}

func aisTimePtrString(g getters.Getter[*time.Time]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		t, err := g(ctx)
		if err != nil || t == nil || t.IsZero() {
			return "", err
		}
		return t.UTC().Format(time.RFC3339), nil
	}
}

func aisPointAxisString(g getters.Getter[lago.PGPoint], longitude bool) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		p, err := g(ctx)
		if err != nil || !p.Valid {
			return "", err
		}
		v := p.P.Y
		if longitude {
			v = p.P.X
		}
		return strconv.FormatFloat(v, 'f', 6, 64), nil
	}
}

func aisJSONGetter(g getters.Getter[datatypes.JSON]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		raw, err := g(ctx)
		if err != nil || len(raw) == 0 {
			return "", err
		}
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			return string(raw), nil
		}
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return string(raw), nil
		}
		return string(b), nil
	}
}

type aisTypedPayloadDisplay struct{ components.Page }

func (e *aisTypedPayloadDisplay) GetKey() string     { return e.Key }
func (e *aisTypedPayloadDisplay) GetRoles() []string { return e.Roles }

func (e *aisTypedPayloadDisplay) Build(ctx context.Context) Node {
	in, ok := ctx.Value("$in").(map[string]any)
	if !ok {
		return Text("")
	}
	id, okID := in["ID"].(uint)
	messageType, okType := in["MessageType"].(string)
	if !okID || !okType || id == 0 || strings.TrimSpace(messageType) == "" {
		return Text("")
	}
	handler, ok := AISStreamMessageTypes.Get(messageType)
	if !ok || handler.Render == nil {
		return Text("")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Error("p_seer_aisstream: typed payload db", "error", err)
		return Div(Class("text-error"), Text("Could not load typed payload."))
	}
	payload, err := handler.Render(ctx, db, id)
	if err != nil {
		slog.Error("p_seer_aisstream: typed payload load", "error", err)
		return Div(Class("text-error"), Text("Could not load typed payload."))
	}
	if payload == "" {
		return Text("")
	}
	return Div(Class("flex flex-col gap-1"),
		Div(Class("font-semibold text-sm"), Text("Typed payload")),
		Pre(Class("bg-base-200 rounded-box p-3 overflow-auto text-xs whitespace-pre-wrap break-all"), Text(fmt.Sprint(payload))),
	)
}
