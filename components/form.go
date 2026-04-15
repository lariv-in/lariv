package components

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/lariv-in/lago/getters"
	"maragu.dev/gomponents"
	"maragu.dev/gomponents/html"
)

type FormComponent[T any] struct {
	Page
	Getter         getters.Getter[T]
	ChildrenInput  []PageInterface
	ChildrenAction []PageInterface
	Classes        string
	Title          string
	Subtitle       string
	Attr           getters.Getter[gomponents.Node]
}

type FormInterface interface {
	PageInterface
	ParseForm(r *http.Request) (map[string]any, map[string]error, error)
}

func (e FormComponent[T]) Build(ctx context.Context) gomponents.Node {
	// If a Getter is set, resolve the object and pass it to children via $in
	childCtx := ctx
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FormComponent getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		if v := reflect.ValueOf(value); v.IsValid() && !v.IsZero() {
			objMap := getters.MapFromStruct(value)
			if currentValues, ok := ctx.Value(getters.ContextKeyIn).(map[string]any); ok && len(currentValues) > 0 {
				for key, value := range currentValues {
					objMap[key] = value
				}
			}
			childCtx = context.WithValue(ctx, getters.ContextKeyIn, objMap)
		}
	}

	inputGroup := gomponents.Group{}
	for _, child := range e.ChildrenInput {
		inputGroup = append(inputGroup, Render(child, childCtx))
	}
	submitGroup := gomponents.Group{html.Class("my-2")}
	for _, child := range e.ChildrenAction {
		submitGroup = append(submitGroup, Render(child, childCtx))
	}

	var headerNodes []gomponents.Node
	if e.Title != "" {
		headerNodes = append(headerNodes, html.Div(html.Class("text-xl font-semibold"), gomponents.Text(e.Title)))
	}
	if e.Subtitle != "" {
		headerNodes = append(headerNodes, html.Div(html.Class("text-sm text-gray-500"), gomponents.Text(e.Subtitle)))
	}

	var formErrorNode gomponents.Node
	if errMap, ok := childCtx.Value(getters.ContextKeyError).(map[string]error); ok {
		if formErr := errMap["_form"]; formErr != nil {
			formErrorNode = html.Span(html.Class("text-sm text-error"), gomponents.Text(formErr.Error()))
		}
	} else if errorMap, ok := childCtx.Value(getters.ContextKeyError).(map[string]any); ok {
		if formErr, exists := errorMap["_form"]; exists && formErr != nil {
			if err, ok := formErr.(error); ok {
				formErrorNode = html.Span(html.Class("text-sm text-error"), gomponents.Text(err.Error()))
			}
		}
	}

	formNodes := []gomponents.Node{
		html.Class(fmt.Sprintf("flex flex-col gap-2 %s", e.Classes)),
	}
	if e.Attr != nil {
		extra, err := e.Attr(childCtx)
		if err != nil {
			slog.Error("FormComponent Attr getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(ctx)
		}
		if extra != nil {
			formNodes = append(formNodes, extra)
		}
	}

	enctype := ""
	for _, input := range FindInputs(&e) {
		if _, ok := input.(MultipartInputInterface); ok {
			enctype = "multipart/form-data"
			break
		}
	}
	if enctype != "" {
		formNodes = append(formNodes, gomponents.Attr("enctype", enctype))
	}

	formNodes = append(formNodes,
		html.Div(headerNodes...),
		html.Div(inputGroup...),
		formErrorNode,
		html.Div(submitGroup...),
	)
	return html.Form(formNodes...)
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
	end := min(offset+nInput, len(children))
	e.ChildrenInput = children[offset:end]
	offset = end
	if offset >= len(children) {
		return
	}
	nAction := len(e.ChildrenAction)
	end = min(offset+nAction, len(children))
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
			if multipartInput, ok := input.(MultipartInputInterface); ok {
				fieldVal, fieldErr := multipartInput.ParseMultipart(r.MultipartForm.File[name], r.Context())
				inputValues[name] = fieldVal
				if fieldErr != nil {
					inputErrors[name] = fieldErr
				}

			} else {
				fieldVal, fieldErr := input.Parse(r.MultipartForm.Value[name], r.Context())
				inputValues[name] = fieldVal
				if fieldErr != nil {
					inputErrors[name] = fieldErr
				}

			}
		} else {
			fieldVal, fieldErr := input.Parse(r.Form[name], r.Context())
			inputValues[name] = fieldVal
			if fieldErr != nil {
				inputErrors[name] = fieldErr
			}

		}
	}

	return inputValues, inputErrors, nil
}
