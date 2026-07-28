package components

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lariv/getters"
)

const multiStepFormErrorFieldPrefix = "$error."

// MultiStepForm represents a sequential wizard or multi-stage form container component.
// It bundles multiple standard [FormInterface] stages together, displaying one active stage at a time and automatically
// carrying values and validation errors from inactive stages inside hidden inputs.
// It automatically renders a navigation ribbon tracking steps, highlighting stages that contain validation errors.
//
// Active stages that are [FormComponent] receive action/ribbon/hidden markup via context keys (no HTML string surgery).
//
// Use Cases:
//   - Constructing user onboarding funnels, multi-page checkout flows, or complex configuration wizards.
//
// Example:
//
//	&components.MultiStepForm{
//	    Stages: []components.FormInterface{
//	        &components.FormComponent[PersonalInfo]{...},
//	        &components.FormComponent[PaymentDetails]{...},
//	    },
//	    Stage: getters.Static(0),
//	}
type MultiStepForm struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Stages represents the slice of FormInterface sub-forms representing individual steps.
	Stages []FormInterface
	// Stage is the dynamic function retrieving the index of the currently active form step.
	Stage getters.Getter[int]
	// Values is the dynamic function retrieving parameters submitted in previous form steps.
	Values getters.Getter[map[string]any]
	// MultiStageURL is the dynamic function retrieving the form endpoint target URL (action attribute).
	MultiStageURL getters.Getter[string]
}

var (
	_ FormInterface          = MultiStepForm{}
	_ ParentInterface        = MultiStepForm{}
	_ MutableParentInterface = (*MultiStepForm)(nil)
)

type multiStepRibbonButton struct {
	Kind    string // "current", "submit", or "disabled"
	Label   string
	Value   string
	Classes string
}

// Build compiles the active Form stage page layout, injecting hidden values and progress step ribbons.
func (e MultiStepForm) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	if len(e.Stages) == 0 {
		return ContainerError{Error: getters.Static(fmt.Errorf("MultiStepForm: no stages configured"))}.Build(cat, ctx, w)
	}

	stageIdx := e.resolveStage(ctx)
	values := e.resolveValues(ctx)
	errors := e.resolveErrors(ctx)
	actionURL := e.resolveMultiStageURL(ctx)

	hiddenHTML, err := e.hiddenFieldsHTML(cat, ctx, stageIdx, values, errors)
	if err != nil {
		slog.Error("MultiStepForm hidden render failed", "error", err, "key", e.Key, "stage", stageIdx)
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}

	ribbonHTML, err := e.ribbonHTML(stageIdx, errors)
	if err != nil {
		slog.Error("MultiStepForm ribbon render failed", "error", err, "key", e.Key, "stage", stageIdx)
		return ContainerError{Error: getters.Static(err)}.Build(cat, ctx, w)
	}

	stageCtx := ctx
	if actionURL != "" {
		stageCtx = context.WithValue(stageCtx, multiStepFormActionKey, actionURL)
	}
	stageCtx = context.WithValue(stageCtx, multiStepFormPrefixKey, template.HTML(ribbonHTML))
	stageCtx = context.WithValue(stageCtx, multiStepFormSuffixKey, template.HTML(hiddenHTML))

	return Render(e.Stages[stageIdx], cat, stageCtx, w)
}

// ParseForm parses submitted form parameters for the active stage using multipart/url encoding.
func (e MultiStepForm) ParseForm(r *http.Request) (map[string]any, map[string]error, error) {
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

	stageIdx := e.ParseStage(r)
	requestNames := requestFieldNames(r, isMultipart)
	inputs := e.parseInputsForStage(stageIdx, requestNames)

	values := map[string]any{}
	fieldErrors := map[string]error{}
	for _, input := range inputs {
		name := input.GetName()
		if isMultipart {
			if multipartInput, ok := input.(MultipartInputInterface); ok {
				fieldVal, fieldErr := multipartInput.ParseMultipart(r.MultipartForm.File[name], r.Context())
				values[name] = fieldVal
				if fieldErr != nil {
					fieldErrors[name] = fieldErr
				}
				continue
			}
			fieldVal, fieldErr := input.Parse(r.MultipartForm.Value[name], r.Context())
			values[name] = fieldVal
			if fieldErr != nil {
				fieldErrors[name] = fieldErr
			}
			continue
		}

		fieldVal, fieldErr := input.Parse(r.Form[name], r.Context())
		values[name] = fieldVal
		if fieldErr != nil {
			fieldErrors[name] = fieldErr
		}
	}
	return values, fieldErrors, nil
}

// GetKey returns the unique key identifier for this MultiStepForm.
func (e MultiStepForm) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this MultiStepForm.
func (e MultiStepForm) GetRoles() []string {
	return e.Roles
}

// GetChildren returns sub-forms representing individual stages.
func (e MultiStepForm) GetChildren() []PageInterface {
	children := make([]PageInterface, 0, len(e.Stages))
	for _, stage := range e.Stages {
		children = append(children, stage)
	}
	return children
}

// SetChildren replaces individual wizard stages from page children interfaces.
func (e *MultiStepForm) SetChildren(children []PageInterface) {
	stages := make([]FormInterface, 0, len(children))
	for _, child := range children {
		form, ok := child.(FormInterface)
		if !ok {
			slog.Error("MultiStepForm child is not a form", "key", e.Key, "type", fmt.Sprintf("%T", child))
			continue
		}
		stages = append(stages, form)
	}
	e.Stages = stages
}

// StageCount returns the total number of wizard steps.
func (e MultiStepForm) StageCount() int {
	return len(e.Stages)
}

// ParseStage extracts the current stage index parameter from incoming request values.
func (e MultiStepForm) ParseStage(r *http.Request) int {
	return e.requestStage(r)
}

// ParseTargetStage extracts the desired next target stage index parameter from submitted parameters.
func (e MultiStepForm) ParseTargetStage(r *http.Request, currentStage int) int {
	target := currentStage
	raw := ""
	if r.MultipartForm != nil {
		raw = firstFormValue(r.MultipartForm.Value["$stage_target"])
	}
	if raw == "" && r.Form != nil {
		raw = firstFormValue(r.Form["$stage_target"])
	}
	if raw == "" {
		if currentStage < len(e.Stages)-1 {
			return currentStage + 1
		}
		return currentStage
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		slog.Error("MultiStepForm target stage parse failed", "error", err, "key", e.Key, "raw", raw)
		return target
	}
	return clampStageIndex(parsed, len(e.Stages))
}

func (e MultiStepForm) resolveStage(ctx context.Context) int {
	stage := 0
	if e.Stage != nil {
		resolved, err := e.Stage(ctx)
		if err != nil {
			slog.Error("MultiStepForm stage getter failed", "error", err, "key", e.Key)
		} else {
			stage = resolved
		}
	} else if resolved, ok := ctx.Value("$stage").(int); ok {
		stage = resolved
	}
	return clampStageIndex(stage, len(e.Stages))
}

func (e MultiStepForm) requestStage(r *http.Request) int {
	stage := 0
	raw := ""
	if r.MultipartForm != nil {
		raw = firstFormValue(r.MultipartForm.Value["$stage"])
	}
	if raw == "" && r.Form != nil {
		raw = firstFormValue(r.Form["$stage"])
	}
	if raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			slog.Error("MultiStepForm stage parse failed", "error", err, "key", e.Key, "raw", raw)
		} else {
			stage = parsed
		}
	}
	return clampStageIndex(stage, len(e.Stages))
}

func (e MultiStepForm) resolveValues(ctx context.Context) map[string]any {
	if e.Values != nil {
		values, err := e.Values(ctx)
		if err != nil {
			slog.Error("MultiStepForm values getter failed", "error", err, "key", e.Key)
			return map[string]any{}
		}
		return cloneAnyMap(values)
	}

	switch value := ctx.Value(getters.ContextKeyIn).(type) {
	case map[string]any:
		return cloneAnyMap(value)
	case nil:
		return map[string]any{}
	default:
		return getters.MapFromStruct(value)
	}
}

func (e MultiStepForm) resolveErrors(ctx context.Context) map[string]error {
	switch value := ctx.Value(getters.ContextKeyError).(type) {
	case map[string]error:
		return cloneErrorMap(value)
	case map[string]any:
		out := map[string]error{}
		for key, item := range value {
			err, ok := item.(error)
			if !ok || err == nil {
				continue
			}
			out[key] = err
		}
		return out
	case nil:
		return map[string]error{}
	default:
		return map[string]error{}
	}
}

func (e MultiStepForm) resolveMultiStageURL(ctx context.Context) string {
	if e.MultiStageURL == nil {
		return ""
	}
	url, err := e.MultiStageURL(ctx)
	if err != nil {
		slog.Error("MultiStepForm url getter failed", "error", err, "key", e.Key)
		return ""
	}
	return url
}

func (e MultiStepForm) hiddenFieldsHTML(cat Catalog, ctx context.Context, stageIdx int, values map[string]any, errors map[string]error) (string, error) {
	var out strings.Builder
	if err := Execute(&out, "multi_step_form_hidden_stage", strconv.Itoa(stageIdx)); err != nil {
		return "", err
	}

	activeNames := e.stageInputNames(stageIdx)
	seen := map[string]struct{}{}
	for i, stage := range e.Stages {
		if i == stageIdx {
			continue
		}
		for _, input := range formInputs(stage) {
			name := input.GetName()
			if name == "" {
				continue
			}
			if _, ok := activeNames[name]; ok {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			value, ok := values[name]
			if !ok || isNilAny(value) {
				continue
			}
			html, renderable := hiddenCarryHTML(input, value, cat, ctx)
			if !renderable {
				slog.Error("MultiStepForm hidden carry unsupported", "key", e.Key, "input", name, "type", fmt.Sprintf("%T", value))
				continue
			}
			out.WriteString(string(html))
			seen[name] = struct{}{}
		}
	}
	errHTML, err := hiddenErrorHTML(errors)
	if err != nil {
		return "", err
	}
	out.WriteString(errHTML)
	return out.String(), nil
}

func (e MultiStepForm) ribbonHTML(stageIdx int, errors map[string]error) (string, error) {
	buttons := e.ribbonButtons(stageIdx, e.stageErrors(errors))
	var out bytes.Buffer
	if err := Execute(&out, "multi_step_form_ribbon", struct {
		Buttons []multiStepRibbonButton
	}{Buttons: buttons}); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (e MultiStepForm) ribbonButtons(stageIdx int, stageErrors map[int]struct{}) []multiStepRibbonButton {
	buttons := make([]multiStepRibbonButton, 0, len(e.Stages))
	for i := range e.Stages {
		label := fmt.Sprintf("Step %d", i+1)
		classes := e.ribbonButtonClasses(i, stageIdx, stageErrors)
		switch {
		case i == stageIdx:
			buttons = append(buttons, multiStepRibbonButton{Kind: "current", Label: label, Classes: classes})
		case i < stageIdx:
			buttons = append(buttons, multiStepRibbonButton{
				Kind: "submit", Label: label, Value: strconv.Itoa(i), Classes: classes,
			})
		default:
			buttons = append(buttons, multiStepRibbonButton{Kind: "disabled", Label: label, Classes: classes})
		}
	}
	return buttons
}

func (e MultiStepForm) stageErrors(errors map[string]error) map[int]struct{} {
	if len(errors) == 0 {
		return map[int]struct{}{}
	}
	stageErrors := map[int]struct{}{}
	for key, err := range errors {
		if key == "" || err == nil {
			continue
		}
		if key == "_form" {
			for i := range e.Stages {
				stageErrors[i] = struct{}{}
			}
			continue
		}
		for i := range e.Stages {
			if _, ok := e.stageInputNames(i)[key]; ok {
				stageErrors[i] = struct{}{}
			}
		}
	}
	return stageErrors
}

func (e MultiStepForm) ribbonButtonClasses(stepIdx, stageIdx int, stageErrors map[int]struct{}) string {
	classes := []string{"btn", "btn-sm"}
	switch {
	case stepIdx == stageIdx:
		classes = append(classes, "btn-primary")
	case stepIdx < stageIdx:
		classes = append(classes, "btn-outline")
	default:
		classes = append(classes, "btn-disabled")
	}
	if _, ok := stageErrors[stepIdx]; ok {
		classes = append(classes, "border-2", "border-error")
	}
	return strings.Join(classes, " ")
}

func (e MultiStepForm) parseInputsForStage(stageIdx int, requestNames map[string]struct{}) []InputInterface {
	activeInputs := formInputs(e.Stages[stageIdx])
	activeNames := map[string]struct{}{}
	result := make([]InputInterface, 0, len(activeInputs))
	for _, input := range activeInputs {
		name := input.GetName()
		activeNames[name] = struct{}{}
		result = append(result, input)
	}

	seen := cloneStringSet(activeNames)
	for _, stage := range e.Stages {
		for _, input := range formInputs(stage) {
			name := input.GetName()
			if _, ok := seen[name]; ok {
				continue
			}
			if _, ok := requestNames[name]; !ok {
				continue
			}
			result = append(result, input)
			seen[name] = struct{}{}
		}
	}
	return result
}

func (e MultiStepForm) stageInputNames(stageIdx int) map[string]struct{} {
	names := map[string]struct{}{}
	for _, input := range formInputs(e.Stages[stageIdx]) {
		names[input.GetName()] = struct{}{}
	}
	return names
}

// StageInputNames returns the set of input field parameter names registered in a given step index.
func (e MultiStepForm) StageInputNames(stageIdx int) map[string]struct{} {
	return cloneStringSet(e.stageInputNames(stageIdx))
}

func formInputs(form FormInterface) []InputInterface {
	parent, ok := form.(ParentInterface)
	if !ok {
		return nil
	}
	return FindInputs(parent)
}

func requestFieldNames(r *http.Request, isMultipart bool) map[string]struct{} {
	names := map[string]struct{}{}
	if isMultipart && r.MultipartForm != nil {
		for name := range r.MultipartForm.Value {
			names[name] = struct{}{}
		}
		for name := range r.MultipartForm.File {
			names[name] = struct{}{}
		}
		return names
	}
	for name := range r.Form {
		names[name] = struct{}{}
	}
	return names
}

// ParseMultiStepErrors parses validation errors encoded with multiStepFormErrorFieldPrefix from the request payload.
func ParseMultiStepErrors(r *http.Request) map[string]error {
	errors := map[string]error{}
	appendErrors := func(values map[string][]string) {
		for name, rawValues := range values {
			key, ok := strings.CutPrefix(name, multiStepFormErrorFieldPrefix)
			if !ok || key == "" {
				continue
			}
			message := firstFormValue(rawValues)
			if message == "" {
				continue
			}
			errors[key] = fmt.Errorf("%s", message)
		}
	}
	if r.MultipartForm != nil {
		appendErrors(r.MultipartForm.Value)
	}
	if r.Form != nil {
		appendErrors(r.Form)
	}
	return errors
}

func hiddenErrorHTML(errors map[string]error) (string, error) {
	if len(errors) == 0 {
		return "", nil
	}
	keys := make([]string, 0, len(errors))
	for key, err := range errors {
		if key == "" || err == nil {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var out strings.Builder
	for _, key := range keys {
		if err := Execute(&out, "multi_step_form_hidden_input", struct {
			Name  string
			Value string
		}{Name: multiStepFormErrorFieldPrefix + key, Value: errors[key].Error()}); err != nil {
			return "", err
		}
	}
	return out.String(), nil
}

func hiddenCarryHTML(input InputInterface, value any, cat Catalog, ctx context.Context) (template.HTML, bool) {
	switch typed := input.(type) {
	case InputCheckbox:
		v, ok := value.(bool)
		if !ok {
			return "", false
		}
		typed.Hidden = true
		typed.Getter = getters.Static(v)
		return pageHTMLFragment(typed, cat, ctx)
	case *InputCheckbox:
		v, ok := value.(bool)
		if !ok {
			return "", false
		}
		clone := *typed
		clone.Hidden = true
		clone.Getter = getters.Static(v)
		return pageHTMLFragment(clone, cat, ctx)
	case InputDate:
		t, ok := timeValue(value)
		if !ok {
			return "", false
		}
		typed.Hidden = true
		typed.Getter = getters.Static(t)
		return pageHTMLFragment(typed, cat, ctx)
	case *InputDate:
		t, ok := timeValue(value)
		if !ok {
			return "", false
		}
		clone := *typed
		clone.Hidden = true
		clone.Getter = getters.Static(t)
		return pageHTMLFragment(clone, cat, ctx)
	case InputTime:
		t, ok := timeValue(value)
		if !ok {
			return "", false
		}
		typed.Hidden = true
		typed.Getter = getters.Static(t)
		return pageHTMLFragment(typed, cat, ctx)
	case *InputTime:
		t, ok := timeValue(value)
		if !ok {
			return "", false
		}
		clone := *typed
		clone.Hidden = true
		clone.Getter = getters.Static(t)
		return pageHTMLFragment(clone, cat, ctx)
	case InputDatetime:
		t, ok := timeValue(value)
		if !ok {
			return "", false
		}
		typed.Hidden = true
		typed.Getter = getters.Static(t)
		return pageHTMLFragment(typed, cat, ctx)
	case *InputDatetime:
		t, ok := timeValue(value)
		if !ok {
			return "", false
		}
		clone := *typed
		clone.Hidden = true
		clone.Getter = getters.Static(t)
		return pageHTMLFragment(clone, cat, ctx)
	default:
		return genericHiddenCarryHTML(input.GetName(), value)
	}
}

func pageHTMLFragment(p PageInterface, cat Catalog, ctx context.Context) (template.HTML, bool) {
	html, err := RenderHTML(p, cat, ctx)
	if err != nil {
		slog.Error("hiddenCarryHTML render failed", "error", err, "key", p.GetKey())
		return "", false
	}
	return html, true
}

func genericHiddenCarryHTML(name string, value any) (template.HTML, bool) {
	var out bytes.Buffer
	switch typed := value.(type) {
	case AssociationIDs:
		vals := make([]string, 0, len(typed.IDs))
		for _, id := range typed.IDs {
			vals = append(vals, strconv.FormatUint(uint64(id), 10))
		}
		if err := Execute(&out, "multi_step_form_hidden_multi", struct {
			Name   string
			Values []string
		}{Name: name, Values: vals}); err != nil {
			return "", false
		}
		return template.HTML(out.String()), true
	case []string:
		if err := Execute(&out, "multi_step_form_hidden_multi", struct {
			Name   string
			Values []string
		}{Name: name, Values: typed}); err != nil {
			return "", false
		}
		return template.HTML(out.String()), true
	case []uint:
		vals := make([]string, 0, len(typed))
		for _, item := range typed {
			vals = append(vals, strconv.FormatUint(uint64(item), 10))
		}
		if err := Execute(&out, "multi_step_form_hidden_multi", struct {
			Name   string
			Values []string
		}{Name: name, Values: vals}); err != nil {
			return "", false
		}
		return template.HTML(out.String()), true
	default:
		if scalar, ok := scalarHiddenValue(value); ok {
			if err := Execute(&out, "multi_step_form_hidden_input", struct {
				Name  string
				Value string
			}{Name: name, Value: scalar}); err != nil {
				return "", false
			}
			return template.HTML(out.String()), true
		}
	}
	return "", false
}

func scalarHiddenValue(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case bool:
		return strconv.FormatBool(typed), true
	case time.Time:
		return typed.Format(time.RFC3339Nano), true
	case fmt.Stringer:
		return typed.String(), true
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return "", false
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), true
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), true
	}
	return "", false
}

func timeValue(value any) (time.Time, bool) {
	switch typed := value.(type) {
	case time.Time:
		return typed, true
	case *time.Time:
		if typed == nil {
			return time.Time{}, false
		}
		return *typed, true
	default:
		return time.Time{}, false
	}
}

func clampStageIndex(stage, total int) int {
	if total <= 0 {
		return 0
	}
	if stage < 0 {
		return 0
	}
	if stage >= total {
		return total - 1
	}
	return stage
}

func firstFormValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func cloneAnyMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneErrorMap(in map[string]error) map[string]error {
	if in == nil {
		return map[string]error{}
	}
	out := make(map[string]error, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneStringSet(in map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for k := range in {
		out[k] = struct{}{}
	}
	return out
}

func isNilAny(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	}
	return false
}
