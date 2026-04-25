package p_seer_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	"google.golang.org/genai"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SeerAssistantSession struct {
	gorm.Model

	Title  string `gorm:"notnull;default:''"`
	UserID uint   `gorm:"index"`
}

func (m SeerAssistantSession) SaveContent(ctx context.Context, content genai.Content) error {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	sessionMessage := SeerAssistantSessionMessage{
		SeerAssistantSessionID: m.ID,
		SeerAssistantSession:   m,
		Role:                   content.Role,
	}
	err = gorm.G[SeerAssistantSessionMessage](db).Create(ctx, &sessionMessage)
	if err != nil {
		return err
	}
	return sessionMessage.SaveParts(ctx, content.Parts)
}

type SeerAssistantSessionMessage struct {
	gorm.Model

	SeerAssistantSessionID uint                 `gorm:"notnull"`
	SeerAssistantSession   SeerAssistantSession `gorm:"notnull"`
	Role                   string               `gorm:"notnull;default:'user'"`
}

func (m SeerAssistantSessionMessage) SaveParts(ctx context.Context, parts []*genai.Part) error {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	messageKinds := SeerAssistantSessionMessageTypes.All()
	for _, part := range parts {
		messageKind := ""
		for kind, kindModel := range messageKinds {
			if kindModel.IsPartType(part) {
				messageKind = kind
				break
			}
		}
		if messageKind == "" {
			return fmt.Errorf("unknown kind of part found: %#v", part)
		}
		var videoMetadata *VideoMetadata
		var videoMetadataID *uint
		if part.VideoMetadata != nil {
			videoMetadata = &VideoMetadata{
				EndOffset:   part.VideoMetadata.EndOffset,
				FPS:         part.VideoMetadata.FPS,
				StartOffset: part.VideoMetadata.StartOffset,
			}
			err = gorm.G[VideoMetadata](db).Create(ctx, videoMetadata)
			if err != nil {
				return err
			}
			videoMetadataID = &videoMetadata.ID
		}
		partMetadata, err := json.Marshal(part.PartMetadata)
		if err != nil {
			return err
		}
		messagePart := SeerAssistantSessionMessagePart{
			Kind:                          messageKind,
			SeerAssistantSessionMessageID: m.ID,
			SeerAssistantSessionMessage:   m,
			Thought:                       part.Thought,
			ThoughtSignature:              part.ThoughtSignature,
			VideoMetadata:                 videoMetadata,
			VideoMetadataID:               videoMetadataID,
			PartMetadata:                  datatypes.JSON(partMetadata),
		}
		err = gorm.G[SeerAssistantSessionMessagePart](db).Create(ctx, &messagePart)
		if err != nil {
			return err
		}
		err = messagePart.SavePartType(ctx, part)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m SeerAssistantSessionMessage) LoadContent(ctx context.Context) (*genai.Content, error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}
	parts, err := gorm.G[SeerAssistantSessionMessagePart](db).Where("seer_assistant_session_message_id = ?", m.ID).Find(ctx)
	if err != nil {
		return nil, err
	}
	content := genai.Content{
		Role:  m.Role,
		Parts: make([]*genai.Part, len(parts)),
	}
	for i, part := range parts {
		partTypeModel, err := loadMessageTypeModel(db, part.Kind, part.ID)
		if err != nil {
			return nil, err
		}
		content.Parts[i], err = partTypeModel.Part(ctx)
		if err != nil {
			return nil, err
		}
	}
	return &content, nil
}

type SeerAssistantSessionMessageType interface {
	GenaiType() string
	Part(context.Context) (*genai.Part, error)
	IsPartType(*genai.Part) bool
	Save(context.Context, *genai.Part) error
}

type seerAssistantSessionMessageTypeWithPart interface {
	SeerAssistantSessionMessageType
	withMessagePart(SeerAssistantSessionMessagePart) SeerAssistantSessionMessageType
}

// Maps between SeerAssistantSessionMessage.Kind and model.
var SeerAssistantSessionMessageTypes = registry.NewRegistry[SeerAssistantSessionMessageType]()

func RegisterMessageType[T SeerAssistantSessionMessageType]() {
	var zero T
	SeerAssistantSessionMessageTypes.Register(zero.GenaiType(), zero)
}

func loadMessageTypeModel(db *gorm.DB, kind string, partID uint) (SeerAssistantSessionMessageType, error) {
	partTypeModel, isTypeKnown := SeerAssistantSessionMessageTypes.Get(kind)
	if !isTypeKnown {
		return nil, fmt.Errorf("unknown kind of part type: %q", kind)
	}
	partTypeModelValue := reflect.New(reflect.TypeOf(partTypeModel))
	err := db.Preload("SeerAssistantSessionMessagePart").Where("seer_assistant_session_message_part_id = ?", partID).First(partTypeModelValue.Interface()).Error
	if err != nil {
		return nil, err
	}
	loadedPartTypeModel, ok := partTypeModelValue.Elem().Interface().(SeerAssistantSessionMessageType)
	if !ok {
		return nil, fmt.Errorf("loaded message part type %q has wrong type", kind)
	}
	return loadedPartTypeModel, nil
}

type SeerAssistantSessionMessagePart struct {
	gorm.Model

	Kind                          string                      `gorm:"notnull"`
	SeerAssistantSessionMessageID uint                        `gorm:"notnull"`
	SeerAssistantSessionMessage   SeerAssistantSessionMessage `gorm:"notnull"`
	Thought                       bool                        `gorm:"notnull;default:false"`
	ThoughtSignature              []byte
	VideoMetadataID               *uint
	VideoMetadata                 *VideoMetadata
	PartMetadata                  datatypes.JSON
}

func (m SeerAssistantSessionMessagePart) SavePartType(ctx context.Context, part *genai.Part) error {
	partType, isPartTypeKnown := SeerAssistantSessionMessageTypes.Get(m.Kind)
	if !isPartTypeKnown {
		return fmt.Errorf("part type is unknown")
	}
	partTypeWithModel, ok := partType.(seerAssistantSessionMessageTypeWithPart)
	if !ok {
		return fmt.Errorf("part type %q cannot bind message part", m.Kind)
	}
	return partTypeWithModel.withMessagePart(m).Save(ctx, part)
}

type SeerAssistantSessionMessagePartModel struct {
	gorm.Model

	SeerAssistantSessionMessagePartID uint                            `gorm:"notnull"`
	SeerAssistantSessionMessagePart   SeerAssistantSessionMessagePart `gorm:"notnull"`
}

func (m SeerAssistantSessionMessagePart) ApplyToPart(part *genai.Part) (*genai.Part, error) {
	part.Thought = m.Thought
	part.ThoughtSignature = m.ThoughtSignature
	if m.VideoMetadata != nil {
		part.VideoMetadata = &genai.VideoMetadata{
			EndOffset:   m.VideoMetadata.EndOffset,
			FPS:         m.VideoMetadata.FPS,
			StartOffset: m.VideoMetadata.StartOffset,
		}
	}
	if len(m.PartMetadata) > 0 {
		var partMetadata map[string]any
		err := json.Unmarshal(m.PartMetadata, &partMetadata)
		if err != nil {
			return part, err
		}
		part.PartMetadata = partMetadata
	}
	return part, nil
}

func (m SeerAssistantSessionMessagePartModel) ApplyToPart(part *genai.Part) (*genai.Part, error) {
	return m.SeerAssistantSessionMessagePart.ApplyToPart(part)
}

type VideoMetadata struct {
	gorm.Model
	EndOffset   time.Duration
	FPS         *float64
	StartOffset time.Duration
}

type SeerAssistantSessionMessageInlineData struct {
	SeerAssistantSessionMessagePartModel

	MIMEType    string `gorm:"notnull"`
	Data        []byte `gorm:"notnull"`
	DisplayName string
}

func (m SeerAssistantSessionMessageInlineData) GenaiType() string {
	return "inlineData"
}

func (m SeerAssistantSessionMessageInlineData) withMessagePart(part SeerAssistantSessionMessagePart) SeerAssistantSessionMessageType {
	m.SeerAssistantSessionMessagePartModel = SeerAssistantSessionMessagePartModel{
		SeerAssistantSessionMessagePartID: part.ID,
		SeerAssistantSessionMessagePart:   part,
	}
	return m
}

func (m SeerAssistantSessionMessageInlineData) IsPartType(part *genai.Part) bool {
	return part != nil && part.InlineData != nil
}

func (m SeerAssistantSessionMessageInlineData) Part(_ context.Context) (*genai.Part, error) {
	return m.ApplyToPart(&genai.Part{
		InlineData: &genai.Blob{
			Data:        m.Data,
			MIMEType:    m.MIMEType,
			DisplayName: m.DisplayName,
		},
	})
}

func (m SeerAssistantSessionMessageInlineData) Save(ctx context.Context, part *genai.Part) error {
	if !m.IsPartType(part) {
		return fmt.Errorf("part is not inlineData")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	m.MIMEType = part.InlineData.MIMEType
	m.Data = part.InlineData.Data
	m.DisplayName = part.InlineData.DisplayName
	return gorm.G[SeerAssistantSessionMessageInlineData](db).Create(ctx, &m)
}

type SeerAssistantSessionMessageFunctionResponse struct {
	SeerAssistantSessionMessagePartModel

	WillContinue       *bool
	Scheduling         string `gorm:"default:'WHEN_IDLE'"`
	FunctionResponseID string
	Name               string `gorm:"notnull"`
	Response           datatypes.JSON
}

func (m SeerAssistantSessionMessageFunctionResponse) GenaiType() string {
	return "functionResponse"
}

func (m SeerAssistantSessionMessageFunctionResponse) withMessagePart(part SeerAssistantSessionMessagePart) SeerAssistantSessionMessageType {
	m.SeerAssistantSessionMessagePartModel = SeerAssistantSessionMessagePartModel{
		SeerAssistantSessionMessagePartID: part.ID,
		SeerAssistantSessionMessagePart:   part,
	}
	return m
}

func (m SeerAssistantSessionMessageFunctionResponse) IsPartType(part *genai.Part) bool {
	return part != nil && part.FunctionResponse != nil
}

func (m SeerAssistantSessionMessageFunctionResponse) Part(ctx context.Context) (*genai.Part, error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}
	parts, err := gorm.G[SeerAssistantSessionMessageFunctionResponsePart](db).Where("seer_assistant_session_message_function_response_id = ?", m.ID).Find(ctx)
	if err != nil {
		return nil, err
	}
	var response map[string]any
	if len(m.Response) > 0 {
		err = json.Unmarshal(m.Response, &response)
		if err != nil {
			return nil, err
		}
	}

	functionResponse := genai.FunctionResponse{
		WillContinue: m.WillContinue,
		Scheduling:   genai.FunctionResponseScheduling(m.Scheduling),
		Parts:        make([]*genai.FunctionResponsePart, len(parts)),
		ID:           m.FunctionResponseID,
		Name:         m.Name,
		Response:     response,
	}

	for i, part := range parts {
		functionResponsePartTypeModel, err := loadFunctionResponsePartTypeModel(db, part.Kind, part.ID)
		if err != nil {
			return nil, err
		}
		functionResponse.Parts[i], err = functionResponsePartTypeModel.FunctionResponsePart(ctx)
		if err != nil {
			return nil, err
		}
	}
	return m.ApplyToPart(&genai.Part{
		FunctionResponse: &functionResponse,
	})
}

func (m SeerAssistantSessionMessageFunctionResponse) Save(ctx context.Context, part *genai.Part) error {
	if !m.IsPartType(part) {
		return fmt.Errorf("part is not functionResponse")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	responseJSON, err := json.Marshal(part.FunctionResponse.Response)
	if err != nil {
		return err
	}
	m.WillContinue = part.FunctionResponse.WillContinue
	m.Scheduling = string(part.FunctionResponse.Scheduling)
	if m.Scheduling == "" {
		m.Scheduling = string(genai.FunctionResponseSchedulingWhenIdle)
	}
	m.FunctionResponseID = part.FunctionResponse.ID
	m.Name = part.FunctionResponse.Name
	m.Response = datatypes.JSON(responseJSON)
	err = gorm.G[SeerAssistantSessionMessageFunctionResponse](db).Create(ctx, &m)
	if err != nil {
		return err
	}
	functionResponsePartKinds := SeerAssistantSessionMessageFunctionResponsePartTypes.All()
	for _, functionResponsePart := range part.FunctionResponse.Parts {
		functionResponsePartKind := ""
		for kind, kindModel := range functionResponsePartKinds {
			if kindModel.IsFunctionResponsePartType(functionResponsePart) {
				functionResponsePartKind = kind
				break
			}
		}
		if functionResponsePartKind == "" {
			return fmt.Errorf("unknown kind of function response part found: %#v", functionResponsePart)
		}
		messageFunctionResponsePart := SeerAssistantSessionMessageFunctionResponsePart{
			SeerAssistantSessionMessageFunctionResponseID: m.ID,
			SeerAssistantSessionMessageFunctionResponse:   m,
			Kind: functionResponsePartKind,
		}
		err = gorm.G[SeerAssistantSessionMessageFunctionResponsePart](db).Create(ctx, &messageFunctionResponsePart)
		if err != nil {
			return err
		}
		err = messageFunctionResponsePart.SaveFunctionResponsePartType(ctx, functionResponsePart)
		if err != nil {
			return err
		}
	}
	return nil
}

type SeerAssistantSessionMessageFunctionResponsePart struct {
	gorm.Model

	SeerAssistantSessionMessageFunctionResponseID uint                                        `gorm:"notnull"`
	SeerAssistantSessionMessageFunctionResponse   SeerAssistantSessionMessageFunctionResponse `gorm:"notnull"`
	Kind                                          string                                      `gorm:"notnull"`
}

type SeerAssistantSessionMessageFunctionResponsePartModel struct {
	gorm.Model

	SeerAssistantSessionMessageFunctionResponsePartID uint                                            `gorm:"notnull"`
	SeerAssistantSessionMessageFunctionResponsePart   SeerAssistantSessionMessageFunctionResponsePart `gorm:"notnull"`
}

type SeerAssistantSessionMessageFunctionResponsePartType interface {
	GenaiType() string
	FunctionResponsePart(context.Context) (*genai.FunctionResponsePart, error)
	IsFunctionResponsePartType(*genai.FunctionResponsePart) bool
	Save(context.Context, *genai.FunctionResponsePart) error
}

type seerAssistantSessionMessageFunctionResponsePartTypeWithPart interface {
	SeerAssistantSessionMessageFunctionResponsePartType
	withFunctionResponsePart(SeerAssistantSessionMessageFunctionResponsePart) SeerAssistantSessionMessageFunctionResponsePartType
}

// Maps between SeerAssistantSessionMessageFunctionResponsePart.Kind and model.
var SeerAssistantSessionMessageFunctionResponsePartTypes = registry.NewRegistry[SeerAssistantSessionMessageFunctionResponsePartType]()

func RegisterFunctionResponsePartType[T SeerAssistantSessionMessageFunctionResponsePartType]() {
	var zero T
	SeerAssistantSessionMessageFunctionResponsePartTypes.Register(zero.GenaiType(), zero)
}

func loadFunctionResponsePartTypeModel(db *gorm.DB, kind string, partID uint) (SeerAssistantSessionMessageFunctionResponsePartType, error) {
	partTypeModel, isTypeKnown := SeerAssistantSessionMessageFunctionResponsePartTypes.Get(kind)
	if !isTypeKnown {
		return nil, fmt.Errorf("unknown kind of function response part type: %q", kind)
	}
	partTypeModelValue := reflect.New(reflect.TypeOf(partTypeModel))
	err := db.Preload("SeerAssistantSessionMessageFunctionResponsePart").Where("seer_assistant_session_message_function_response_part_id = ?", partID).First(partTypeModelValue.Interface()).Error
	if err != nil {
		return nil, err
	}
	loadedPartTypeModel, ok := partTypeModelValue.Elem().Interface().(SeerAssistantSessionMessageFunctionResponsePartType)
	if !ok {
		return nil, fmt.Errorf("loaded function response part type %q has wrong type", kind)
	}
	return loadedPartTypeModel, nil
}

func (m SeerAssistantSessionMessageFunctionResponsePart) SaveFunctionResponsePartType(ctx context.Context, part *genai.FunctionResponsePart) error {
	partType, isPartTypeKnown := SeerAssistantSessionMessageFunctionResponsePartTypes.Get(m.Kind)
	if !isPartTypeKnown {
		return fmt.Errorf("function response part type is unknown")
	}
	partTypeWithModel, ok := partType.(seerAssistantSessionMessageFunctionResponsePartTypeWithPart)
	if !ok {
		return fmt.Errorf("function response part type %q cannot bind function response part", m.Kind)
	}
	return partTypeWithModel.withFunctionResponsePart(m).Save(ctx, part)
}

type SeerAssistantSessionMessageFunctionResponseBlob struct {
	SeerAssistantSessionMessageFunctionResponsePartModel

	MIMEType    string `gorm:"notnull"`
	Data        []byte `gorm:"notnull"`
	DisplayName string
}

func (m SeerAssistantSessionMessageFunctionResponseBlob) GenaiType() string {
	return "inlineData"
}

func (m SeerAssistantSessionMessageFunctionResponseBlob) withFunctionResponsePart(part SeerAssistantSessionMessageFunctionResponsePart) SeerAssistantSessionMessageFunctionResponsePartType {
	m.SeerAssistantSessionMessageFunctionResponsePartModel = SeerAssistantSessionMessageFunctionResponsePartModel{
		SeerAssistantSessionMessageFunctionResponsePartID: part.ID,
		SeerAssistantSessionMessageFunctionResponsePart:   part,
	}
	return m
}

func (m SeerAssistantSessionMessageFunctionResponseBlob) IsFunctionResponsePartType(part *genai.FunctionResponsePart) bool {
	return part != nil && part.InlineData != nil
}

func (m SeerAssistantSessionMessageFunctionResponseBlob) FunctionResponsePart(_ context.Context) (*genai.FunctionResponsePart, error) {
	return &genai.FunctionResponsePart{
		InlineData: &genai.FunctionResponseBlob{
			MIMEType:    m.MIMEType,
			Data:        m.Data,
			DisplayName: m.DisplayName,
		},
	}, nil
}

func (m SeerAssistantSessionMessageFunctionResponseBlob) Save(ctx context.Context, part *genai.FunctionResponsePart) error {
	if !m.IsFunctionResponsePartType(part) {
		return fmt.Errorf("function response part is not inlineData")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	m.MIMEType = part.InlineData.MIMEType
	m.Data = part.InlineData.Data
	m.DisplayName = part.InlineData.DisplayName
	return gorm.G[SeerAssistantSessionMessageFunctionResponseBlob](db).Create(ctx, &m)
}

type SeerAssistantSessionMessageFunctionResponseFileData struct {
	SeerAssistantSessionMessageFunctionResponsePartModel

	FileURI     string `gorm:"notnull"`
	MIMEType    string `gorm:"notnull"`
	DisplayName string
}

func (m SeerAssistantSessionMessageFunctionResponseFileData) GenaiType() string {
	return "fileData"
}

func (m SeerAssistantSessionMessageFunctionResponseFileData) withFunctionResponsePart(part SeerAssistantSessionMessageFunctionResponsePart) SeerAssistantSessionMessageFunctionResponsePartType {
	m.SeerAssistantSessionMessageFunctionResponsePartModel = SeerAssistantSessionMessageFunctionResponsePartModel{
		SeerAssistantSessionMessageFunctionResponsePartID: part.ID,
		SeerAssistantSessionMessageFunctionResponsePart:   part,
	}
	return m
}

func (m SeerAssistantSessionMessageFunctionResponseFileData) IsFunctionResponsePartType(part *genai.FunctionResponsePart) bool {
	return part != nil && part.FileData != nil
}

func (m SeerAssistantSessionMessageFunctionResponseFileData) FunctionResponsePart(_ context.Context) (*genai.FunctionResponsePart, error) {
	return &genai.FunctionResponsePart{
		FileData: &genai.FunctionResponseFileData{
			FileURI:     m.FileURI,
			MIMEType:    m.MIMEType,
			DisplayName: m.DisplayName,
		},
	}, nil
}

func (m SeerAssistantSessionMessageFunctionResponseFileData) Save(ctx context.Context, part *genai.FunctionResponsePart) error {
	if !m.IsFunctionResponsePartType(part) {
		return fmt.Errorf("function response part is not fileData")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	m.FileURI = part.FileData.FileURI
	m.MIMEType = part.FileData.MIMEType
	m.DisplayName = part.FileData.DisplayName
	return gorm.G[SeerAssistantSessionMessageFunctionResponseFileData](db).Create(ctx, &m)
}

type SeerAssistantSessionMessageText struct {
	SeerAssistantSessionMessagePartModel

	Text string `gorm:"notnull"`
}

func (m SeerAssistantSessionMessageText) GenaiType() string {
	return "text"
}

func (m SeerAssistantSessionMessageText) withMessagePart(part SeerAssistantSessionMessagePart) SeerAssistantSessionMessageType {
	m.SeerAssistantSessionMessagePartModel = SeerAssistantSessionMessagePartModel{
		SeerAssistantSessionMessagePartID: part.ID,
		SeerAssistantSessionMessagePart:   part,
	}
	return m
}

func (m SeerAssistantSessionMessageText) IsPartType(part *genai.Part) bool {
	return part != nil && part.Text != ""
}

func (m SeerAssistantSessionMessageText) Part(_ context.Context) (*genai.Part, error) {
	return m.ApplyToPart(&genai.Part{Text: m.Text})
}

func (m SeerAssistantSessionMessageText) Save(ctx context.Context, part *genai.Part) error {
	if part == nil {
		return fmt.Errorf("part is nil")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	m.Text = part.Text
	return gorm.G[SeerAssistantSessionMessageText](db).Create(ctx, &m)
}

type SeerAssistantSessionMessageMediaResolution struct {
	SeerAssistantSessionMessagePartModel

	Level     string `gorm:"notnull"`
	NumTokens *int32
}

func (m SeerAssistantSessionMessageMediaResolution) GenaiType() string {
	return "mediaResolution"
}

func (m SeerAssistantSessionMessageMediaResolution) withMessagePart(part SeerAssistantSessionMessagePart) SeerAssistantSessionMessageType {
	m.SeerAssistantSessionMessagePartModel = SeerAssistantSessionMessagePartModel{
		SeerAssistantSessionMessagePartID: part.ID,
		SeerAssistantSessionMessagePart:   part,
	}
	return m
}

func (m SeerAssistantSessionMessageMediaResolution) IsPartType(part *genai.Part) bool {
	return part != nil && part.MediaResolution != nil
}

func (m SeerAssistantSessionMessageMediaResolution) Part(_ context.Context) (*genai.Part, error) {
	return m.ApplyToPart(&genai.Part{
		MediaResolution: &genai.PartMediaResolution{
			Level:     genai.PartMediaResolutionLevel(m.Level),
			NumTokens: m.NumTokens,
		},
	})
}

func (m SeerAssistantSessionMessageMediaResolution) Save(ctx context.Context, part *genai.Part) error {
	if !m.IsPartType(part) {
		return fmt.Errorf("part is not mediaResolution")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	m.Level = string(part.MediaResolution.Level)
	m.NumTokens = part.MediaResolution.NumTokens
	return gorm.G[SeerAssistantSessionMessageMediaResolution](db).Create(ctx, &m)
}

type SeerAssistantSessionMessageCodeExecutionResult struct {
	SeerAssistantSessionMessagePartModel

	Outcome          string  `gorm:"notnull"`
	Output           *string `gorm:"notnull"`
	ExecutableCodeID *string
}

func (m SeerAssistantSessionMessageCodeExecutionResult) GenaiType() string {
	return "codeExecutionResult"
}

func (m SeerAssistantSessionMessageCodeExecutionResult) withMessagePart(part SeerAssistantSessionMessagePart) SeerAssistantSessionMessageType {
	m.SeerAssistantSessionMessagePartModel = SeerAssistantSessionMessagePartModel{
		SeerAssistantSessionMessagePartID: part.ID,
		SeerAssistantSessionMessagePart:   part,
	}
	return m
}

func (m SeerAssistantSessionMessageCodeExecutionResult) IsPartType(part *genai.Part) bool {
	return part != nil && part.CodeExecutionResult != nil
}

func (m SeerAssistantSessionMessageCodeExecutionResult) Part(_ context.Context) (*genai.Part, error) {
	out := ""
	if m.Output != nil {
		out = *m.Output
	}
	id := ""
	if m.ExecutableCodeID != nil {
		id = *m.ExecutableCodeID
	}

	return m.ApplyToPart(&genai.Part{
		CodeExecutionResult: &genai.CodeExecutionResult{
			Outcome: genai.Outcome(m.Outcome),
			Output:  out,
			ID:      id,
		},
	})
}

func (m SeerAssistantSessionMessageCodeExecutionResult) Save(ctx context.Context, part *genai.Part) error {
	if !m.IsPartType(part) {
		return fmt.Errorf("part is not codeExecutionResult")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	output := part.CodeExecutionResult.Output
	m.Outcome = string(part.CodeExecutionResult.Outcome)
	m.Output = &output
	if part.CodeExecutionResult.ID != "" {
		id := part.CodeExecutionResult.ID
		m.ExecutableCodeID = &id
	}
	return gorm.G[SeerAssistantSessionMessageCodeExecutionResult](db).Create(ctx, &m)
}

type SeerAssistantSessionExecutableCode struct {
	SeerAssistantSessionMessagePartModel

	Code             string
	Language         string
	ExecutableCodeID string `gorm:"index"`
}

func (m SeerAssistantSessionExecutableCode) GenaiType() string {
	return "executableCode"
}

func (m SeerAssistantSessionExecutableCode) withMessagePart(part SeerAssistantSessionMessagePart) SeerAssistantSessionMessageType {
	m.SeerAssistantSessionMessagePartModel = SeerAssistantSessionMessagePartModel{
		SeerAssistantSessionMessagePartID: part.ID,
		SeerAssistantSessionMessagePart:   part,
	}
	return m
}

func (m SeerAssistantSessionExecutableCode) IsPartType(part *genai.Part) bool {
	return part != nil && part.ExecutableCode != nil
}

func (m SeerAssistantSessionExecutableCode) Part(_ context.Context) (*genai.Part, error) {
	return m.ApplyToPart(&genai.Part{
		ExecutableCode: &genai.ExecutableCode{
			Code:     m.Code,
			Language: genai.Language(m.Language),
			ID:       m.ExecutableCodeID,
		},
	})
}

func (m SeerAssistantSessionExecutableCode) Save(ctx context.Context, part *genai.Part) error {
	if !m.IsPartType(part) {
		return fmt.Errorf("part is not executableCode")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	m.Code = part.ExecutableCode.Code
	m.Language = string(part.ExecutableCode.Language)
	m.ExecutableCodeID = part.ExecutableCode.ID
	return gorm.G[SeerAssistantSessionExecutableCode](db).Create(ctx, &m)
}

type SeerAssistantSessionMessageToolCall struct {
	SeerAssistantSessionMessagePartModel

	ToolCallID *string `gorm:"index"`
	ToolType   *string
	Args       *datatypes.JSON
}

func (m SeerAssistantSessionMessageToolCall) GenaiType() string {
	return "toolCall"
}

func (m SeerAssistantSessionMessageToolCall) withMessagePart(part SeerAssistantSessionMessagePart) SeerAssistantSessionMessageType {
	m.SeerAssistantSessionMessagePartModel = SeerAssistantSessionMessagePartModel{
		SeerAssistantSessionMessagePartID: part.ID,
		SeerAssistantSessionMessagePart:   part,
	}
	return m
}

func (m SeerAssistantSessionMessageToolCall) IsPartType(part *genai.Part) bool {
	return part != nil && part.ToolCall != nil
}

func (m SeerAssistantSessionMessageToolCall) Part(_ context.Context) (*genai.Part, error) {
	var args map[string]any
	if m.Args != nil && len(*m.Args) > 0 {
		if err := json.Unmarshal(*m.Args, &args); err != nil {
			return nil, err
		}
	}
	id := ""
	if m.ToolCallID != nil {
		id = *m.ToolCallID
	}
	var tt genai.ToolType
	if m.ToolType != nil {
		tt = genai.ToolType(*m.ToolType)
	}
	return m.ApplyToPart(&genai.Part{
		ToolCall: &genai.ToolCall{
			ID:       id,
			ToolType: tt,
			Args:     args,
		},
	})
}

func (m SeerAssistantSessionMessageToolCall) Save(ctx context.Context, part *genai.Part) error {
	if !m.IsPartType(part) {
		return fmt.Errorf("part is not toolCall")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	if part.ToolCall.ID != "" {
		id := part.ToolCall.ID
		m.ToolCallID = &id
	}
	if part.ToolCall.ToolType != "" {
		toolType := string(part.ToolCall.ToolType)
		m.ToolType = &toolType
	}
	if part.ToolCall.Args != nil {
		argsJSON, err := json.Marshal(part.ToolCall.Args)
		if err != nil {
			return err
		}
		args := datatypes.JSON(argsJSON)
		m.Args = &args
	}
	return gorm.G[SeerAssistantSessionMessageToolCall](db).Create(ctx, &m)
}

type SeerAssistantSessionMessageFileData struct {
	SeerAssistantSessionMessagePartModel

	DisplayName *string
	FileURI     string `gorm:"notnull"`
	MIMEType    string `gorm:"notnull"`
}

func (m SeerAssistantSessionMessageFileData) GenaiType() string {
	return "fileData"
}

func (m SeerAssistantSessionMessageFileData) withMessagePart(part SeerAssistantSessionMessagePart) SeerAssistantSessionMessageType {
	m.SeerAssistantSessionMessagePartModel = SeerAssistantSessionMessagePartModel{
		SeerAssistantSessionMessagePartID: part.ID,
		SeerAssistantSessionMessagePart:   part,
	}
	return m
}

func (m SeerAssistantSessionMessageFileData) IsPartType(part *genai.Part) bool {
	return part != nil && part.FileData != nil
}

func (m SeerAssistantSessionMessageFileData) Part(_ context.Context) (*genai.Part, error) {
	display := ""
	if m.DisplayName != nil {
		display = *m.DisplayName
	}
	return m.ApplyToPart(&genai.Part{
		FileData: &genai.FileData{
			DisplayName: display,
			FileURI:     m.FileURI,
			MIMEType:    m.MIMEType,
		},
	})
}

func (m SeerAssistantSessionMessageFileData) Save(ctx context.Context, part *genai.Part) error {
	if !m.IsPartType(part) {
		return fmt.Errorf("part is not fileData")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	if part.FileData.DisplayName != "" {
		displayName := part.FileData.DisplayName
		m.DisplayName = &displayName
	}
	m.FileURI = part.FileData.FileURI
	m.MIMEType = part.FileData.MIMEType
	return gorm.G[SeerAssistantSessionMessageFileData](db).Create(ctx, &m)
}

type SeerAssistantSessionMessageFunctionCall struct {
	SeerAssistantSessionMessagePartModel

	FunctionCallID string `gorm:"index"`
	Args           datatypes.JSON
	Name           string
	WillContinue   *bool
}

func (m SeerAssistantSessionMessageFunctionCall) GenaiType() string {
	return "functionCall"
}

func (m SeerAssistantSessionMessageFunctionCall) withMessagePart(part SeerAssistantSessionMessagePart) SeerAssistantSessionMessageType {
	m.SeerAssistantSessionMessagePartModel = SeerAssistantSessionMessagePartModel{
		SeerAssistantSessionMessagePartID: part.ID,
		SeerAssistantSessionMessagePart:   part,
	}
	return m
}

func (m SeerAssistantSessionMessageFunctionCall) IsPartType(part *genai.Part) bool {
	return part != nil && part.FunctionCall != nil
}

func (m SeerAssistantSessionMessageFunctionCall) Part(_ context.Context) (*genai.Part, error) {
	var args map[string]any
	if len(m.Args) > 0 {
		if err := json.Unmarshal(m.Args, &args); err != nil {
			return nil, err
		}
	}
	return m.ApplyToPart(&genai.Part{
		FunctionCall: &genai.FunctionCall{
			ID:           m.FunctionCallID,
			Name:         m.Name,
			Args:         args,
			WillContinue: m.WillContinue,
		},
	})
}

func (m SeerAssistantSessionMessageFunctionCall) Save(ctx context.Context, part *genai.Part) error {
	if !m.IsPartType(part) {
		return fmt.Errorf("part is not functionCall")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	if part.FunctionCall.Args != nil {
		argsJSON, err := json.Marshal(part.FunctionCall.Args)
		if err != nil {
			return err
		}
		m.Args = datatypes.JSON(argsJSON)
	}
	m.FunctionCallID = part.FunctionCall.ID
	m.Name = part.FunctionCall.Name
	m.WillContinue = part.FunctionCall.WillContinue
	return gorm.G[SeerAssistantSessionMessageFunctionCall](db).Create(ctx, &m)
}

type SeerAssistantSessionMessageToolResponse struct {
	SeerAssistantSessionMessagePartModel

	ToolCallID *string `gorm:"index"`
	ToolType   *string
	Response   *datatypes.JSON
}

func (m SeerAssistantSessionMessageToolResponse) GenaiType() string {
	return "toolResponse"
}

func (m SeerAssistantSessionMessageToolResponse) withMessagePart(part SeerAssistantSessionMessagePart) SeerAssistantSessionMessageType {
	m.SeerAssistantSessionMessagePartModel = SeerAssistantSessionMessagePartModel{
		SeerAssistantSessionMessagePartID: part.ID,
		SeerAssistantSessionMessagePart:   part,
	}
	return m
}

func (m SeerAssistantSessionMessageToolResponse) IsPartType(part *genai.Part) bool {
	return part != nil && part.ToolResponse != nil
}

func (m SeerAssistantSessionMessageToolResponse) Part(_ context.Context) (*genai.Part, error) {
	var response map[string]any
	if m.Response != nil && len(*m.Response) > 0 {
		if err := json.Unmarshal(*m.Response, &response); err != nil {
			return nil, err
		}
	}
	id := ""
	if m.ToolCallID != nil {
		id = *m.ToolCallID
	}
	var tt genai.ToolType
	if m.ToolType != nil {
		tt = genai.ToolType(*m.ToolType)
	}
	return m.ApplyToPart(&genai.Part{
		ToolResponse: &genai.ToolResponse{
			ID:       id,
			ToolType: tt,
			Response: response,
		},
	})
}

func (m SeerAssistantSessionMessageToolResponse) Save(ctx context.Context, part *genai.Part) error {
	if !m.IsPartType(part) {
		return fmt.Errorf("part is not toolResponse")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	if part.ToolResponse.ID != "" {
		id := part.ToolResponse.ID
		m.ToolCallID = &id
	}
	if part.ToolResponse.ToolType != "" {
		toolType := string(part.ToolResponse.ToolType)
		m.ToolType = &toolType
	}
	if part.ToolResponse.Response != nil {
		responseJSON, err := json.Marshal(part.ToolResponse.Response)
		if err != nil {
			return err
		}
		response := datatypes.JSON(responseJSON)
		m.Response = &response
	}
	return gorm.G[SeerAssistantSessionMessageToolResponse](db).Create(ctx, &m)
}

func init() {
	RegisterMessageType[SeerAssistantSessionMessageInlineData]()
	RegisterMessageType[SeerAssistantSessionMessageFunctionResponse]()
	RegisterMessageType[SeerAssistantSessionMessageText]()
	RegisterMessageType[SeerAssistantSessionMessageMediaResolution]()
	RegisterMessageType[SeerAssistantSessionMessageCodeExecutionResult]()
	RegisterMessageType[SeerAssistantSessionExecutableCode]()
	RegisterMessageType[SeerAssistantSessionMessageToolCall]()
	RegisterMessageType[SeerAssistantSessionMessageFileData]()
	RegisterMessageType[SeerAssistantSessionMessageFunctionCall]()
	RegisterMessageType[SeerAssistantSessionMessageToolResponse]()

	RegisterFunctionResponsePartType[SeerAssistantSessionMessageFunctionResponseBlob]()
	RegisterFunctionResponsePartType[SeerAssistantSessionMessageFunctionResponseFileData]()

	lago.OnDBInit("p_seer_assistant.models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[SeerAssistantSession](db)
		lago.RegisterModel[SeerAssistantSessionMessage](db)
		lago.RegisterModel[SeerAssistantSessionMessagePart](db)
		lago.RegisterModel[VideoMetadata](db)
		lago.RegisterModel[SeerAssistantSessionExecutableCode](db)
		lago.RegisterModel[SeerAssistantSessionMessageInlineData](db)
		lago.RegisterModel[SeerAssistantSessionMessageFunctionResponse](db)
		lago.RegisterModel[SeerAssistantSessionMessageFunctionResponsePart](db)
		lago.RegisterModel[SeerAssistantSessionMessageFunctionResponseBlob](db)
		lago.RegisterModel[SeerAssistantSessionMessageFunctionResponseFileData](db)
		lago.RegisterModel[SeerAssistantSessionMessageText](db)
		lago.RegisterModel[SeerAssistantSessionMessageMediaResolution](db)
		lago.RegisterModel[SeerAssistantSessionMessageCodeExecutionResult](db)
		lago.RegisterModel[SeerAssistantSessionMessageToolCall](db)
		lago.RegisterModel[SeerAssistantSessionMessageFileData](db)
		lago.RegisterModel[SeerAssistantSessionMessageFunctionCall](db)
		lago.RegisterModel[SeerAssistantSessionMessageToolResponse](db)
		return db
	})
}
