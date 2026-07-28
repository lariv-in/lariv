package components

import (
	"context"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/lariv-in/lariv/getters"
)

// Context keys used by [MultiStepForm] to decorate a stage [FormComponent] without string surgery.
type multiStepFormCtxKey string

const (
	multiStepFormActionKey multiStepFormCtxKey = "$multiStepFormAction"
	multiStepFormPrefixKey multiStepFormCtxKey = "$multiStepFormPrefix"
	multiStepFormSuffixKey multiStepFormCtxKey = "$multiStepFormSuffix"
)

// FormInterface defines the behavior of a component that acts as a form, supplying form parameters parser.
type FormInterface interface {
	PageInterface
	// ParseForm parses request data, returning a map of validated values, a map of field errors, or a system error.
	ParseForm(r *http.Request) (map[string]any, map[string]error, error)
}

// FormComponent represents a generic form wrapper component containing input fields and action buttons.
// It integrates automatically with input field value binding (via the Getter mapping values into context under "$in"),
// multipart/form-data detection (via MultipartInputInterface), and field/global error displays (via the error context).
//
// Use Cases:
//   - Standard data entry workflows (e.g. signup forms, login screens, user settings edits, record creation panels).
//
// Example:
//
//	&components.FormComponent[User]{
//	    Title:    "Update Profile",
//	    Getter:   currentUserGetter,
//	    ChildrenInput: []components.PageInterface{
//	        &components.InputText{Label: "Display Name", Name: "display_name"},
//	    },
//	    ChildrenAction: []components.PageInterface{
//	        &components.ButtonSubmit{Label: "Save Changes"},
//	    },
//	}
type FormComponent[T any] struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Getter is the dynamic function retrieving the object of type T used to bind default values inside inputs.
	Getter getters.Getter[T]
	// ChildrenInput represents the input components rendered inside the form container.
	ChildrenInput []PageInterface
	// ChildrenAction represents the actions/buttons rendered at the bottom of the form.
	ChildrenAction []PageInterface
	// Classes represents additional CSS classes applied to the output HTML form wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// Title is the header title text displayed at the top of the form.
	Title string
	// Subtitle is the descriptive subheading displayed beneath the form Title.
	Subtitle string
	// Attr is an optional Getter yielding additional HTML/HTMX attributes to apply to the form element.
	Attr getters.Getter[HTMLAttributes]
}

// Build compiles the FormComponent into an HTML form, rendering headings, inputs, validation errors, and actions.
func (e FormComponent[T]) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	// If a Getter is set, resolve the object and pass it to children via $in
	childCtx := ctx
	if e.Getter != nil {
		value, err := e.Getter(ctx)
		if err != nil {
			slog.Error("FormComponent getter failed", "error", err, "key", e.Key)
			return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
		}
		objMap := getters.MapFromStruct(value)
		if currentValues, ok := ctx.Value(getters.ContextKeyIn).(map[string]any); ok && len(currentValues) > 0 {
			for key, value := range currentValues {
				objMap[key] = value
			}
		}
		childCtx = context.WithValue(ctx, getters.ContextKeyIn, objMap)
	}

	inputs, err := RenderChildren(cat, childCtx, e.ChildrenInput)
	if err != nil {
		return err
	}
	actions, err := RenderChildren(cat, childCtx, e.ChildrenAction)
	if err != nil {
		return err
	}

	var formError string
	if errMap, ok := childCtx.Value(getters.ContextKeyError).(map[string]error); ok {
		if formErr := errMap["_form"]; formErr != nil {
			formError = formErr.Error()
		}
	} else if errorMap, ok := childCtx.Value(getters.ContextKeyError).(map[string]any); ok {
		if formErr, exists := errorMap["_form"]; exists && formErr != nil {
			if err, ok := formErr.(error); ok {
				formError = err.Error()
			}
		}
	}

	attrs, err := ResolveAttrs(childCtx, e.Attr)
	if err != nil {
		slog.Error("FormComponent Attr getter failed", "error", err, "key", e.Key)
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}

	enctype := ""
	for _, input := range FindInputs(&e) {
		if _, ok := input.(MultipartInputInterface); ok {
			enctype = "multipart/form-data"
			break
		}
	}

	action, _ := childCtx.Value(multiStepFormActionKey).(string)
	prefix, _ := childCtx.Value(multiStepFormPrefixKey).(template.HTML)
	suffix, _ := childCtx.Value(multiStepFormSuffixKey).(template.HTML)

	return Execute(w, "form", struct {
		Classes  string
		Enctype  string
		Action   string
		Attrs    HTMLAttributes
		Prefix   template.HTML
		Title    string
		Subtitle string
		Inputs   template.HTML
		FormError string
		Actions  template.HTML
		Suffix   template.HTML
	}{
		Classes:   e.Classes,
		Enctype:   enctype,
		Action:    action,
		Attrs:     attrs,
		Prefix:    prefix,
		Title:     e.Title,
		Subtitle:  e.Subtitle,
		Inputs:    inputs,
		FormError: formError,
		Actions:   actions,
		Suffix:    suffix,
	})
}

// GetKey returns the unique key identifier for this FormComponent.
func (e FormComponent[T]) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this FormComponent.
func (e FormComponent[T]) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the combined slice of input and action child components.
func (e FormComponent[T]) GetChildren() []PageInterface {
	return append(e.ChildrenInput, e.ChildrenAction...)
}

// SetChildren overwrites the input and action child components.
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

// ParseForm parses form parameters from the incoming http.Request, traversing children to validate and clean fields.
// Automatically differentiates between standard post values and multipart values.
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
