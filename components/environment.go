package components

import (
	"context"
	"fmt"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type Environment struct {
	Page
	Label   string
	Key     getters.Getter // key used in the environment cookie map
	Options getters.Getter // should return []string of option values
	Default getters.Getter
	Classes string
}

func (e Environment) Build(ctx context.Context) Node {
	key, _ := getters.IfOrGetter(e.Key, ctx, "").(string)
	options, _ := getters.IfOrGetter(e.Options, ctx, []string{}).([]string)

	// Read current value from $environment context
	var current string
	if envMap, ok := ctx.Value("$environment").(map[string]string); ok {
		var isEnvironmentPresent bool
		current, isEnvironmentPresent = envMap[key]
		if !isEnvironmentPresent {
			fmt.Println(e.Default(ctx))
			current = e.Default(ctx).(string)
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
