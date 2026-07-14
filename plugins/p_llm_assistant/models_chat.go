package p_llm_assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/registry"
	"google.golang.org/genai"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type LlmAssistantSession struct {
	gorm.Model

	Title  string `gorm:"notnull;default:''"`
	UserID uint   `gorm:"index"`
}

func (m LlmAssistantSession) SaveContent(ctx context.Context, content genai.Content) error {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	sessionMessage := LlmAssistantSessionMessage{
		LlmAssistantSessionID: m.ID,
		LlmAssistantSession:   m,
		Role:                  content.Role,
	}
	err = gorm.G[LlmAssistantSessionMessage](db).Create(ctx, &sessionMessage)
	if err != nil {
		return err
	}
	return sessionMessage.SaveParts(ctx, content.Parts)
}

type LlmAssistantSessionMessage struct {
	gorm.Model

	LlmAssistantSessionID uint                `gorm:"notnull"`
	LlmAssistantSession   LlmAssistantSession `gorm:"notnull"`
	Role                  string              `gorm:"notnull;default:'user'"`
}

// genaiPartIsEmpty is true for a non-nil Part whose only “content” would still be ignored
// by the API — e.g. streaming placeholders. These are stored as empty text parts.
func genaiPartIsEmpty(part *genai.Part) bool {
	if part == nil {
		return false
	}
	return part.MediaResolution == nil &&
		part.CodeExecutionResult == nil &&
		part.ExecutableCode == nil &&
		part.FileData == nil &&
		part.FunctionCall == nil &&
		part.FunctionResponse == nil &&
		part.InlineData == nil &&
		part.Text == "" &&
		!part.Thought &&
		len(part.ThoughtSignature) == 0 &&
		part.VideoMetadata == nil &&
		part.ToolCall == nil &&
		part.ToolResponse == nil &&
		len(part.PartMetadata) == 0
}

// genaiPartPassesChatValidateContent mirrors google.golang.org/genai.validateContent's
// per-part logic: non-empty Text or one of the payload pointers it checks. Thought,
// ThoughtSignature, ToolCall, MediaResolution, etc. do not count — so a row stored as
// kind "text" with empty Text but signatures must be rehydrated with placeholder Text
// or Chats curated history drops the model turn and breaks FC/FR ordering (API 400).
func genaiPartPassesChatValidateContent(p *genai.Part) bool {
	if p == nil {
		return false
	}
	if p.Text != "" {
		return true
	}
	return p.InlineData != nil ||
		p.FileData != nil ||
		p.FunctionCall != nil ||
		p.FunctionResponse != nil ||
		p.ExecutableCode != nil ||
		p.CodeExecutionResult != nil
}

// sanitizeContentPartsForGenaiChat ensures every part passes Chat.validateContent in
// google.golang.org/genai (Text or the six payload pointers it checks). ToolCall,
// ToolResponse, MediaResolution, VideoMetadata, Thought, ThoughtSignature alone do not
// count — without this, extractCuratedHistory can drop model turns and break FC/FR order.
func sanitizeContentPartsForGenaiChat(c *genai.Content) {
	if c == nil {
		return
	}
	for _, p := range c.Parts {
		if p == nil {
			continue
		}
		if !genaiPartPassesChatValidateContent(p) {
			p.Text = "\u200b"
		}
	}
}

func (m LlmAssistantSessionMessage) SaveParts(ctx context.Context, parts []*genai.Part) error {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	messageKinds := LlmAssistantSessionMessageTypes.All()
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
		messagePart := LlmAssistantSessionMessagePart{
			Kind:                         messageKind,
			LlmAssistantSessionMessageID: m.ID,
			LlmAssistantSessionMessage:   m,
			Thought:                      part.Thought,
			ThoughtSignature:             part.ThoughtSignature,
			VideoMetadata:                videoMetadata,
			VideoMetadataID:              videoMetadataID,
			PartMetadata:                 datatypes.JSON(partMetadata),
		}
		err = gorm.G[LlmAssistantSessionMessagePart](db).Create(ctx, &messagePart)
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

func (m LlmAssistantSessionMessage) LoadContent(ctx context.Context) (*genai.Content, error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}
	parts, err := gorm.G[LlmAssistantSessionMessagePart](db).Where("llm_assistant_session_message_id = ?", m.ID).Find(ctx)
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

type LlmAssistantSessionMessageType interface {
	GenaiType() string
	Part(context.Context) (*genai.Part, error)
	IsPartType(*genai.Part) bool
	Save(context.Context, *genai.Part) error
}

type llmAssistantSessionMessageTypeWithPart interface {
	LlmAssistantSessionMessageType
	withMessagePart(LlmAssistantSessionMessagePart) LlmAssistantSessionMessageType
}

// Maps between LlmAssistantSessionMessage.Kind and model.
var LlmAssistantSessionMessageTypes = registry.NewRegistry[LlmAssistantSessionMessageType]()

func RegisterMessageType[T LlmAssistantSessionMessageType]() {
	var zero T
	LlmAssistantSessionMessageTypes.Register(zero.GenaiType(), zero)
}

func loadMessageTypeModel(db *gorm.DB, kind string, partID uint) (LlmAssistantSessionMessageType, error) {
	partTypeModel, isTypeKnown := LlmAssistantSessionMessageTypes.Get(kind)
	if !isTypeKnown {
		return nil, fmt.Errorf("unknown kind of part type: %q", kind)
	}
	partTypeModelValue := reflect.New(reflect.TypeOf(partTypeModel))
	err := db.Preload("LlmAssistantSessionMessagePart").Where("llm_assistant_session_message_part_id = ?", partID).First(partTypeModelValue.Interface()).Error
	if err != nil {
		return nil, err
	}
	loadedPartTypeModel, ok := partTypeModelValue.Elem().Interface().(LlmAssistantSessionMessageType)
	if !ok {
		return nil, fmt.Errorf("loaded message part type %q has wrong type", kind)
	}
	return loadedPartTypeModel, nil
}

type LlmAssistantSessionMessagePart struct {
	gorm.Model

	Kind                         string                     `gorm:"notnull"`
	LlmAssistantSessionMessageID uint                       `gorm:"notnull"`
	LlmAssistantSessionMessage   LlmAssistantSessionMessage `gorm:"notnull"`
	Thought                      bool                       `gorm:"notnull;default:false"`
	ThoughtSignature             []byte
	VideoMetadataID              *uint
	VideoMetadata                *VideoMetadata
	PartMetadata                 datatypes.JSON
}

func (m LlmAssistantSessionMessagePart) SavePartType(ctx context.Context, part *genai.Part) error {
	partType, isPartTypeKnown := LlmAssistantSessionMessageTypes.Get(m.Kind)
	if !isPartTypeKnown {
		return fmt.Errorf("part type is unknown")
	}
	partTypeWithModel, ok := partType.(llmAssistantSessionMessageTypeWithPart)
	if !ok {
		return fmt.Errorf("part type %q cannot bind message part", m.Kind)
	}
	return partTypeWithModel.withMessagePart(m).Save(ctx, part)
}

type LlmAssistantSessionMessagePartModel struct {
	gorm.Model

	LlmAssistantSessionMessagePartID uint                           `gorm:"notnull"`
	LlmAssistantSessionMessagePart   LlmAssistantSessionMessagePart `gorm:"notnull"`
}

func (m LlmAssistantSessionMessagePart) ApplyToPart(part *genai.Part) (*genai.Part, error) {
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

func (m LlmAssistantSessionMessagePartModel) ApplyToPart(part *genai.Part) (*genai.Part, error) {
	return m.LlmAssistantSessionMessagePart.ApplyToPart(part)
}

type VideoMetadata struct {
	gorm.Model
	EndOffset   time.Duration
	FPS         *float64
	StartOffset time.Duration
}

type LlmAssistantSessionMessageInlineData struct {
	LlmAssistantSessionMessagePartModel

	MIMEType    string `gorm:"notnull"`
	Data        []byte `gorm:"notnull"`
	DisplayName string
}

func (m LlmAssistantSessionMessageInlineData) GenaiType() string {
	return "inlineData"
}

func (m LlmAssistantSessionMessageInlineData) withMessagePart(part LlmAssistantSessionMessagePart) LlmAssistantSessionMessageType {
	m.LlmAssistantSessionMessagePartModel = LlmAssistantSessionMessagePartModel{
		LlmAssistantSessionMessagePartID: part.ID,
		LlmAssistantSessionMessagePart:   part,
	}
	return m
}

func (m LlmAssistantSessionMessageInlineData) IsPartType(part *genai.Part) bool {
	return part != nil && part.InlineData != nil
}

func (m LlmAssistantSessionMessageInlineData) Part(_ context.Context) (*genai.Part, error) {
	return m.ApplyToPart(&genai.Part{
		InlineData: &genai.Blob{
			Data:        m.Data,
			MIMEType:    m.MIMEType,
			DisplayName: m.DisplayName,
		},
	})
}

func (m LlmAssistantSessionMessageInlineData) Save(ctx context.Context, part *genai.Part) error {
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
	return gorm.G[LlmAssistantSessionMessageInlineData](db).Create(ctx, &m)
}

type LlmAssistantSessionMessageFunctionResponse struct {
	LlmAssistantSessionMessagePartModel

	WillContinue       *bool
	Scheduling         string `gorm:"default:'WHEN_IDLE'"`
	FunctionResponseID string
	Name               string `gorm:"notnull"`
	Response           datatypes.JSON
}

func (m LlmAssistantSessionMessageFunctionResponse) GenaiType() string {
	return "functionResponse"
}

func (m LlmAssistantSessionMessageFunctionResponse) withMessagePart(part LlmAssistantSessionMessagePart) LlmAssistantSessionMessageType {
	m.LlmAssistantSessionMessagePartModel = LlmAssistantSessionMessagePartModel{
		LlmAssistantSessionMessagePartID: part.ID,
		LlmAssistantSessionMessagePart:   part,
	}
	return m
}

func (m LlmAssistantSessionMessageFunctionResponse) IsPartType(part *genai.Part) bool {
	return part != nil && part.FunctionResponse != nil
}

func (m LlmAssistantSessionMessageFunctionResponse) Part(ctx context.Context) (*genai.Part, error) {
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return nil, err
	}
	parts, err := gorm.G[LlmAssistantSessionMessageFunctionResponsePart](db).Where("llm_assistant_session_message_function_response_id = ?", m.ID).Find(ctx)
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

func (m LlmAssistantSessionMessageFunctionResponse) Save(ctx context.Context, part *genai.Part) error {
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
	err = gorm.G[LlmAssistantSessionMessageFunctionResponse](db).Create(ctx, &m)
	if err != nil {
		return err
	}
	functionResponsePartKinds := LlmAssistantSessionMessageFunctionResponsePartTypes.All()
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
		messageFunctionResponsePart := LlmAssistantSessionMessageFunctionResponsePart{
			LlmAssistantSessionMessageFunctionResponseID: m.ID,
			LlmAssistantSessionMessageFunctionResponse:   m,
			Kind: functionResponsePartKind,
		}
		err = gorm.G[LlmAssistantSessionMessageFunctionResponsePart](db).Create(ctx, &messageFunctionResponsePart)
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

type LlmAssistantSessionMessageFunctionResponsePart struct {
	gorm.Model

	LlmAssistantSessionMessageFunctionResponseID uint                                       `gorm:"notnull"`
	LlmAssistantSessionMessageFunctionResponse   LlmAssistantSessionMessageFunctionResponse `gorm:"notnull"`
	Kind                                         string                                     `gorm:"notnull"`
}

type LlmAssistantSessionMessageFunctionResponsePartModel struct {
	gorm.Model

	LlmAssistantSessionMessageFunctionResponsePartID uint                                           `gorm:"notnull"`
	LlmAssistantSessionMessageFunctionResponsePart   LlmAssistantSessionMessageFunctionResponsePart `gorm:"notnull"`
}

type LlmAssistantSessionMessageFunctionResponsePartType interface {
	GenaiType() string
	FunctionResponsePart(context.Context) (*genai.FunctionResponsePart, error)
	IsFunctionResponsePartType(*genai.FunctionResponsePart) bool
	Save(context.Context, *genai.FunctionResponsePart) error
}

type llmAssistantSessionMessageFunctionResponsePartTypeWithPart interface {
	LlmAssistantSessionMessageFunctionResponsePartType
	withFunctionResponsePart(LlmAssistantSessionMessageFunctionResponsePart) LlmAssistantSessionMessageFunctionResponsePartType
}

// Maps between LlmAssistantSessionMessageFunctionResponsePart.Kind and model.
var LlmAssistantSessionMessageFunctionResponsePartTypes = registry.NewRegistry[LlmAssistantSessionMessageFunctionResponsePartType]()

func RegisterFunctionResponsePartType[T LlmAssistantSessionMessageFunctionResponsePartType]() {
	var zero T
	LlmAssistantSessionMessageFunctionResponsePartTypes.Register(zero.GenaiType(), zero)
}

func loadFunctionResponsePartTypeModel(db *gorm.DB, kind string, partID uint) (LlmAssistantSessionMessageFunctionResponsePartType, error) {
	partTypeModel, isTypeKnown := LlmAssistantSessionMessageFunctionResponsePartTypes.Get(kind)
	if !isTypeKnown {
		return nil, fmt.Errorf("unknown kind of function response part type: %q", kind)
	}
	partTypeModelValue := reflect.New(reflect.TypeOf(partTypeModel))
	err := db.Preload("LlmAssistantSessionMessageFunctionResponsePart").Where("llm_assistant_session_message_function_response_part_id = ?", partID).First(partTypeModelValue.Interface()).Error
	if err != nil {
		return nil, err
	}
	loadedPartTypeModel, ok := partTypeModelValue.Elem().Interface().(LlmAssistantSessionMessageFunctionResponsePartType)
	if !ok {
		return nil, fmt.Errorf("loaded function response part type %q has wrong type", kind)
	}
	return loadedPartTypeModel, nil
}

func (m LlmAssistantSessionMessageFunctionResponsePart) SaveFunctionResponsePartType(ctx context.Context, part *genai.FunctionResponsePart) error {
	partType, isPartTypeKnown := LlmAssistantSessionMessageFunctionResponsePartTypes.Get(m.Kind)
	if !isPartTypeKnown {
		return fmt.Errorf("function response part type is unknown")
	}
	partTypeWithModel, ok := partType.(llmAssistantSessionMessageFunctionResponsePartTypeWithPart)
	if !ok {
		return fmt.Errorf("function response part type %q cannot bind function response part", m.Kind)
	}
	return partTypeWithModel.withFunctionResponsePart(m).Save(ctx, part)
}

type LlmAssistantSessionMessageFunctionResponseBlob struct {
	LlmAssistantSessionMessageFunctionResponsePartModel

	MIMEType    string `gorm:"notnull"`
	Data        []byte `gorm:"notnull"`
	DisplayName string
}

func (m LlmAssistantSessionMessageFunctionResponseBlob) GenaiType() string {
	return "inlineData"
}

func (m LlmAssistantSessionMessageFunctionResponseBlob) withFunctionResponsePart(part LlmAssistantSessionMessageFunctionResponsePart) LlmAssistantSessionMessageFunctionResponsePartType {
	m.LlmAssistantSessionMessageFunctionResponsePartModel = LlmAssistantSessionMessageFunctionResponsePartModel{
		LlmAssistantSessionMessageFunctionResponsePartID: part.ID,
		LlmAssistantSessionMessageFunctionResponsePart:   part,
	}
	return m
}

func (m LlmAssistantSessionMessageFunctionResponseBlob) IsFunctionResponsePartType(part *genai.FunctionResponsePart) bool {
	return part != nil && part.InlineData != nil
}

func (m LlmAssistantSessionMessageFunctionResponseBlob) FunctionResponsePart(_ context.Context) (*genai.FunctionResponsePart, error) {
	return &genai.FunctionResponsePart{
		InlineData: &genai.FunctionResponseBlob{
			MIMEType:    m.MIMEType,
			Data:        m.Data,
			DisplayName: m.DisplayName,
		},
	}, nil
}

func (m LlmAssistantSessionMessageFunctionResponseBlob) Save(ctx context.Context, part *genai.FunctionResponsePart) error {
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
	return gorm.G[LlmAssistantSessionMessageFunctionResponseBlob](db).Create(ctx, &m)
}

type LlmAssistantSessionMessageFunctionResponseFileData struct {
	LlmAssistantSessionMessageFunctionResponsePartModel

	FileURI     string `gorm:"notnull"`
	MIMEType    string `gorm:"notnull"`
	DisplayName string
}

func (m LlmAssistantSessionMessageFunctionResponseFileData) GenaiType() string {
	return "fileData"
}

func (m LlmAssistantSessionMessageFunctionResponseFileData) withFunctionResponsePart(part LlmAssistantSessionMessageFunctionResponsePart) LlmAssistantSessionMessageFunctionResponsePartType {
	m.LlmAssistantSessionMessageFunctionResponsePartModel = LlmAssistantSessionMessageFunctionResponsePartModel{
		LlmAssistantSessionMessageFunctionResponsePartID: part.ID,
		LlmAssistantSessionMessageFunctionResponsePart:   part,
	}
	return m
}

func (m LlmAssistantSessionMessageFunctionResponseFileData) IsFunctionResponsePartType(part *genai.FunctionResponsePart) bool {
	return part != nil && part.FileData != nil
}

func (m LlmAssistantSessionMessageFunctionResponseFileData) FunctionResponsePart(_ context.Context) (*genai.FunctionResponsePart, error) {
	return &genai.FunctionResponsePart{
		FileData: &genai.FunctionResponseFileData{
			FileURI:     m.FileURI,
			MIMEType:    m.MIMEType,
			DisplayName: m.DisplayName,
		},
	}, nil
}

func (m LlmAssistantSessionMessageFunctionResponseFileData) Save(ctx context.Context, part *genai.FunctionResponsePart) error {
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
	return gorm.G[LlmAssistantSessionMessageFunctionResponseFileData](db).Create(ctx, &m)
}

type LlmAssistantSessionMessageText struct {
	LlmAssistantSessionMessagePartModel

	Text string `gorm:"notnull"`
}

func (m LlmAssistantSessionMessageText) GenaiType() string {
	return "text"
}

func (m LlmAssistantSessionMessageText) withMessagePart(part LlmAssistantSessionMessagePart) LlmAssistantSessionMessageType {
	m.LlmAssistantSessionMessagePartModel = LlmAssistantSessionMessagePartModel{
		LlmAssistantSessionMessagePartID: part.ID,
		LlmAssistantSessionMessagePart:   part,
	}
	return m
}

func (m LlmAssistantSessionMessageText) IsPartType(part *genai.Part) bool {
	if part == nil {
		return false
	}
	if part.InlineData != nil || part.FileData != nil ||
		part.FunctionCall != nil || part.FunctionResponse != nil ||
		part.CodeExecutionResult != nil || part.ExecutableCode != nil ||
		part.MediaResolution != nil || part.ToolCall != nil || part.ToolResponse != nil {
		return false
	}
	return part.Text != "" || part.Thought || len(part.ThoughtSignature) > 0 || genaiPartIsEmpty(part)
}

func (m LlmAssistantSessionMessageText) Part(_ context.Context) (*genai.Part, error) {
	part, err := m.ApplyToPart(&genai.Part{Text: m.Text})
	if err != nil {
		return nil, err
	}
	if !genaiPartPassesChatValidateContent(part) {
		part.Text = "\u200b"
	}
	return part, nil
}

func (m LlmAssistantSessionMessageText) Save(ctx context.Context, part *genai.Part) error {
	if part == nil {
		return fmt.Errorf("part is nil")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	m.Text = part.Text
	return gorm.G[LlmAssistantSessionMessageText](db).Create(ctx, &m)
}

type LlmAssistantSessionMessageMediaResolution struct {
	LlmAssistantSessionMessagePartModel

	Level     string `gorm:"notnull"`
	NumTokens *int32
}

func (m LlmAssistantSessionMessageMediaResolution) GenaiType() string {
	return "mediaResolution"
}

func (m LlmAssistantSessionMessageMediaResolution) withMessagePart(part LlmAssistantSessionMessagePart) LlmAssistantSessionMessageType {
	m.LlmAssistantSessionMessagePartModel = LlmAssistantSessionMessagePartModel{
		LlmAssistantSessionMessagePartID: part.ID,
		LlmAssistantSessionMessagePart:   part,
	}
	return m
}

func (m LlmAssistantSessionMessageMediaResolution) IsPartType(part *genai.Part) bool {
	return part != nil && part.MediaResolution != nil
}

func (m LlmAssistantSessionMessageMediaResolution) Part(_ context.Context) (*genai.Part, error) {
	return m.ApplyToPart(&genai.Part{
		MediaResolution: &genai.PartMediaResolution{
			Level:     genai.PartMediaResolutionLevel(m.Level),
			NumTokens: m.NumTokens,
		},
	})
}

func (m LlmAssistantSessionMessageMediaResolution) Save(ctx context.Context, part *genai.Part) error {
	if !m.IsPartType(part) {
		return fmt.Errorf("part is not mediaResolution")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return err
	}
	m.Level = string(part.MediaResolution.Level)
	m.NumTokens = part.MediaResolution.NumTokens
	return gorm.G[LlmAssistantSessionMessageMediaResolution](db).Create(ctx, &m)
}

type LlmAssistantSessionMessageCodeExecutionResult struct {
	LlmAssistantSessionMessagePartModel

	Outcome          string  `gorm:"notnull"`
	Output           *string `gorm:"notnull"`
	ExecutableCodeID *string
}

func (m LlmAssistantSessionMessageCodeExecutionResult) GenaiType() string {
	return "codeExecutionResult"
}

func (m LlmAssistantSessionMessageCodeExecutionResult) withMessagePart(part LlmAssistantSessionMessagePart) LlmAssistantSessionMessageType {
	m.LlmAssistantSessionMessagePartModel = LlmAssistantSessionMessagePartModel{
		LlmAssistantSessionMessagePartID: part.ID,
		LlmAssistantSessionMessagePart:   part,
	}
	return m
}

func (m LlmAssistantSessionMessageCodeExecutionResult) IsPartType(part *genai.Part) bool {
	return part != nil && part.CodeExecutionResult != nil
}

func (m LlmAssistantSessionMessageCodeExecutionResult) Part(_ context.Context) (*genai.Part, error) {
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

func (m LlmAssistantSessionMessageCodeExecutionResult) Save(ctx context.Context, part *genai.Part) error {
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
	return gorm.G[LlmAssistantSessionMessageCodeExecutionResult](db).Create(ctx, &m)
}

type LlmAssistantSessionExecutableCode struct {
	LlmAssistantSessionMessagePartModel

	Code             string
	Language         string
	ExecutableCodeID string `gorm:"index"`
}

func (m LlmAssistantSessionExecutableCode) GenaiType() string {
	return "executableCode"
}

func (m LlmAssistantSessionExecutableCode) withMessagePart(part LlmAssistantSessionMessagePart) LlmAssistantSessionMessageType {
	m.LlmAssistantSessionMessagePartModel = LlmAssistantSessionMessagePartModel{
		LlmAssistantSessionMessagePartID: part.ID,
		LlmAssistantSessionMessagePart:   part,
	}
	return m
}

func (m LlmAssistantSessionExecutableCode) IsPartType(part *genai.Part) bool {
	return part != nil && part.ExecutableCode != nil
}

func (m LlmAssistantSessionExecutableCode) Part(_ context.Context) (*genai.Part, error) {
	return m.ApplyToPart(&genai.Part{
		ExecutableCode: &genai.ExecutableCode{
			Code:     m.Code,
			Language: genai.Language(m.Language),
			ID:       m.ExecutableCodeID,
		},
	})
}

func (m LlmAssistantSessionExecutableCode) Save(ctx context.Context, part *genai.Part) error {
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
	return gorm.G[LlmAssistantSessionExecutableCode](db).Create(ctx, &m)
}

type LlmAssistantSessionMessageToolCall struct {
	LlmAssistantSessionMessagePartModel

	ToolCallID *string `gorm:"index"`
	ToolType   *string
	Args       *datatypes.JSON
}

func (m LlmAssistantSessionMessageToolCall) GenaiType() string {
	return "toolCall"
}

func (m LlmAssistantSessionMessageToolCall) withMessagePart(part LlmAssistantSessionMessagePart) LlmAssistantSessionMessageType {
	m.LlmAssistantSessionMessagePartModel = LlmAssistantSessionMessagePartModel{
		LlmAssistantSessionMessagePartID: part.ID,
		LlmAssistantSessionMessagePart:   part,
	}
	return m
}

func (m LlmAssistantSessionMessageToolCall) IsPartType(part *genai.Part) bool {
	return part != nil && part.ToolCall != nil
}

func (m LlmAssistantSessionMessageToolCall) Part(_ context.Context) (*genai.Part, error) {
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

func (m LlmAssistantSessionMessageToolCall) Save(ctx context.Context, part *genai.Part) error {
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
	return gorm.G[LlmAssistantSessionMessageToolCall](db).Create(ctx, &m)
}

type LlmAssistantSessionMessageFileData struct {
	LlmAssistantSessionMessagePartModel

	DisplayName *string
	FileURI     string `gorm:"notnull"`
	MIMEType    string `gorm:"notnull"`
}

func (m LlmAssistantSessionMessageFileData) GenaiType() string {
	return "fileData"
}

func (m LlmAssistantSessionMessageFileData) withMessagePart(part LlmAssistantSessionMessagePart) LlmAssistantSessionMessageType {
	m.LlmAssistantSessionMessagePartModel = LlmAssistantSessionMessagePartModel{
		LlmAssistantSessionMessagePartID: part.ID,
		LlmAssistantSessionMessagePart:   part,
	}
	return m
}

func (m LlmAssistantSessionMessageFileData) IsPartType(part *genai.Part) bool {
	return part != nil && part.FileData != nil
}

func (m LlmAssistantSessionMessageFileData) Part(_ context.Context) (*genai.Part, error) {
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

func (m LlmAssistantSessionMessageFileData) Save(ctx context.Context, part *genai.Part) error {
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
	return gorm.G[LlmAssistantSessionMessageFileData](db).Create(ctx, &m)
}

type LlmAssistantSessionMessageFunctionCall struct {
	LlmAssistantSessionMessagePartModel

	FunctionCallID string `gorm:"index"`
	Args           datatypes.JSON
	Name           string
	WillContinue   *bool
}

func (m LlmAssistantSessionMessageFunctionCall) GenaiType() string {
	return "functionCall"
}

func (m LlmAssistantSessionMessageFunctionCall) withMessagePart(part LlmAssistantSessionMessagePart) LlmAssistantSessionMessageType {
	m.LlmAssistantSessionMessagePartModel = LlmAssistantSessionMessagePartModel{
		LlmAssistantSessionMessagePartID: part.ID,
		LlmAssistantSessionMessagePart:   part,
	}
	return m
}

func (m LlmAssistantSessionMessageFunctionCall) IsPartType(part *genai.Part) bool {
	return part != nil && part.FunctionCall != nil
}

func (m LlmAssistantSessionMessageFunctionCall) Part(_ context.Context) (*genai.Part, error) {
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

func (m LlmAssistantSessionMessageFunctionCall) Save(ctx context.Context, part *genai.Part) error {
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
	return gorm.G[LlmAssistantSessionMessageFunctionCall](db).Create(ctx, &m)
}

type LlmAssistantSessionMessageToolResponse struct {
	LlmAssistantSessionMessagePartModel

	ToolCallID *string `gorm:"index"`
	ToolType   *string
	Response   *datatypes.JSON
}

func (m LlmAssistantSessionMessageToolResponse) GenaiType() string {
	return "toolResponse"
}

func (m LlmAssistantSessionMessageToolResponse) withMessagePart(part LlmAssistantSessionMessagePart) LlmAssistantSessionMessageType {
	m.LlmAssistantSessionMessagePartModel = LlmAssistantSessionMessagePartModel{
		LlmAssistantSessionMessagePartID: part.ID,
		LlmAssistantSessionMessagePart:   part,
	}
	return m
}

func (m LlmAssistantSessionMessageToolResponse) IsPartType(part *genai.Part) bool {
	return part != nil && part.ToolResponse != nil
}

func (m LlmAssistantSessionMessageToolResponse) Part(_ context.Context) (*genai.Part, error) {
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

func (m LlmAssistantSessionMessageToolResponse) Save(ctx context.Context, part *genai.Part) error {
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
	return gorm.G[LlmAssistantSessionMessageToolResponse](db).Create(ctx, &m)
}

type Skill struct {
	gorm.Model
	Name        string               `gorm:"unique;notnull;default:''"`
	Description string               `gorm:"notnull;default:''"`
	Content     string               `gorm:"notnull;default:''"`
	Files       []p_filesystem.VNode `gorm:"many2many:llm_assistant_skill_files;"`
}

func init() {
	RegisterMessageType[LlmAssistantSessionMessageInlineData]()
	RegisterMessageType[LlmAssistantSessionMessageFunctionResponse]()
	RegisterMessageType[LlmAssistantSessionMessageText]()
	RegisterMessageType[LlmAssistantSessionMessageMediaResolution]()
	RegisterMessageType[LlmAssistantSessionMessageCodeExecutionResult]()
	RegisterMessageType[LlmAssistantSessionExecutableCode]()
	RegisterMessageType[LlmAssistantSessionMessageToolCall]()
	RegisterMessageType[LlmAssistantSessionMessageFileData]()
	RegisterMessageType[LlmAssistantSessionMessageFunctionCall]()
	RegisterMessageType[LlmAssistantSessionMessageToolResponse]()

	RegisterFunctionResponsePartType[LlmAssistantSessionMessageFunctionResponseBlob]()
	RegisterFunctionResponsePartType[LlmAssistantSessionMessageFunctionResponseFileData]()

	registerPluginDBInitHook("p_llm_assistant.models", func(db *gorm.DB) *gorm.DB {
		if err := db.AutoMigrate(&LlmAssistantSession{}, &LlmAssistantSessionMessage{}, &LlmAssistantSessionMessagePart{}, &VideoMetadata{}, &LlmAssistantSessionExecutableCode{}, &LlmAssistantSessionMessageInlineData{}, &LlmAssistantSessionMessageFunctionResponse{}, &LlmAssistantSessionMessageFunctionResponsePart{}, &LlmAssistantSessionMessageFunctionResponseBlob{}, &LlmAssistantSessionMessageFunctionResponseFileData{}, &LlmAssistantSessionMessageText{}, &LlmAssistantSessionMessageMediaResolution{}, &LlmAssistantSessionMessageCodeExecutionResult{}, &LlmAssistantSessionMessageToolCall{}, &LlmAssistantSessionMessageFileData{}, &LlmAssistantSessionMessageFunctionCall{}, &LlmAssistantSessionMessageToolResponse{}, &Skill{}); err != nil {
			panic(err)
		}
		return db
	})
}
