package views

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"syscall"

	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

// View represents the core page controller coordinating middleware layers and page template rendering pipelines.
// It matches incoming HTTP requests, executes an ordered sequence of middleware [Layer] steps (e.g. data fetching, authentication, or updates),
// parses input parameters, and renders target HTML pages compiled from [components.PageInterface] trees.
//
// Use Cases:
//   - Defining request endpoints mapping back-office dashboards, user profiles, or transactional forms.
//   - Structuring reusable middleware layers to execute prior to HTML rendering phases.
//
// Example:
//
//	var UserDetailView = &views.View{
//		PageName:   "users.detail",
//		PageLookup: myRegistryLookup,
//		Layers: []registry.Pair[string, views.Layer]{
//			registry.NewPair("detail", views.LayerDetail[User]{Key: getters.Static("$record")}),
//		},
//	}
type View struct {
	// PageName represents the unique identifier string referencing the page component.
	PageName string
	// PageLookup represents the resolver function mapping page keys to [components.PageInterface] objects.
	PageLookup func(name string) (components.PageInterface, bool)
	// Layers represents the collection of middleware layers wrapping page views.
	Layers []registry.Pair[string, Layer]

	mu            sync.RWMutex
	cachedHandler http.Handler
}

// GetHandler compiles the View's middleware layers and rendering handlers into a single nested [http.Handler] flow.
func (v *View) GetHandler() http.Handler {
	v.mu.RLock()
	if v.cachedHandler != nil {
		h := v.cachedHandler
		v.mu.RUnlock()
		return h
	}
	v.mu.RUnlock()

	v.mu.Lock()
	defer v.mu.Unlock()
	if v.cachedHandler != nil {
		return v.cachedHandler
	}

	var handler http.Handler = http.HandlerFunc(v.RenderPage)
	for i := len(v.Layers) - 1; i >= 0; i-- {
		handler = v.Layers[i].Value.Next(*v, handler)
	}
	v.cachedHandler = handler
	return handler
}

// ServeHTTP satisfies the standard http.Handler interface, executing View handlers.
func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.GetHandler().ServeHTTP(w, r)
}

// GetPage resolves and returns the configured page component of this view.
func (v *View) GetPage() (components.PageInterface, bool) {
	return v.PageLookup(v.PageName)
}

// RenderPage renders the resolved page component writing output HTML templates directly.
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

// isBenignResponseWriteError returns true if the error represents a benign connection termination event.
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

// ParseForm traverses the page structure, locates the primary [components.FormInterface] child,
// parses parameter inputs, and yields parsed values and validation errors.
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

// WithLayer appends a new middleware layer block to the View execution stack.
func (v *View) WithLayer(name string, layer Layer) *View {
	v.mu.Lock()
	defer v.mu.Unlock()
	// Append layer; keys are labels only and are not required to be unique.
	v.Layers = append(v.Layers, registry.Pair[string, Layer]{Key: name, Value: layer})
	v.cachedHandler = nil
	return v
}

// InsertLayerBefore inserts a middleware layer with the given name immediately before the first layer matching beforeName.
// If the target layer is missing, it appends it to the end of the stack.
func (v *View) InsertLayerBefore(beforeName, name string, layer Layer) *View {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cachedHandler = nil
	p := registry.Pair[string, Layer]{Key: name, Value: layer}
	for i, mw := range v.Layers {
		if mw.Key == beforeName {
			v.Layers = append(v.Layers[:i], append([]registry.Pair[string, Layer]{p}, v.Layers[i:]...)...)
			return v
		}
	}
	// Fallback: behave like WithLayer when beforeName is not found.
	v.Layers = append(v.Layers, p)
	return v
}

// InsertLayerAfter inserts a middleware layer with the given name immediately after the first layer matching afterName.
// If the target layer is missing, it appends it to the end of the stack.
func (v *View) InsertLayerAfter(afterName, name string, layer Layer) *View {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cachedHandler = nil
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
	v.Layers = append(v.Layers, p)
	return v
}

// WithLayers appends multiple middleware layer blocks to the View execution stack.
func (v *View) WithLayers(layers ...registry.Pair[string, Layer]) *View {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cachedHandler = nil
	for _, layer := range layers {
		v.Layers = append(v.Layers, registry.Pair[string, Layer]{Key: layer.Key, Value: layer.Value})
	}
	return v
}

// PatchLayers applies function modifications to multiple middleware layers by matching keys.
func (v *View) PatchLayers(layers ...registry.Pair[string, func(Layer) Layer]) *View {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cachedHandler = nil
	for _, layer := range layers {
		for i, mw := range v.Layers {
			if mw.Key == layer.Key {
				v.Layers[i].Value = layer.Value(mw.Value)
			}
		}
	}
	return v
}

// PatchLayer applies a function patcher to the first middleware layer matching the name key.
func (v *View) PatchLayer(name string, patcher func(Layer) Layer) *View {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cachedHandler = nil
	for i, mw := range v.Layers {
		if mw.Key == name {
			v.Layers[i].Value = patcher(mw.Value)
		}
	}
	return v
}
