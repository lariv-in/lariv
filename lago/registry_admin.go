package lago

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

func setFieldFromString(fv reflect.Value, s string) {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			fv.SetInt(n)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if n, err := strconv.ParseUint(s, 10, 64); err == nil {
			fv.SetUint(n)
		}
	case reflect.Float32, reflect.Float64:
		if n, err := strconv.ParseFloat(s, 64); err == nil {
			fv.SetFloat(n)
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(s); err == nil {
			fv.SetBool(b)
		}
	case reflect.Slice:
		if fv.Type().Elem().Kind() == reflect.Uint8 {
			fv.SetBytes([]byte(s))
		}
	}
}

type AdminPanel[T any] struct {
	SearchField string
	ListFields  []string
	Preload     []string
}

func (a AdminPanel[T]) IsAdminPanel() bool {
	return true
}

func (a AdminPanel[T]) ModelName() string {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Name()
}

func (a AdminPanel[T]) List(db *gorm.DB, page, pageSize int) ([]map[string]any, error) {
	offset := (page - 1) * pageSize
	chain := gorm.G[T](db).Where("TRUE").Offset(offset).Limit(pageSize)
	for _, p := range a.Preload {
		chain = chain.Preload(p, nil)
	}
	results, err := chain.Find(context.Background())
	if err != nil {
		return nil, err
	}

	rows := make([]map[string]any, len(results))
	for i, item := range results {
		rows[i] = getters.MapFromStruct(item)
	}
	return rows, nil
}

func (a AdminPanel[T]) GetListFields() []string {
	return a.ListFields
}

func (a AdminPanel[T]) Save(db *gorm.DB, id string, values map[string]*string) error {
	record, err := gorm.G[T](db).Where("id = ?", id).First(context.Background())
	if err != nil {
		return err
	}

	v := reflect.ValueOf(&record).Elem()
	t := v.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() || field.Anonymous {
			continue
		}
		strPtr, exists := values[field.Name]
		if !exists || strPtr == nil {
			continue
		}
		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}
		setFieldFromString(fv, *strPtr)
	}

	return db.Save(&record).Error
}

func (a AdminPanel[T]) EditableFields() []string {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	var fields []string
	for f := range t.Fields() {
		if !f.IsExported() || f.Anonymous {
			continue
		}
		if f.Type.Kind() == reflect.Struct {
			continue
		}
		if f.Type.Kind() == reflect.Slice && f.Type.Elem().Kind() != reflect.Uint8 {
			continue
		}
		fields = append(fields, f.Name)
	}
	return fields
}

func (a AdminPanel[T]) Create(db *gorm.DB, values map[string]*string) error {
	var record T
	v := reflect.ValueOf(&record).Elem()
	t := v.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() || field.Anonymous {
			continue
		}
		strPtr, exists := values[field.Name]
		if !exists || strPtr == nil {
			continue
		}
		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}
		setFieldFromString(fv, *strPtr)
	}
	return gorm.G[T](db).Create(context.Background(), &record)
}

func (a AdminPanel[T]) ImportCSV(db *gorm.DB, path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	headers, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("reading CSV headers: %w", err)
	}

	var zero T
	structType := reflect.TypeOf(zero)
	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}

	// Map CSV column index to struct field index
	colToField := make([]int, len(headers))
	for i, header := range headers {
		colToField[i] = -1
		for j := range structType.NumField() {
			if structType.Field(j).Name == header {
				colToField[i] = j
				break
			}
		}
	}

	rows, err := reader.ReadAll()
	if err != nil {
		return 0, fmt.Errorf("reading CSV rows: %w", err)
	}

	created := 0
	for _, row := range rows {
		record := reflect.New(structType).Elem()
		for i, val := range row {
			fi := colToField[i]
			if fi < 0 {
				continue
			}
			fv := record.Field(fi)
			if fv.CanSet() {
				setFieldFromString(fv, val)
			}
		}
		rec := record.Addr().Interface().(*T)
		if err := gorm.G[T](db).Create(context.Background(), rec); err != nil {
			return created, fmt.Errorf("row %d: %w", created+1, err)
		}
		created++
	}
	return created, nil
}

func (a AdminPanel[T]) ExportCSV(db *gorm.DB, path string) (int, error) {
	chain := gorm.G[T](db).Where("TRUE")
	for _, p := range a.Preload {
		chain = chain.Preload(p, nil)
	}
	results, err := chain.Find(context.Background())
	if err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, fmt.Errorf("no records to export")
	}

	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	rows := make([]map[string]any, len(results))
	for i, item := range results {
		rows[i] = getters.MapFromStruct(item)
	}

	// Collect headers from first row
	headers := make([]string, 0, len(rows[0]))
	for k := range rows[0] {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	if err := writer.Write(headers); err != nil {
		return 0, err
	}

	for _, row := range rows {
		record := make([]string, len(headers))
		for i, h := range headers {
			record[i] = fmt.Sprintf("%v", row[h])
		}
		if err := writer.Write(record); err != nil {
			return 0, err
		}
	}

	return len(rows), nil
}

type AdminPanelInterface interface {
	IsAdminPanel() bool
	ModelName() string
	GetListFields() []string
	EditableFields() []string
	List(db *gorm.DB, page, pageSize int) ([]map[string]any, error)
	Save(db *gorm.DB, id string, values map[string]*string) error
	Create(db *gorm.DB, values map[string]*string) error
	ImportCSV(db *gorm.DB, path string) (int, error)
	ExportCSV(db *gorm.DB, path string) (int, error)
}

var RegistryAdmin *registry.Registry[AdminPanelInterface] = registry.NewRegistry[AdminPanelInterface]()
