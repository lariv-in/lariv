package components

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"time"

	"maragu.dev/gomponents"
)

type PageInterface interface {
	Build(context.Context) gomponents.Node
	GetKey() string
	GetRoles() []string
}

// Page struct defines fields that are common in all components
type Page struct {
	Key   string
	Roles []string
}

// #region agent log
func debugLogComponent(runID, hypothesisID, location, message string, data map[string]any) {
	f, err := os.OpenFile("/home/sandy/source_repos/lago/.cursor/debug-84938a.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(map[string]any{
		"sessionId":    "84938a",
		"runId":        runID,
		"hypothesisId": hypothesisID,
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	})
}

// #endregion

func Render(p PageInterface, ctx context.Context) gomponents.Node {
	roles := GetRequiredRoles(p)
	currentRole, _ := ctx.Value("$role").(string)
	allowed := roles == nil || slices.Contains(roles, currentRole)
	// #region agent log
	debugLogComponent("initial", "H3", "components/page.go:47", "component render role gate", map[string]any{
		"pageType":    reflect.TypeOf(p).String(),
		"roles":       roles,
		"currentRole": currentRole,
		"allowed":     allowed,
	})
	// #endregion

	if roles == nil {
		return p.Build(ctx)
	}

	if slices.Contains(roles, currentRole) {
		return p.Build(ctx)
	}
	return gomponents.Group{}
}

func GetRequiredRoles(p PageInterface) []string {
	v := reflect.ValueOf(p)
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	page, ok := v.FieldByName("Page").Interface().(Page)
	if !ok {
		return nil
	}
	return page.Roles
}
