package components

import (
	"context"
	"fmt"
	"strconv"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputTernary struct {
	Page
	Label      string
	Name       string
	Getter     getters.Getter[bool]
	TrueLabel  string
	FalseLabel string
	NoneLabel  string
	Classes    string
}

func (e InputTernary) Build(ctx context.Context) Node {
	value, err := getters.IfOrGetter(e.Getter, ctx, false)

	trueLabel := e.TrueLabel
	if trueLabel == "" {
		trueLabel = "Yes"
	}
	falseLabel := e.FalseLabel
	if falseLabel == "" {
		falseLabel = "No"
	}
	noneLabel := e.NoneLabel
	if noneLabel == "" {
		noneLabel = "Not Set"
	}

	noneSelected := ""
	trueSelected := ""
	falseSelected := ""
	if err != nil {
		noneSelected = "selected"
	} else if value {
		trueSelected = "selected"
	} else {
		falseSelected = "selected"
	}

	return Div(Class(fmt.Sprintf("my-1 %s", e.Classes)),
		Label(Class("label text-sm font-bold"), Text(e.Label)),
		Select(Name(e.Name), Class("select select-bordered w-full"),
			Option(Value(""), If(noneSelected != "", Attr("selected", "")), Text(noneLabel)),
			Option(Value("True"), If(trueSelected != "", Attr("selected", "")), Text(trueLabel)),
			Option(Value("False"), If(falseSelected != "", Attr("selected", "")), Text(falseLabel)),
		),
	)
}

func (e InputTernary) Parse(v any, _ context.Context) (any, error) {
	vals, _ := v.([]string)
	if len(vals) == 0 || vals[0] == "" {
		return nil, nil
	}
	b, err := strconv.ParseBool(vals[0])
	if err != nil {
		return nil, nil
	}
	return b, nil
}

func (e InputTernary) GetName() string {
	return e.Name
}
