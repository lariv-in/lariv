package components

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var multiStepFormActionAttr = regexp.MustCompile(`\saction="[^"]*"`)

const multiStepFormErrorFieldPrefix = "$error."

type MultiStepForm struct {
	Page
	Stages        []FormInterface
	Stage         getters.Getter[int]
	Values        getters.Getter[map[string]any]
	MultiStageURL getters.Getter[string]
}

var _ FormInterface = MultiStepForm{}
var _ ParentInterface = MultiStepForm{}
var _ MutableParentInterface = (*MultiStepForm)(nil)

func (e MultiStepForm) Build(ctx context.Context) Node {
	if len(e.Stages) == 0 {
		return ContainerError{Error: getters.Static(fmt.Errorf("MultiStepForm: no stages configured"))}.Build(ctx)
	}

	stageIdx := e.resolveStage(ctx)
	values := e.resolveValues(ctx)
	errors := e.resolveErrors(ctx)
	actionURL := e.resolveMultiStageURL(ctx)

	stageHTML, err := renderNodeToString(Render(e.Stages[stageIdx], ctx))
	if err != nil {
		slog.Error("MultiStepForm stage render failed", "error", err, "key", e.Key, "stage", stageIdx)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}

	hiddenHTML, err := e.hiddenFieldsHTML(ctx, stageIdx, values, errors)
	if err != nil {
		slog.Error("MultiStepForm hidden render failed", "error", err, "key", e.Key, "stage", stageIdx)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}

	ribbonHTML, err := e.ribbonHTML(stageIdx, errors)
	if err != nil {
		slog.Error("MultiStepForm ribbon render failed", "error", err, "key", e.Key, "stage", stageIdx)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}

	html, err := injectIntoRenderedForm(stageHTML, actionURL, ribbonHTML, hiddenHTML)
	if err != nil {
		slog.Error("MultiStepForm form injection failed", "error", err, "key", e.Key, "stage", stageIdx)
		return ContainerError{Error: getters.Static(err)}.Build(ctx)
	}

	return Raw(html)
}

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

func (e MultiStepForm) GetKey() string {
	return e.Key
}

func (e MultiStepForm) GetRoles() []string {
	return e.Roles
}

func (e MultiStepForm) GetChildren() []PageInterface {
	children := make([]PageInterface, 0, len(e.Stages))
	for _, stage := range e.Stages {
		children = append(children, stage)
	}
	return children
}

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

func (e MultiStepForm) StageCount() int {
	return len(e.Stages)
}

func (e MultiStepForm) ParseStage(r *http.Request) int {
	return e.requestStage(r)
}

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

func (e MultiStepForm) hiddenFieldsHTML(ctx context.Context, stageIdx int, values map[string]any, errors map[string]error) (string, error) {
	nodes := []Node{
		Input(Type("hidden"), Name("$stage"), Value(strconv.Itoa(stageIdx))),
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
			node, renderable := hiddenCarryNode(input, value, ctx)
			if !renderable {
				slog.Error("MultiStepForm hidden carry unsupported", "key", e.Key, "input", name, "type", fmt.Sprintf("%T", value))
				continue
			}
			nodes = append(nodes, node)
			seen[name] = struct{}{}
		}
	}
	nodes = append(nodes, hiddenErrorNodes(errors)...)

	var out strings.Builder
	for _, node := range nodes {
		html, err := renderNodeToString(node)
		if err != nil {
			return "", err
		}
		out.WriteString(html)
	}
	return out.String(), nil
}

func (e MultiStepForm) ribbonHTML(stageIdx int, errors map[string]error) (string, error) {
	nodes := []Node{
		Div(
			Class("flex flex-wrap items-center gap-2 mb-4"),
			Group(e.ribbonButtons(stageIdx, e.stageErrors(errors))),
		),
	}
	return renderNodesToString(nodes)
}

func (e MultiStepForm) ribbonButtons(stageIdx int, stageErrors map[int]struct{}) []Node {
	nodes := make([]Node, 0, len(e.Stages))
	for i := range e.Stages {
		label := fmt.Sprintf("Step %d", i+1)
		classes := e.ribbonButtonClasses(i, stageIdx, stageErrors)
		switch {
		case i == stageIdx:
			nodes = append(nodes, Button(
				Type("button"),
				Class(classes),
				Text(label),
			))
		case i < stageIdx:
			nodes = append(nodes, Button(
				Type("submit"),
				Name("$stage_target"),
				Value(strconv.Itoa(i)),
				Class(classes),
				Text(label),
			))
		default:
			nodes = append(nodes, Button(
				Type("button"),
				Class(classes),
				Disabled(),
				Text(label),
			))
		}
	}
	return nodes
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

func hiddenErrorNodes(errors map[string]error) []Node {
	if len(errors) == 0 {
		return nil
	}
	keys := make([]string, 0, len(errors))
	for key, err := range errors {
		if key == "" || err == nil {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	nodes := make([]Node, 0, len(keys))
	for _, key := range keys {
		nodes = append(nodes, Input(
			Type("hidden"),
			Name(multiStepFormErrorFieldPrefix+key),
			Value(errors[key].Error()),
		))
	}
	return nodes
}

func hiddenCarryNode(input InputInterface, value any, ctx context.Context) (Node, bool) {
	switch typed := input.(type) {
	case InputCheckbox:
		v, ok := value.(bool)
		if !ok {
			return nil, false
		}
		typed.Hidden = true
		typed.Getter = getters.Static(v)
		return typed.Build(ctx), true
	case *InputCheckbox:
		v, ok := value.(bool)
		if !ok {
			return nil, false
		}
		clone := *typed
		clone.Hidden = true
		clone.Getter = getters.Static(v)
		return clone.Build(ctx), true
	case InputDate:
		t, ok := timeValue(value)
		if !ok {
			return nil, false
		}
		typed.Hidden = true
		typed.Getter = getters.Static(t)
		return typed.Build(ctx), true
	case *InputDate:
		t, ok := timeValue(value)
		if !ok {
			return nil, false
		}
		clone := *typed
		clone.Hidden = true
		clone.Getter = getters.Static(t)
		return clone.Build(ctx), true
	case InputTime:
		t, ok := timeValue(value)
		if !ok {
			return nil, false
		}
		typed.Hidden = true
		typed.Getter = getters.Static(t)
		return typed.Build(ctx), true
	case *InputTime:
		t, ok := timeValue(value)
		if !ok {
			return nil, false
		}
		clone := *typed
		clone.Hidden = true
		clone.Getter = getters.Static(t)
		return clone.Build(ctx), true
	case InputDatetime:
		t, ok := timeValue(value)
		if !ok {
			return nil, false
		}
		typed.Hidden = true
		typed.Getter = getters.Static(t)
		return typed.Build(ctx), true
	case *InputDatetime:
		t, ok := timeValue(value)
		if !ok {
			return nil, false
		}
		clone := *typed
		clone.Hidden = true
		clone.Getter = getters.Static(t)
		return clone.Build(ctx), true
	default:
		return genericHiddenCarryNode(input.GetName(), value)
	}
}

func genericHiddenCarryNode(name string, value any) (Node, bool) {
	switch typed := value.(type) {
	case AssociationIDs:
		group := Group{}
		for _, id := range typed.IDs {
			group = append(group, Input(Type("hidden"), Name(name), Value(strconv.FormatUint(uint64(id), 10))))
		}
		return group, true
	case []string:
		group := Group{}
		for _, item := range typed {
			group = append(group, Input(Type("hidden"), Name(name), Value(item)))
		}
		return group, true
	case []uint:
		group := Group{}
		for _, item := range typed {
			group = append(group, Input(Type("hidden"), Name(name), Value(strconv.FormatUint(uint64(item), 10))))
		}
		return group, true
	default:
		if scalar, ok := scalarHiddenValue(value); ok {
			return Input(Type("hidden"), Name(name), Value(scalar)), true
		}
	}
	return nil, false
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

func renderNodeToString(node Node) (string, error) {
	var out bytes.Buffer
	if err := node.Render(&out); err != nil {
		return "", err
	}
	return out.String(), nil
}

func renderNodesToString(nodes []Node) (string, error) {
	var out strings.Builder
	for _, node := range nodes {
		html, err := renderNodeToString(node)
		if err != nil {
			return "", err
		}
		out.WriteString(html)
	}
	return out.String(), nil
}

func injectIntoRenderedForm(html, actionURL, prefixHTML, hiddenHTML string) (string, error) {
	formStart := strings.Index(html, "<form")
	if formStart == -1 {
		return "", fmt.Errorf("MultiStepForm: rendered stage missing form tag")
	}

	formTagEndOffset := strings.Index(html[formStart:], ">")
	if formTagEndOffset == -1 {
		return "", fmt.Errorf("MultiStepForm: rendered stage has malformed form tag")
	}
	formTagEnd := formStart + formTagEndOffset

	if actionURL != "" {
		formTag := html[formStart : formTagEnd+1]
		if multiStepFormActionAttr.MatchString(formTag) {
			formTag = multiStepFormActionAttr.ReplaceAllString(formTag, fmt.Sprintf(` action="%s"`, actionURL))
		} else {
			formTag = strings.TrimSuffix(formTag, ">") + fmt.Sprintf(` action="%s">`, actionURL)
		}
		html = html[:formStart] + formTag + html[formTagEnd+1:]
		formTagEndOffset = strings.Index(html[formStart:], ">")
		if formTagEndOffset == -1 {
			return "", fmt.Errorf("MultiStepForm: rendered stage has malformed form tag after action injection")
		}
		formTagEnd = formStart + formTagEndOffset
	}

	formEnd := strings.LastIndex(html, "</form>")
	if formEnd == -1 {
		return "", fmt.Errorf("MultiStepForm: rendered stage missing closing form tag")
	}
	return html[:formTagEnd+1] + prefixHTML + html[formTagEnd+1:formEnd] + hiddenHTML + html[formEnd:], nil
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
