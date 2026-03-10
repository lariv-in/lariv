package components

import (
	"context"
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type InputTernary struct {
	Label      string
	Name       string
	Getter     Getter
	TrueLabel  string
	FalseLabel string
	NoneLabel  string
	Classes    string
}

func (e InputTernary) Build(ctx context.Context) Node {
	value := IfOrGetter(e.Getter, ctx, nil)

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
	if b, ok := value.(bool); ok {
		if b {
			trueSelected = "selected"
		} else {
			falseSelected = "selected"
		}
	} else {
		noneSelected = "selected"
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

func (e InputTernary) Parse(v string) (any, error) {
	switch v {
	case "True", "true", "1":
		return true, nil
	case "False", "false", "0":
		return false, nil
	default:
		return nil, nil
	}
}

func (e InputTernary) GetName() string {
	return e.Name
}
