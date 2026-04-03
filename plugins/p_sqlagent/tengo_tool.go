package sqlagent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"reflect"
	"strings"

	"github.com/d5/tengo/v2"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"gorm.io/gorm"
)

const tengoToolName = "run_tengo"

type sqlAgentGormTxKey struct{}

// ContextWithGormTx attaches a *gorm.DB (typically a transaction from db.Transaction) so run_tengo can expose it to Tengo as variable "tx".
func ContextWithGormTx(ctx context.Context, gdb *gorm.DB) context.Context {
	return context.WithValue(ctx, sqlAgentGormTxKey{}, gdb)
}

func gormTxFromContext(ctx context.Context) (*gorm.DB, bool) {
	v := ctx.Value(sqlAgentGormTxKey{})
	if v == nil {
		return nil, false
	}
	tx, ok := v.(*gorm.DB)
	return tx, ok && tx != nil
}

var int64PtrType = reflect.TypeOf((*int64)(nil))

// gormRefInt64 is a mutable int64 cell that converts to *int64 for GORM APIs (Count, etc.) and reads back via ["value"].
type gormRefInt64 struct {
	tengo.ObjectImpl
	v int64
}

func (r *gormRefInt64) TypeName() string { return "gorm-ref-int64" }

func (r *gormRefInt64) String() string { return fmt.Sprintf("<gorm-ref-int64 %d>", r.v) }

func (r *gormRefInt64) Copy() tengo.Object { return &gormRefInt64{v: r.v} }

func (r *gormRefInt64) IndexGet(index tengo.Object) (tengo.Object, error) {
	key, ok := index.(*tengo.String)
	if !ok {
		return nil, tengo.ErrInvalidIndexType
	}
	switch key.Value {
	case "value", "v":
		return &tengo.Int{Value: r.v}, nil
	default:
		return nil, tengo.ErrInvalidIndexValueType
	}
}

func newGormRefInt64(args ...tengo.Object) (tengo.Object, error) {
	switch len(args) {
	case 0:
		return &gormRefInt64{}, nil
	case 1:
		i, ok := tengo.ToInt64(args[0])
		if !ok {
			return nil, fmt.Errorf("gorm_ref_int64: expected number, got %s", args[0].TypeName())
		}
		return &gormRefInt64{v: i}, nil
	default:
		return nil, fmt.Errorf("gorm_ref_int64: want 0 or 1 arguments, got %d", len(args))
	}
}

// gormDBObject wraps *gorm.DB so Tengo scripts can call GORM methods via tx["Method"](args...).
type gormDBObject struct {
	tengo.ObjectImpl
	db *gorm.DB
}

func (o *gormDBObject) TypeName() string { return "gorm-db" }

func (o *gormDBObject) String() string {
	if o.db == nil {
		return "<gorm.DB nil>"
	}
	return fmt.Sprintf("<gorm.DB %p>", o.db)
}

func (o *gormDBObject) Copy() tengo.Object {
	return &gormDBObject{db: o.db}
}

func (o *gormDBObject) IndexGet(index tengo.Object) (tengo.Object, error) {
	key, ok := index.(*tengo.String)
	if !ok {
		return nil, tengo.ErrInvalidIndexType
	}
	if o.db == nil {
		return nil, errors.New("gorm.DB is nil")
	}
	rv := reflect.ValueOf(o.db)
	m := rv.MethodByName(key.Value)
	if !m.IsValid() {
		return nil, fmt.Errorf("gorm.DB has no method %q", key.Value)
	}
	return &tengo.UserFunction{
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			out, err := callBoundMethod(m, args)
			if err != nil {
				return nil, err
			}
			return reflectOutputsToObject(out, m.Type())
		},
	}, nil
}

func callBoundMethod(bound reflect.Value, args []tengo.Object) ([]reflect.Value, error) {
	mt := bound.Type()
	n := mt.NumIn()
	if !mt.IsVariadic() {
		if len(args) != n {
			return nil, fmt.Errorf("wrong number of arguments: want %d got %d", n, len(args))
		}
		in := make([]reflect.Value, n)
		for i := 0; i < n; i++ {
			v, err := tengoObjectToValue(args[i], mt.In(i))
			if err != nil {
				return nil, fmt.Errorf("arg %d: %w", i, err)
			}
			in[i] = v
		}
		return bound.Call(in), nil
	}
	if len(args) < n-1 {
		return nil, fmt.Errorf("wrong number of arguments: want at least %d got %d", n-1, len(args))
	}
	in := make([]reflect.Value, n)
	for i := 0; i < n-1; i++ {
		v, err := tengoObjectToValue(args[i], mt.In(i))
		if err != nil {
			return nil, fmt.Errorf("arg %d: %w", i, err)
		}
		in[i] = v
	}
	lastType := mt.In(n - 1)
	elemType := lastType.Elem()
	slice := reflect.MakeSlice(lastType, 0, len(args)-(n-1))
	for i, a := range args[n-1:] {
		v, err := tengoObjectToValue(a, elemType)
		if err != nil {
			return nil, fmt.Errorf("variadic arg %d: %w", i, err)
		}
		slice = reflect.Append(slice, v)
	}
	in[n-1] = slice
	return bound.Call(in), nil
}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func reflectOutputsToObject(out []reflect.Value, mt reflect.Type) (tengo.Object, error) {
	if len(out) == 0 {
		return tengo.UndefinedValue, nil
	}
	if len(out) == 2 && out[1].Type().Implements(errorInterface) {
		if err, ok := out[1].Interface().(error); ok && err != nil {
			return nil, err
		}
		return goValueToObject(out[0])
	}
	if len(out) == 1 {
		return goValueToObject(out[0])
	}
	objs := make([]tengo.Object, len(out))
	for i, v := range out {
		o, err := goValueToObject(v)
		if err != nil {
			return nil, err
		}
		objs[i] = o
	}
	return &tengo.Array{Value: objs}, nil
}

func goValueToObject(v reflect.Value) (tengo.Object, error) {
	if !v.IsValid() {
		return tengo.UndefinedValue, nil
	}
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return tengo.UndefinedValue, nil
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return tengo.UndefinedValue, nil
	}
	x := v.Interface()
	if db, ok := x.(*gorm.DB); ok {
		return &gormDBObject{db: db}, nil
	}
	obj, err := tengo.FromInterface(x)
	if err == nil {
		return obj, nil
	}
	return &tengo.String{Value: fmt.Sprint(x)}, nil
}

func tengoObjectToValue(o tengo.Object, want reflect.Type) (reflect.Value, error) {
	if want.Kind() == reflect.Interface && want.NumMethod() == 0 {
		if g, ok := o.(*gormDBObject); ok {
			return reflect.ValueOf(g.db), nil
		}
		if ref, ok := o.(*gormRefInt64); ok {
			return reflect.ValueOf(&ref.v), nil
		}
		x := tengo.ToInterface(o)
		if xo, ok := x.(tengo.Object); ok {
			x = tengo.ToInterface(xo)
		}
		if x == nil {
			return reflect.Zero(want), nil
		}
		rv := reflect.ValueOf(x)
		if rv.Type().AssignableTo(want) {
			return rv, nil
		}
		if rv.Type().ConvertibleTo(want) {
			return rv.Convert(want), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot convert %T to empty interface expectation", x)
	}

	switch want.Kind() {
	case reflect.String:
		s, ok := tengo.ToString(o)
		if !ok {
			return reflect.Value{}, errors.New("expected string")
		}
		return reflect.ValueOf(s), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, ok := tengo.ToInt64(o)
		if !ok {
			return reflect.Value{}, errors.New("expected int")
		}
		return reflect.ValueOf(i).Convert(want), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, ok := tengo.ToInt64(o)
		if !ok || i < 0 {
			return reflect.Value{}, errors.New("expected non-negative int")
		}
		return reflect.ValueOf(uint64(i)).Convert(want), nil
	case reflect.Float32, reflect.Float64:
		f, ok := tengo.ToFloat64(o)
		if !ok {
			return reflect.Value{}, errors.New("expected float")
		}
		return reflect.ValueOf(f).Convert(want), nil
	case reflect.Bool:
		b, ok := tengo.ToBool(o)
		if !ok {
			return reflect.Value{}, errors.New("expected bool")
		}
		return reflect.ValueOf(b), nil
	case reflect.Slice:
		if want.Elem().Kind() == reflect.Uint8 {
			bs, ok := tengo.ToByteSlice(o)
			if !ok {
				return reflect.Value{}, errors.New("expected bytes")
			}
			return reflect.ValueOf(bs), nil
		}
		arr, ok := o.(*tengo.Array)
		if !ok {
			return reflect.Value{}, errors.New("expected array")
		}
		sl := reflect.MakeSlice(want, len(arr.Value), len(arr.Value))
		for i, el := range arr.Value {
			ev, err := tengoObjectToValue(el, want.Elem())
			if err != nil {
				return reflect.Value{}, err
			}
			sl.Index(i).Set(ev)
		}
		return sl, nil
	case reflect.Ptr:
		if o == tengo.UndefinedValue {
			return reflect.Zero(want), nil
		}
		if want == reflect.TypeOf((*gorm.DB)(nil)) {
			if g, ok := o.(*gormDBObject); ok {
				return reflect.ValueOf(g.db), nil
			}
		}
		if want == int64PtrType {
			if ref, ok := o.(*gormRefInt64); ok {
				return reflect.ValueOf(&ref.v), nil
			}
		}
	}
	return reflect.Value{}, fmt.Errorf("unsupported parameter type %s", want)
}

func scriptAdd(name string, value any, script *tengo.Script) error {
	if db, ok := value.(*gorm.DB); ok {
		return script.Add(name, &gormDBObject{db: db})
	}
	return script.Add(name, value)
}

func resultFromCompiled(compiled *tengo.Compiled) any {
	if !compiled.IsDefined("result") {
		return nil
	}
	v := compiled.Get("result")
	if v.IsUndefined() {
		return nil
	}
	return objectToJSONable(v.Object())
}

func objectToJSONable(o tengo.Object) any {
	switch x := o.(type) {
	case *gormRefInt64:
		return x.v
	case *gormDBObject:
		return map[string]any{"_gorm_db": true}
	case *tengo.Error:
		return map[string]any{"error": x.String()}
	}
	v := tengo.ToInterface(o)
	if obj, ok := v.(tengo.Object); ok {
		return obj.String()
	}
	return v
}

type tengoToolInput struct {
	Code string         `json:"code"`
	Env  map[string]any `json:"env,omitempty"`
}

type tengoToolOutput struct {
	Result any `json:"result"`
}

func tengoToolHandler(tctx tool.Context, in tengoToolInput) (tengoToolOutput, error) {
	slog.Info("sqlagent: run_tengo", "code", in.Code, "env", in.Env)
	code := strings.TrimSpace(in.Code)
	if code == "" {
		err := errors.New("code is required")
		logError("sqlagent: run_tengo", err)
		return tengoToolOutput{}, err
	}
	script := tengo.NewScript([]byte(code))
	if err := script.Add("result", nil); err != nil {
		logError("sqlagent: tengo script Add result", err)
		return tengoToolOutput{}, err
	}
	if err := script.Add("gorm_ref_int64", tengo.CallableFunc(newGormRefInt64)); err != nil {
		logError("sqlagent: tengo script Add gorm_ref_int64", err)
		return tengoToolOutput{}, err
	}
	env := maps.Clone(in.Env)
	if env == nil {
		env = make(map[string]any)
	}
	for k, v := range env {
		if err := scriptAdd(k, v, script); err != nil {
			logError("sqlagent: tengo script Add", err, "name", k)
			return tengoToolOutput{}, err
		}
	}
	if tx, ok := gormTxFromContext(tctx); ok {
		if err := scriptAdd("tx", tx, script); err != nil {
			logError("sqlagent: tengo script Add tx", err)
			return tengoToolOutput{}, err
		}
	}
	compiled, err := script.RunContext(tctx)
	if err != nil {
		logError("sqlagent: tengo RunContext", err)
		return tengoToolOutput{}, err
	}
	return tengoToolOutput{Result: resultFromCompiled(compiled)}, nil
}

func newTengoTool() (tool.Tool, error) {
	t, err := functiontool.New(functiontool.Config{
		Name: tengoToolName,
		Description: `Runs a Tengo script (https://github.com/d5/tengo) — Tengo syntax, NOT Go. Fields: "code" (string, Tengo source), optional "env" (object: extra variables, JSON-safe values only).

Output: global "result" is predeclared (undefined). Set it with assignment using "=": result = <expression>. Prefer "result = ..." over "result := ..." so you assign the tool output global.

GORM on "tx" (request-scoped *gorm.DB): call methods ONLY with string keys and parentheses — tx["Table"]("contacts"), tx["Where"]("id = ?", 1). Do NOT use Go-only syntax: no &, no var, no types (e.g. int64), no interface{}, no map[string]interface{}{}, no tx.Model(...) with Go map literals.

Go APIs that need *int64 (e.g. gorm Count): Tengo has no pointers. Use the builtin gorm_ref_int64() to allocate a mutable int64 cell, pass it to the method, then read n["value"] (or n["v"]). Example row count:
  n := gorm_ref_int64(); tx["Table"]("contacts")["Count"](n); result = n["value"]
Optional initial value: gorm_ref_int64(0). Same pattern for other GORM methods that take *int64.

Other examples: env {"n": 3} and code: result = n * 2

Returns {"result": ...} from the final value of global "result" (nil if never set).`,
	}, tengoToolHandler)
	if err != nil {
		logError("sqlagent: functiontool New run_tengo", err)
	}
	return t, err
}
