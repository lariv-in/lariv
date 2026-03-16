package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Environment struct {
	Page
	Label   string
	Key     getters.Getter[string]   // key used in the environment cookie map
	Options getters.Getter[[]string] // option values
	Default getters.Getter[string]
	Classes string
}

func (e Environment) GetKey() string {
	return e.Page.Key
}

func (e Environment) GetRoles() []string {
	return e.Roles
}

func (e Environment) Build(ctx context.Context) Node {
	key := ""
	if e.Key != nil {
		k, err := e.Key(ctx)
		if err != nil {
			slog.Error("Environment Key getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.GetterStatic(err)}.Build(ctx)
		}
		key = k
	}
	options := []string{}
	if e.Options != nil {
		opts, err := e.Options(ctx)
		if err != nil {
			slog.Error("Environment Options getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.GetterStatic(err)}.Build(ctx)
		}
		options = opts
	}

	// Read current value from $environment context
	var current string
	if envMap, ok := ctx.Value("$environment").(map[string]string); ok {
		var isEnvironmentPresent bool
		current, isEnvironmentPresent = envMap[key]
		if !isEnvironmentPresent && e.Default != nil {
			def, err := e.Default(ctx)
			if err != nil {
				slog.Error("Environment Default getter failed", "error", err, "key", e.Key)
				return ContainerError{Error: getters.GetterStatic(err)}.Build(ctx)
			}
			current = def
		}
	}

	optionNodes := []Node{
		Option(Value(""), If(current == "", Attr("selected", "")), Text("—")),
	}
	for _, opt := range options {
		optionNodes = append(optionNodes,
			Option(Value(opt), If(current == opt, Attr("selected", "")), Text(opt)),
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
