package components

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/lariv-in/getters"
	"github.com/lariv-in/views"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FormComponent struct {
	Page
	Getter         getters.Getter
	Url            getters.Getter
	Method         string
	ChildrenInput  []PageInterface
	ChildrenAction []PageInterface
	Classes        string
	Title          string
	Subtitle       string
}

func (e FormComponent) Build(ctx context.Context) Node {
	// If a Getter is set, resolve the object and pass it to children via $in
	childCtx := ctx
	if e.Getter != nil {
		value := e.Getter(ctx)
		if value != nil {
			objMap := getters.MapFromStruct(value)
			childCtx = context.WithValue(ctx, "$in", objMap)
		}
	}

	inputGroup := Group{}
	for _, child := range e.ChildrenInput {
		inputGroup = append(inputGroup, Render(child, childCtx))
	}
	submitGroup := Group{}
	for _, child := range e.ChildrenAction {
		submitGroup = append(submitGroup, Render(child, childCtx))
	}
	urlString := fmt.Sprintf("%s", getters.IfOrGetter(e.Url, childCtx, ""))

	var headerNodes []Node
	if e.Title != "" {
		headerNodes = append(headerNodes, Div(Class("text-xl font-semibold"), Text(e.Title)))
	}
	if e.Subtitle != "" {
		headerNodes = append(headerNodes, Div(Class("text-sm text-gray-500"), Text(e.Subtitle)))
	}

	var formErrorNode Node
	fmt.Println(childCtx.Value(views.GlobalContextError))
	if errorMap, ok := childCtx.Value("$error").(map[string]any); ok {
		if formErr, exists := errorMap["_form"]; exists && formErr != nil {
			if err, ok := formErr.(error); ok {
				formErrorNode = Span(Class("text-sm text-error"), Text(err.Error()))
			}
		}
	}

	return Form(
		Class(fmt.Sprintf("flex flex-col %s", e.Classes)),
		If(e.Method != "", Method(e.Method)),
		If(urlString != "", Action(urlString)),
		Group(headerNodes),
		inputGroup,
		formErrorNode,
		submitGroup)
}

func (e FormComponent) GetChildren() []PageInterface {
	return append(e.ChildrenInput, e.ChildrenAction...)
}

// Calls ParseMultipartForm or ParseForm based on Content-Type and for each Child under it that implements InputIterface, calls its clean method and stores that value in the map, and stores the error in the error map
func (e FormComponent) ParseForm(r *http.Request) (map[string]any, map[string]error, error) {
	var err error
	isMultipart := false
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		err = r.ParseMultipartForm(4 * 1024 * 1024)
		isMultipart = true
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
		if isMultipart {
			inputValues[name], inputErrors[name] = input.Parse(r.MultipartForm.Value[name], r.Context())
		} else {
			inputValues[name], inputErrors[name] = input.Parse(r.Form[name], r.Context())
		}
	}

	return inputValues, inputErrors, nil
}
