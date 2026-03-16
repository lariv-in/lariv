package components

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/lariv-in/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type FormComponent[T any] struct {
	Page
	Getter         getters.Getter[T]
	Url            getters.Getter[string]
	Method         string
	ChildrenInput  []PageInterface
	ChildrenAction []PageInterface
	Classes        string
	Title          string
	Subtitle       string
}

type FormInterface interface {
	PageInterface
	ParseForm(r *http.Request) (map[string]any, map[string]error, error)
}

func (e FormComponent[T]) Build(ctx context.Context) Node {
	// If a Getter is set, resolve the object and pass it to children via $in
	childCtx := ctx
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FormComponent getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.GetterStatic(err)}.Build(ctx)
		}
		if v := reflect.ValueOf(value); v.IsValid() && !v.IsZero() {
			objMap := getters.MapFromStruct(value)
			childCtx = context.WithValue(ctx, getters.ContextKeyIn, objMap)
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
	urlString := ""
	if e.Url != nil {
		u, err := e.Url(childCtx)
		if err != nil {
			slog.Error("FormComponent Url getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.GetterStatic(err)}.Build(ctx)
		}
		urlString = u
	}

	var headerNodes []Node
	if e.Title != "" {
		headerNodes = append(headerNodes, Div(Class("text-xl font-semibold"), Text(e.Title)))
	}
	if e.Subtitle != "" {
		headerNodes = append(headerNodes, Div(Class("text-sm text-gray-500"), Text(e.Subtitle)))
	}

	var formErrorNode Node
	if errorMap, ok := childCtx.Value(getters.ContextKeyError).(map[string]any); ok {
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

func (e FormComponent[T]) GetKey() string {
	return e.Key
}

func (e FormComponent[T]) GetRoles() []string {
	return e.Roles
}

func (e FormComponent[T]) GetChildren() []PageInterface {
	return append(e.ChildrenInput, e.ChildrenAction...)
}

func (e *FormComponent[T]) SetChildren(children []PageInterface) {
	offset := 0
	nInput := len(e.ChildrenInput)
	end := offset + nInput
	if end > len(children) {
		end = len(children)
	}
	e.ChildrenInput = children[offset:end]
	offset = end
	if offset >= len(children) {
		return
	}
	nAction := len(e.ChildrenAction)
	end = offset + nAction
	if end > len(children) {
		end = len(children)
	}
	e.ChildrenAction = children[offset:end]
	offset = end
	if offset < len(children) {
		e.ChildrenAction = append(e.ChildrenAction, children[offset:]...)
	}
}

// Calls ParseMultipartForm or ParseForm based on Content-Type and for each Child under it that implements InputIterface, calls its clean method and stores that value in the map, and stores the error in the error map
func (e FormComponent[T]) ParseForm(r *http.Request) (map[string]any, map[string]error, error) {
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

	inputs := FindInputs(&e)

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
