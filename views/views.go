package views

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"reflect"
	"strings"
	"syscall"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
)

type View struct {
	PageName   string
	PageLookup func(name string) (components.PageInterface, bool)
	Layers     []registry.Pair[string, Layer]
}

func (v *View) GetHandler() http.Handler {
	var handler http.Handler = http.HandlerFunc(v.RenderPage)
	for i := len(v.Layers) - 1; i >= 0; i-- {
		handler = v.Layers[i].Value.Next(*v, handler)
	}
	return handler
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.GetHandler().ServeHTTP(w, r)
}

func (v *View) GetPage() (components.PageInterface, bool) {
	return v.PageLookup(v.PageName)
}

func (v *View) RenderPage(w http.ResponseWriter, r *http.Request) {
	page, isPagePresent := v.GetPage()
	if !isPagePresent {
		http.NotFound(w, r)
		return
	}
	ctx := r.Context()
	if errs, ok := ctx.Value(getters.ContextKeyError).(map[string]error); ok && len(errs) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
	err := components.Render(page, ctx).Render(w)
	if err != nil {
		// Do not panic when the client is already gone (common with slow layers + devtools, live reload).
		if isBenignResponseWriteError(err) {
			slog.Debug("views: render after client closed", "error", err)
			return
		}
		panic(err)
	}
}

// isBenignResponseWriteError is true for typical disconnect errors from ResponseWriter.
func isBenignResponseWriteError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, syscall.EPIPE) {
		return true
	}
	if errors.Is(err, syscall.ECONNRESET) {
		return true
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	if errors.Is(err, net.ErrClosed) {
		return true
	}
	s := err.Error()
	return strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "use of closed network connection")
}

// ParseForm finds the first FormComponent in the view's page, parses the request form,
// and returns the values and field errors. Returns true if parsing failed (error already written to w).
func (v *View) ParseForm(w http.ResponseWriter, r *http.Request) (map[string]any, map[string]error, error) {
	page, _ := v.GetPage()
	var parent components.ParentInterface
	// If the page already implements ParentInterface, use it directly.
	if p, ok := page.(components.ParentInterface); ok {
		parent = p
	} else {
		// Many scaffolds (ShellScaffold, ShellTopbarScaffold, etc.) implement
		// ParentInterface only on the pointer type, but are stored as values.
		// For traversal we can safely take the address of the value here.
		val := reflect.ValueOf(page)
		if val.Kind() != reflect.Pointer {
			ptr := val.Addr()
			if p, ok := ptr.Interface().(components.ParentInterface); ok {
				parent = p
			}
		}
	}

	if parent == nil {
		// Log the actual interface assertion issue so it is visible in logs.
		slog.Error("view page does not implement components.ParentInterface",
			"pageType", reflect.TypeOf(page),
			"pageName", v.PageName)
		http.Error(w, "Internal Server Error: No form container found", http.StatusInternalServerError)
		return nil, nil, errors.New("Internal Server Error: No form container found")
	}

	forms := components.FindChildren[components.FormInterface](parent)
	if len(forms) == 0 {
		http.Error(w, "Internal Server Error: No form found", http.StatusInternalServerError)
		return nil, nil, errors.New("Internal Server Error: No form found")
	}
	values, fieldErrors, err := forms[0].ParseForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, nil, err
	}
	if fieldErrors == nil {
		fieldErrors = make(map[string]error)
	}
	return values, fieldErrors, nil
}

func (v *View) WithLayer(name string, layer Layer) *View {
	// Append layer; keys are labels only and are not required to be unique.
	v.Layers = append(v.Layers, registry.Pair[string, Layer]{Key: name, Value: layer})
	return v
}

// InsertLayerBefore inserts a layer with the given name immediately
// before the first layer whose Key matches beforeName. If no such
// layer exists, it appends it to the end.
func (v *View) InsertLayerBefore(beforeName, name string, layer Layer) *View {
	p := registry.Pair[string, Layer]{Key: name, Value: layer}
	for i, mw := range v.Layers {
		if mw.Key == beforeName {
			v.Layers = append(v.Layers[:i], append([]registry.Pair[string, Layer]{p}, v.Layers[i:]...)...)
			return v
		}
	}
	// Fallback: behave like WithLayer when beforeName is not found.
	return v.WithLayer(name, layer)
}

// InsertLayerAfter inserts a layer with the given name immediately
// after the first layer whose Key matches afterName. If no such
// layer exists, it appends it to the end.
func (v *View) InsertLayerAfter(afterName, name string, layer Layer) *View {
	p := registry.Pair[string, Layer]{Key: name, Value: layer}
	for i, mw := range v.Layers {
		if mw.Key == afterName {
			// Insert after index i.
			pos := i + 1
			if pos >= len(v.Layers) {
				v.Layers = append(v.Layers, p)
			} else {
				v.Layers = append(v.Layers[:pos], append([]registry.Pair[string, Layer]{p}, v.Layers[pos:]...)...)
			}
			return v
		}
	}
	// Fallback: behave like WithLayer when afterName is not found.
	return v.WithLayer(name, layer)
}

func (v *View) WithLayers(layers ...registry.Pair[string, Layer]) *View {
	for _, layer := range layers {
		v.WithLayer(layer.Key, layer.Value)
	}
	return v
}

func (v *View) PatchLayers(layers ...registry.Pair[string, func(Layer) Layer]) *View {
	for _, layer := range layers {
		for i, mw := range v.Layers {
			if mw.Key == layer.Key {
				v.Layers[i].Value = layer.Value(mw.Value)
			}
		}
	}
	return v
}

func (v *View) PatchLayer(name string, patcher func(Layer) Layer) *View {
	for i, mw := range v.Layers {
		if mw.Key == name {
			v.Layers[i].Value = patcher(mw.Value)
		}
	}
	return v
}
