package sqlagent

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"reflect"
	"strings"

	"github.com/lariv-in/lago/lago"
	"gorm.io/gorm"
	gormschema "gorm.io/gorm/schema"
)

// registrySchemaToolName identifies the hidden tool row that carries registry JSON for the model only.
const registrySchemaToolName = "registry_schema"

// registrySchemaUserMessagePrefix legacy: older chats stored bootstrap as a user message with this prefix.
const registrySchemaUserMessagePrefix = "<!--sqlagent:registry-schema-->\n"

// registrySchemasPayload is JSON-serializable metadata derived from gorm.Statement.Schema for each registered model.
type registrySchemasPayload struct {
	Version string           `json:"version"`
	Tables  []tableSchemaDTO `json:"tables"`
}

type tableSchemaDTO struct {
	RegistryKey string            `json:"registry_key"`
	Table       string            `json:"table"`
	ModelName   string            `json:"model_name"`
	ModelString string            `json:"model_string"`
	Fields      []fieldSchemaDTO  `json:"fields"`
	Relations   []relationSchemaDTO `json:"relations,omitempty"`
}

type fieldSchemaDTO struct {
	Name          string `json:"name"`
	DBName        string `json:"db_name"`
	BindName      string `json:"bind_name,omitempty"`
	DataType      string `json:"data_type"`
	GORMDataType  string `json:"gorm_data_type"`
	FieldType     string `json:"field_type,omitempty"`
	PrimaryKey    bool   `json:"primary_key"`
	AutoIncrement bool   `json:"auto_increment,omitempty"`
	NotNull       bool   `json:"not_null,omitempty"`
	Unique        bool   `json:"unique,omitempty"`
	Size          int    `json:"size,omitempty"`
	Precision     int    `json:"precision,omitempty"`
	Scale         int    `json:"scale,omitempty"`
	Comment       string `json:"comment,omitempty"`
}

type relationSchemaDTO struct {
	Name         string                   `json:"name"`
	Type         gormschema.RelationshipType `json:"type"`
	RelatedTable string                   `json:"related_table,omitempty"`
}

func isRegistrySchemaBootstrapUserContent(s string) bool {
	return strings.HasPrefix(s, registrySchemaUserMessagePrefix)
}

func isRegistrySchemaToolMessage(msg *ConversationMessage) bool {
	return msg != nil && msg.Kind == MessageKindTool && msg.ToolMessage != nil &&
		msg.ToolMessage.Name == registrySchemaToolName
}

// MarshalRegistrySchemasJSON walks lago.RegistryModel, parses each model with GORM, and returns indented JSON.
func MarshalRegistrySchemasJSON(db *gorm.DB) ([]byte, error) {
	payload, err := buildRegistrySchemasPayload(db)
	if err != nil {
		logError("sqlagent: build registry schemas payload", err)
		return nil, err
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		logError("sqlagent: marshal registry schemas JSON", err)
	}
	return b, err
}

func buildRegistrySchemasPayload(db *gorm.DB) (registrySchemasPayload, error) {
	pairs := lago.RegistryModel.AllStable()
	out := registrySchemasPayload{Version: "1", Tables: make([]tableSchemaDTO, 0, len(*pairs))}
	for _, pair := range *pairs {
		dto, err := schemaFromRegistryPair(db, pair.Key, pair.Value)
		if err != nil {
			slog.Warn("sqlagent: skip registry model schema", "registry_key", pair.Key, "error", err)
			continue
		}
		out.Tables = append(out.Tables, dto)
	}
	return out, nil
}

func schemaFromRegistryPair(db *gorm.DB, registryKey string, model any) (tableSchemaDTO, error) {
	var zero tableSchemaDTO
	rv := reflect.ValueOf(model)
	if !rv.IsValid() {
		return zero, errors.New("invalid model value")
	}
	if rv.Kind() != reflect.Pointer {
		ptr := reflect.New(rv.Type())
		ptr.Elem().Set(rv)
		rv = ptr
	}
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(rv.Interface()); err != nil {
		return zero, err
	}
	s := stmt.Schema
	if s == nil {
		return zero, gormschema.ErrUnsupportedDataType
	}
	return tableSchemaDTO{
		RegistryKey: registryKey,
		Table:       s.Table,
		ModelName:   s.Name,
		ModelString: s.String(),
		Fields:      fieldsToDTOs(s.Fields),
		Relations:   relationshipsToDTOs(&s.Relationships),
	}, nil
}

func fieldsToDTOs(fields []*gormschema.Field) []fieldSchemaDTO {
	out := make([]fieldSchemaDTO, 0, len(fields))
	for _, f := range fields {
		if f == nil {
			continue
		}
		ft := ""
		if f.FieldType != nil {
			ft = f.FieldType.String()
		}
		out = append(out, fieldSchemaDTO{
			Name:          f.Name,
			DBName:        f.DBName,
			BindName:      f.BindName(),
			DataType:      string(f.DataType),
			GORMDataType:  string(f.GORMDataType),
			FieldType:     ft,
			PrimaryKey:    f.PrimaryKey,
			AutoIncrement: f.AutoIncrement,
			NotNull:       f.NotNull,
			Unique:        f.Unique,
			Size:          f.Size,
			Precision:     f.Precision,
			Scale:         f.Scale,
			Comment:       f.Comment,
		})
	}
	return out
}

func relationshipsToDTOs(rs *gormschema.Relationships) []relationSchemaDTO {
	if rs == nil {
		return nil
	}
	var out []relationSchemaDTO
	add := func(rels []*gormschema.Relationship) {
		for _, r := range rels {
			if r == nil {
				continue
			}
			d := relationSchemaDTO{Name: r.Name, Type: r.Type}
			if r.FieldSchema != nil {
				d.RelatedTable = r.FieldSchema.Table
			}
			out = append(out, d)
		}
	}
	add(rs.HasOne)
	add(rs.BelongsTo)
	add(rs.HasMany)
	add(rs.Many2Many)
	return out
}

// ensureRegistrySchemaFirstMessage inserts a hidden tool message with serialized registry schemas when the conversation is still empty.
func ensureRegistrySchemaFirstMessage(db *gorm.DB, conversationID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var n int64
		if err := tx.Model(&ConversationMessage{}).Where("conversation_id = ?", conversationID).Count(&n).Error; err != nil {
			logError("sqlagent: count messages for schema bootstrap", err, "conversation_id", conversationID)
			return err
		}
		if n > 0 {
			return nil
		}
		blob, err := MarshalRegistrySchemasJSON(tx)
		if err != nil {
			return err
		}
		msg := ConversationMessage{
			ConversationID: conversationID,
			SortOrder:      0,
			Kind:           MessageKindTool,
		}
		if err := gorm.G[ConversationMessage](tx).Create(context.Background(), &msg); err != nil {
			logError("sqlagent: create schema tool message envelope", err, "conversation_id", conversationID)
			return err
		}
		tm := ToolMessage{
			ConversationMessageID: msg.ID,
			Name:                  registrySchemaToolName,
			Detail:                string(blob),
		}
		if err := gorm.G[ToolMessage](tx).Create(context.Background(), &tm); err != nil {
			logError("sqlagent: create schema tool message body", err, "conversation_id", conversationID)
			return err
		}
		return nil
	})
}
