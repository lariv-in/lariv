package components

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Environment[T comparable] struct {
	Page
	Label   string
	Key     getters.Getter[string] // key used in the environment cookie map
	Options getters.Getter[[]registry.Pair[T, string]]
	Default getters.Getter[T]
	Classes string
}

func (e Environment[T]) GetKey() string {
	return e.Page.Key
}

func (e Environment[T]) GetRoles() []string {
	return e.Roles
}

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
		optionNodes = append(optionNodes,
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

	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		If(e.Label != "", Label(Class("label text-sm font-bold"), Text(e.Label))),
		Select(
			Name(key),
			Class("select select-bordered w-full"),
			Attr("onchange", onChange),
			Group(optionNodes),
		),
	)
}
