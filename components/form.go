package components

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FormComponent struct {
	Getter         Getter
	Url            Getter
	Method         string
	ChildrenInput  []PageInterface
	ChildrenAction []PageInterface
	Classes        string
}

func (e FormComponent) Build(ctx context.Context) Node {
	inputGroup := Group{}
	for _, child := range e.ChildrenInput {
		inputGroup = append(inputGroup, child.Build(ctx))
	}
	submitGroup := Group{}
	for _, child := range e.ChildrenAction {
		submitGroup = append(submitGroup, child.Build(ctx))
	}
	urlString := fmt.Sprintf("%s", IfOrGetter(e.Url, ctx, ""))
	return Form(
		Class(fmt.Sprintf("flex flex-col %s", e.Classes)),
		If(e.Method != "", Method(e.Method)),
		If(urlString != "", Action(urlString)),
		inputGroup,
		submitGroup)
}

func (e FormComponent) GetChildren() []PageInterface {
	return append(e.ChildrenInput, e.ChildrenAction...)
}

// Calls ParseMultipartForm or ParseForm based on Content-Type and for each Child under it that implements InputIterface, calls its clean method and stores that value in the map, and stores the error in the error map
func (e FormComponent) ParseForm(r *http.Request) (map[string]any, map[string]error, error) {
	var err error
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		err = r.ParseMultipartForm(4 * 1024 * 1024)
	} else {
		err = r.ParseForm()
	}

	if err != nil {
		return nil, nil, err
	}

	inputValues, inputErrors := map[string]any{}, map[string]error{}

	inputs := FindInputs(e)

	for _, input := range inputs {
		name := input.GetName()
		inputValues[name], inputErrors[name] = input.Parse(r.FormValue(name))
	}

	return inputValues, inputErrors, nil
}
