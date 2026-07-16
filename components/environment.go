package components

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Environment represents a dropdown selector that controls settings stored in a client-side "environment" cookie map.
// When the selected value changes, the component updates the client cookie map and triggers an HTMX page reload,
// adjusting the server-side contextual filters (e.g., active school branch, current academic session, selected warehouse).
//
// Use Cases:
//   - Switching the active tenant or branch in multi-tenant portals.
//   - Toggling global view contexts like active academic sessions or fiscal years.
//
// Example:
//
//	&components.Environment[uint]{
//	    Label:   "Selected Branch",
//	    Key:     getters.Static("branch_id"),
//	    Options: branchOptionsGetter, // Getter[[]registry.Pair[uint, string]]
//	    Default: defaultBranchIDGetter,
//	}
type Environment[T comparable] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Label represents the human-readable description displayed above the selector.
	Label string
	// Key is a Getter resolving to the key name used to store the setting inside the environment cookie JSON.
	Key getters.Getter[string]
	// Options is a Getter resolving to a slice of pairs representing the available select options.
	Options getters.Getter[[]registry.Pair[T, string]]
	// Default is a Getter resolving to the fallback option value if no cookie value is set.
	Default getters.Getter[T]
	// Classes represents additional CSS classes applied to the outer div wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
}

// GetKey returns the unique key identifier for this Environment component.
func (e Environment[T]) GetKey() string {
	return e.Page.Key
}

// GetRoles returns the authorized roles required to view this Environment.
func (e Environment[T]) GetRoles() []string {
	return e.Roles
}

// Build compiles the Environment component into a dropdown select menu Node.
func (e Environment[T]) Build(ctx context.Context) Node {
	var zero T

	key := ""
	if e.Key != nil {
		k, err := e.Key(ctx)
		if err != nil {
			slog.Error("Environment Key getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		key = k
	}

	options := []registry.Pair[T, string]{}
	if e.Options != nil {
		opts, err := e.Options(ctx)
		if err != nil {
			slog.Error("Environment Options getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		options = opts
	}

	rawSel := ""
	inMap := false
	if envMap, ok := ctx.Value("$environment").(map[string]string); ok {
		var raw string
		raw, inMap = envMap[key]
		rawSel = strings.TrimSpace(raw)
	}
	if !inMap && e.Default != nil {
		def, err := e.Default(ctx)
		if err != nil {
			slog.Error("Environment Default getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		if any(def) != any(zero) {
			rawSel = fmt.Sprint(def)
		}
	}

	optionNodes := []Node{
		Option(Value(""), If(rawSel == "", Attr("selected", "")), Text("—")),
	}
	for _, opt := range options {
		ks := fmt.Sprint(opt.Key)
		optionNodes = append(
			optionNodes,
			Option(Value(ks), If(rawSel == ks, Attr("selected", "")), Text(opt.Value)),
		)
	}

	onChange := fmt.Sprintf(`(function(){
		var env={};
		try{
			var c=document.cookie.split('; ').find(function(r){return r.startsWith('environment=')});
			if(c) env=JSON.parse(decodeURIComponent(c.split('=').slice(1).join('=')));
		}catch(e){}
		env[%q]=this.value;
		document.cookie='environment='+encodeURIComponent(JSON.stringify(env))+'; path=/';
		htmx.ajax('GET',window.location.pathname,{target:'body',swap:'outerHTML'});
	}).call(this)`, key)

	return Div(
		Class(fmt.Sprintf("my-1 %s", e.Classes)),
		If(e.Label != "", Label(Class("label text-sm font-bold"), Text(e.Label))),
		Select(
			Name(key),
			Class("select select-bordered w-full"),
			Attr("onchange", onChange),
			Group(optionNodes),
		),
	)
}
