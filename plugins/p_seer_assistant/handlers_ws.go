package p_seer_assistant

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/plugins/p_google_genai"
	"github.com/lariv-in/lago/plugins/p_users"
	"golang.org/x/net/websocket"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

type UserMessage struct {
	SessionID uint   `json:"session_id"`
	Message   string `json:"message"`
}

func (m UserMessage) ToHTML() string {
	return fmt.Sprintf(
		`<input id="seer_assistant_session_id" hx-swap-oob="true" type="hidden" name="session_id" value="%d"><div id="seer_assistant_transcript" hx-swap-oob="beforeend"><div class="chat chat-end mb-2"><div class="chat-header text-xs opacity-70">You</div><div class="chat-bubble chat-bubble-primary whitespace-pre-wrap">%s</div></div></div>`,
		m.SessionID,
		html.EscapeString(m.Message),
	)
}

func (m UserMessage) Save(r *http.Request) (UserMessage, error) {
	ctx := r.Context()
	user, ok := ctx.Value("$user").(p_users.User)
	if !ok {
		return UserMessage{}, fmt.Errorf("User is not present in the context")
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return UserMessage{}, err
	}
	session := SeerAssistantSession{
		UserID: user.ID,
	}
	if m.SessionID == 0 {
		if session, err = CreateSession(ctx, db, user.ID); err != nil {
			return UserMessage{}, err
		}
	} else {
		if session, err = gorm.G[SeerAssistantSession](db).Where("id = ?", m.SessionID).First(ctx); err != nil {
			return UserMessage{}, err
		}
	}
	m.SessionID = session.ID
	if user.ID != session.UserID {
		return UserMessage{}, fmt.Errorf("session belongs to another user")
	}
	if err = session.SaveContent(ctx, *genai.NewContentFromText(m.Message, genai.RoleUser)); err != nil {
		return UserMessage{}, err
	}
	return UserMessage{
		SessionID: session.ID,
		Message:   m.Message,
	}, nil
}

func unmarshal(msg []byte, payloadType byte, v interface{}) (err error) {
	return json.Unmarshal(msg, v)
}

func marshal(v interface{}) ([]byte, byte, error) {
	err, isErr := v.(error)
	if isErr {
		escaped := html.EscapeString(err.Error())
		errHTML := fmt.Sprintf(
			`<div id="seer_assistant_errors" hx-swap-oob="true"><div class="alert alert-error text-sm">%s</div></div>%s`,
			escaped,
			assistantChatFormReadyHTML(),
		)
		return []byte(errHTML), websocket.TextFrame, nil
	}
	msg, isMsg := v.(UserMessage)
	if isMsg {
		messageString := msg.ToHTML()
		return []byte(messageString), websocket.TextFrame, nil
	}
	messageString, err := json.Marshal(v)
	return messageString, websocket.TextFrame, err
}

var codec = websocket.Codec{
	Unmarshal: unmarshal,
	Marshal:   marshal,
}

// assistantWebSocketSyncFromDB validates optional ?session_id= and dry-runs [genai.Chats.Create]
// so transcript from the DB is compatible with the chat session API before any message is sent.
func assistantWebSocketSyncFromDB(req *http.Request) {
	ctx := req.Context()
	sidStr := req.URL.Query().Get("session_id")
	if sidStr == "" {
		return
	}
	sid64, err := strconv.ParseUint(sidStr, 10, 64)
	if err != nil || sid64 == 0 {
		slog.Warn("seer_assistant: ws open invalid session_id", "value", sidStr, "error", err)
		return
	}
	sid := uint(sid64)
	user, ok := ctx.Value("$user").(p_users.User)
	if !ok {
		return
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		slog.Warn("seer_assistant: ws open no db", "error", err)
		return
	}
	var n int64
	if err := db.WithContext(ctx).Model(&SeerAssistantSession{}).Where("id = ? AND user_id = ?", sid, user.ID).Count(&n).Error; err != nil || n == 0 {
		slog.Warn("seer_assistant: ws open session not found or denied", "session_id", sid, "error", err)
		return
	}
	contents, err := LoadSessionContents(ctx, db, sid)
	if err != nil {
		slog.Warn("seer_assistant: ws open load session failed", "session_id", sid, "error", err)
		return
	}
	client, err := p_google_genai.NewClient(ctx)
	if err != nil {
		slog.Warn("seer_assistant: ws open genai client", "error", err)
		return
	}
	maxOut := AssistantAppConfig.ChatMaxOutputTokens
	if maxOut <= 0 {
		maxOut = 1024
	}
	model := SeerAssistantPlugin.ChatModel
	if _, err := client.Chats.Create(ctx, model, assistantChatGenConfig(maxOut), contents); err != nil {
		slog.Warn("seer_assistant: ws open chats create dry-run failed", "session_id", sid, "error", err)
		return
	}
	slog.Info("seer_assistant: ws chat synced with db", "session_id", sid, "content_messages", len(contents))
}

func assistantWebSocketConn(ws *websocket.Conn) {
	req := ws.Request()
	if req == nil {
		return
	}
	assistantWebSocketSyncFromDB(req)
	for {
		var userMessage UserMessage
		err := codec.Receive(ws, &userMessage)
		if err != nil {
			if err != io.EOF {
				slog.Warn("seer assistant websocket receive failed", "error", err)
				if sendErr := codec.Send(ws, err); sendErr != nil {
					slog.Error("Error while sending websocket error", "error", sendErr)
				}
			}
			return
		}
		savedMessage, err := userMessage.Save(req)
		if err != nil {
			slog.Warn("seer assistant websocket save failed", "error", err)
			if sendErr := codec.Send(ws, err); sendErr != nil {
				slog.Error("Error while sending websocket error", "error", sendErr)
			}
			return
		}
		if err = codec.Send(ws, savedMessage); err != nil {
			slog.Warn("seer assistant websocket send user message failed", "error", err)
			if sendErr := codec.Send(ws, err); sendErr != nil {
				slog.Error("Error while sending websocket error", "error", sendErr)
			}
			return
		}
		contentChan, errChan := RunAssistant(req, savedMessage.SessionID)
		var streamedAssistant *genai.Content
		for contentChan != nil || errChan != nil {
			select {
			case content, ok := <-contentChan:
				if !ok {
					contentChan = nil
					continue
				}
				var sendErr error
				role := strings.TrimSpace(content.Role)
				switch {
				case role == "tool" || assistantContentHasToolResponseParts(content):
					// Synthetic role "tool" (legacy) or persisted shape: user + function/tool response parts.
					streamedAssistant = nil
					sendErr = writeWSHTML(ws, assistantClearStreamHTML())
					if sendErr == nil {
						sendErr = writeWSHTML(ws, assistantToolHTML(content))
					}
				default:
					streamedAssistant = mergeAssistantContent(streamedAssistant, content)
					sendErr = writeWSHTML(ws, assistantStreamHTML(streamedAssistant))
				}
				if sendErr != nil {
					slog.Warn("seer assistant websocket send stream failed", "error", sendErr)
					if err = codec.Send(ws, sendErr); err != nil {
						slog.Error("Error while sending websocket error", "error", err)
					}
					return
				}
			case err, ok := <-errChan:
				if !ok {
					errChan = nil
					continue
				}
				if err != nil {
					slog.Warn("seer assistant run failed", "error", err)
					if sendErr := codec.Send(ws, err); sendErr != nil {
						slog.Error("Error while sending websocket error", "error", sendErr)
					}
					return
				}
			}
		}
		if streamedAssistant != nil {
			if err = writeWSHTML(ws, assistantClearStreamHTML()); err != nil {
				slog.Warn("seer assistant websocket clear stream failed", "error", err)
				if sendErr := codec.Send(ws, err); sendErr != nil {
					slog.Error("Error while sending websocket error", "error", sendErr)
				}
				return
			}
			if err = writeWSHTML(ws, assistantFinalHTML(streamedAssistant)); err != nil {
				slog.Warn("seer assistant websocket final render failed", "error", err)
				if sendErr := codec.Send(ws, err); sendErr != nil {
					slog.Error("Error while sending websocket error", "error", sendErr)
				}
				return
			}
		}
		if err := writeWSHTML(ws, assistantChatFormReadyHTML()); err != nil {
			slog.Warn("seer assistant websocket send chat form ready failed", "error", err)
			return
		}
	}
}

func writeWSHTML(ws *websocket.Conn, s string) error {
	_, err := ws.Write([]byte(s))
	return err
}

func errorOOB(err error) string {
	msg := html.EscapeString(err.Error())
	return fmt.Sprintf(
		`<div id="seer_assistant_errors" hx-swap-oob="true"><div class="alert alert-error text-sm">%s</div></div>`,
		msg,
	)
}

func assistantChatFormReadyHTML() string {
	return `<button id="seer_assistant_chat_send" hx-swap-oob="true" type="submit" class="btn btn-primary self-end">Send</button>`
}

func assistantClearStreamHTML() string {
	return `<div id="seer_assistant_stream" hx-swap-oob="true" class="min-h-[1.5rem] border border-dashed border-base-300 rounded p-2 text-sm"></div>`
}

// assistantStreamHTML replaces #seer_assistant_stream with the same HTML pipeline as
// transcript/final ([assistantGenaiContentHTML]) applied to the merged assistant turn so far.
func assistantStreamHTML(merged *genai.Content) string {
	if merged == nil {
		return ""
	}
	inner := strings.TrimSpace(assistantGenaiContentHTML(merged))
	if inner == "" {
		return ""
	}
	return fmt.Sprintf(
		`<div id="seer_assistant_stream" hx-swap-oob="true" class="min-h-[1.5rem] border border-dashed border-base-300 rounded p-2 text-sm">%s</div>`,
		inner,
	)
}

func assistantToolHTML(content *genai.Content) string {
	inner := assistantGenaiContentHTML(content)
	if inner == "" {
		inner = `<span class="opacity-50 text-sm">(empty)</span>`
	}
	return fmt.Sprintf(
		`<div id="seer_assistant_transcript" hx-swap-oob="beforeend"><div class="chat chat-start mb-2"><div class="chat-header text-xs opacity-70">Tool</div><div class="chat-bubble chat-bubble-neutral text-sm text-base-content">%s</div></div></div>`,
		inner,
	)
}

func assistantFinalHTML(content *genai.Content) string {
	inner := assistantGenaiContentHTML(content)
	if inner == "" {
		inner = `<span class="opacity-50">(empty)</span>`
	}
	return fmt.Sprintf(
		`<div id="seer_assistant_transcript" hx-swap-oob="beforeend"><div class="chat chat-start mb-2"><div class="chat-header text-xs opacity-70">Assistant</div><div class="chat-bubble chat-bubble-secondary">%s</div></div></div>`,
		inner,
	)
}

// assistantGenaiContentHTML renders session/stream *genai.Content for WS OOB HTML.
// Parts are streamed / token-chunk slices: consecutive text parts are joined, then one
// markdown pass ([components.RenderMarkdown]); non-text parts render as structured HTML.
func assistantGenaiContentHTML(c *genai.Content) string {
	if c == nil {
		return ""
	}
	var b strings.Builder
	var textRun struct {
		sb      strings.Builder
		thought bool
		active  bool
	}
	flushTextRun := func() {
		if !textRun.active {
			return
		}
		md := strings.TrimSpace(textRun.sb.String())
		th := textRun.thought
		textRun.active = false
		textRun.sb.Reset()
		if md == "" {
			return
		}
		b.WriteString(assistantMarkdownBlockHTML(md, th))
	}
	for _, part := range c.Parts {
		if part == nil {
			continue
		}
		if part.Text != "" {
			if textRun.active && part.Thought != textRun.thought {
				flushTextRun()
			}
			if !textRun.active {
				textRun.active = true
				textRun.thought = part.Thought
			}
			textRun.sb.WriteString(part.Text)
			continue
		}
		flushTextRun()
		if frag := assistantGenaiPartHTMLNonText(part); frag != "" {
			b.WriteString(frag)
		}
	}
	flushTextRun()
	return strings.TrimSpace(b.String())
}

func assistantMarkdownBlockHTML(md string, thought bool) string {
	if md == "" {
		return ""
	}
	inner := components.RenderMarkdown(md)
	if thought {
		return `<div class="assistant-part assistant-part-thought rounded-md border border-warning/30 bg-warning/10 p-2">` +
			`<div class="mb-1 text-xs font-medium text-warning">Thought</div>` +
			`<div class="prose prose-sm max-w-none">` + inner + `</div></div>`
	}
	return `<div class="assistant-part assistant-part-text prose prose-sm max-w-none">` + inner + `</div>`
}

func assistantGenaiPartHTMLNonText(part *genai.Part) string {
	if part == nil {
		return ""
	}
	var out string
	switch {
	case part.FunctionCall != nil:
		out = assistantFunctionCallHTML(part.FunctionCall)
	case part.FunctionResponse != nil:
		out = assistantFunctionResponseHTML(part.FunctionResponse)
	case part.ToolCall != nil:
		out = assistantToolCallHTML(part.ToolCall)
	case part.ToolResponse != nil:
		out = assistantToolResponseHTML(part.ToolResponse)
	case part.ExecutableCode != nil:
		out = assistantExecutableCodeHTML(part.ExecutableCode)
	case part.CodeExecutionResult != nil:
		out = assistantCodeExecutionResultHTML(part.CodeExecutionResult)
	case part.FileData != nil:
		out = assistantFileDataHTML(part.FileData)
	case part.InlineData != nil:
		out = assistantInlineBlobHTML(part.InlineData)
	default:
		return ""
	}
	if len(part.PartMetadata) > 0 {
		out += `<details class="mt-2 text-xs opacity-80"><summary>Part metadata</summary>` +
			assistantMapHTML(part.PartMetadata) + `</details>`
	}
	if part.VideoMetadata != nil {
		out += assistantVideoMetadataHTML(part.VideoMetadata)
	}
	return out
}

func assistantFunctionCallHTML(fc *genai.FunctionCall) string {
	if fc == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="assistant-part assistant-part-fn-call text-sm">`)
	b.WriteString(`<div class="mb-2 font-semibold text-base-content">Function call`)
	if fc.Name != "" {
		b.WriteString(`: <span class="font-mono">`)
		b.WriteString(html.EscapeString(fc.Name))
		b.WriteString(`</span>`)
	}
	b.WriteString(`</div>`)
	if fc.ID != "" {
		b.WriteString(`<div class="mb-1 text-xs opacity-70">ID <code>`)
		b.WriteString(html.EscapeString(fc.ID))
		b.WriteString(`</code></div>`)
	}
	if fc.WillContinue != nil {
		b.WriteString(fmt.Sprintf(`<div class="mb-1 text-xs">willContinue: <span class="font-mono">%t</span></div>`, *fc.WillContinue))
	}
	if len(fc.Args) > 0 {
		b.WriteString(`<div class="text-xs font-medium opacity-70 mb-1">Arguments</div>`)
		b.WriteString(assistantMapHTML(fc.Args))
	} else {
		b.WriteString(`<div class="text-xs opacity-50">No arguments</div>`)
	}
	if len(fc.PartialArgs) > 0 {
		b.WriteString(`<div class="mt-2 text-xs font-medium opacity-70">Streaming arguments</div>`)
		b.WriteString(assistantPartialArgsHTML(fc.PartialArgs))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func assistantPartialArgsHTML(pas []*genai.PartialArg) string {
	var b strings.Builder
	b.WriteString(`<table class="mt-1 w-full text-xs border-collapse"><thead><tr class="border-b border-base-300"><th class="py-1 text-left font-medium">JSON path</th><th class="py-1 text-left font-medium">Value</th></tr></thead><tbody>`)
	for _, pa := range pas {
		if pa == nil {
			continue
		}
		b.WriteString(`<tr class="align-top border-b border-base-300/50"><td class="py-1 pr-2 font-mono">`)
		b.WriteString(html.EscapeString(pa.JsonPath))
		b.WriteString(`</td><td class="py-1">`)
		b.WriteString(assistantPartialArgValueHTML(pa))
		b.WriteString(`</td></tr>`)
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}

func assistantPartialArgValueHTML(pa *genai.PartialArg) string {
	switch {
	case pa.StringValue != "":
		return `<span class="whitespace-pre-wrap">` + html.EscapeString(pa.StringValue) + `</span>`
	case pa.NumberValue != nil:
		return fmt.Sprintf(`<span class="font-mono">%g</span>`, *pa.NumberValue)
	case pa.BoolValue != nil:
		return fmt.Sprintf(`<span class="font-mono">%t</span>`, *pa.BoolValue)
	case pa.NULLValue != "":
		return `<span class="opacity-50">null</span>`
	default:
		return `<span class="opacity-50">—</span>`
	}
}

func assistantFunctionResponseHTML(fr *genai.FunctionResponse) string {
	if fr == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="assistant-part assistant-part-fn-resp text-sm">`)
	b.WriteString(`<div class="mb-2 font-semibold text-base-content">Function response`)
	if fr.Name != "" {
		b.WriteString(`: <span class="font-mono">`)
		b.WriteString(html.EscapeString(fr.Name))
		b.WriteString(`</span>`)
	}
	b.WriteString(`</div>`)
	if fr.ID != "" {
		b.WriteString(`<div class="mb-1 text-xs opacity-70">Call ID <code>`)
		b.WriteString(html.EscapeString(fr.ID))
		b.WriteString(`</code></div>`)
	}
	if fr.WillContinue != nil {
		b.WriteString(fmt.Sprintf(`<div class="mb-1 text-xs">willContinue: <span class="font-mono">%t</span></div>`, *fr.WillContinue))
	}
	if fr.Scheduling != "" {
		b.WriteString(`<div class="mb-1 text-xs opacity-70">Scheduling <code>`)
		b.WriteString(html.EscapeString(string(fr.Scheduling)))
		b.WriteString(`</code></div>`)
	}
	if len(fr.Response) > 0 {
		b.WriteString(`<div class="text-xs font-medium opacity-70 mb-1">Response</div>`)
		b.WriteString(assistantMapHTML(fr.Response))
	}
	if len(fr.Parts) > 0 {
		b.WriteString(`<div class="mt-2 text-xs font-medium opacity-70">Media parts</div>`)
		for _, p := range fr.Parts {
			if p == nil {
				continue
			}
			if p.InlineData != nil {
				b.WriteString(assistantFunctionResponseBlobHTML(p.InlineData))
			}
			if p.FileData != nil {
				b.WriteString(assistantFunctionResponseFileDataHTML(p.FileData))
			}
		}
	}
	b.WriteString(`</div>`)
	return b.String()
}

func assistantFunctionResponseBlobHTML(blob *genai.FunctionResponseBlob) string {
	if blob == nil {
		return ""
	}
	n := len(blob.Data)
	label := blob.MIMEType
	if blob.DisplayName != "" {
		label = blob.DisplayName + " · " + label
	}
	return fmt.Sprintf(
		`<div class="rounded border border-base-300 p-2 text-xs"><div class="font-medium">Inline media</div><div>%s</div><div class="mt-1 opacity-70">%s</div></div>`,
		html.EscapeString(label),
		html.EscapeString(strconv.Itoa(n)+" bytes"),
	)
}

func assistantFunctionResponseFileDataHTML(fd *genai.FunctionResponseFileData) string {
	if fd == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="rounded border border-base-300 p-2 text-xs"><div class="font-medium">File</div>`)
	if fd.DisplayName != "" {
		b.WriteString(`<div>` + html.EscapeString(fd.DisplayName) + `</div>`)
	}
	b.WriteString(`<div class="opacity-70">` + html.EscapeString(fd.MIMEType) + `</div>`)
	b.WriteString(`<div class="mt-1 break-all">` + assistantSafeURIHTML(fd.FileURI) + `</div></div>`)
	return b.String()
}

func assistantToolCallHTML(tc *genai.ToolCall) string {
	if tc == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="assistant-part assistant-part-tool-call text-sm">`)
	b.WriteString(`<div class="mb-2 font-semibold text-base-content">Tool call`)
	if tc.ToolType != "" {
		b.WriteString(`: <span class="font-mono">`)
		b.WriteString(html.EscapeString(string(tc.ToolType)))
		b.WriteString(`</span>`)
	}
	b.WriteString(`</div>`)
	if tc.ID != "" {
		b.WriteString(`<div class="mb-1 text-xs opacity-70">ID <code>`)
		b.WriteString(html.EscapeString(tc.ID))
		b.WriteString(`</code></div>`)
	}
	if len(tc.Args) > 0 {
		b.WriteString(`<div class="text-xs font-medium opacity-70 mb-1">Arguments</div>`)
		b.WriteString(assistantMapHTML(tc.Args))
	} else {
		b.WriteString(`<div class="text-xs opacity-50">No arguments</div>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

func assistantToolResponseHTML(tr *genai.ToolResponse) string {
	if tr == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="assistant-part assistant-part-tool-resp text-sm">`)
	b.WriteString(`<div class="mb-2 font-semibold text-base-content">Tool response`)
	if tr.ToolType != "" {
		b.WriteString(`: <span class="font-mono">`)
		b.WriteString(html.EscapeString(string(tr.ToolType)))
		b.WriteString(`</span>`)
	}
	b.WriteString(`</div>`)
	if tr.ID != "" {
		b.WriteString(`<div class="mb-1 text-xs opacity-70">Call ID <code>`)
		b.WriteString(html.EscapeString(tr.ID))
		b.WriteString(`</code></div>`)
	}
	if len(tr.Response) > 0 {
		b.WriteString(`<div class="text-xs font-medium opacity-70 mb-1">Response</div>`)
		b.WriteString(assistantMapHTML(tr.Response))
	} else {
		b.WriteString(`<div class="text-xs opacity-50">Empty response object</div>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

func assistantExecutableCodeHTML(ec *genai.ExecutableCode) string {
	if ec == nil {
		return ""
	}
	lang := string(ec.Language)
	var meta strings.Builder
	if ec.ID != "" {
		meta.WriteString(`<div class="mb-1 text-xs opacity-70">ID <code>`)
		meta.WriteString(html.EscapeString(ec.ID))
		meta.WriteString(`</code></div>`)
	}
	return fmt.Sprintf(
		`<div class="assistant-part assistant-part-code rounded-box border border-base-300 p-2 text-sm">%s<div class="mb-1 text-xs font-medium opacity-70">Language: <code>%s</code></div><pre class="mt-1 max-h-64 overflow-auto rounded bg-base-300/50 p-2 text-xs"><code class="language-%s">%s</code></pre></div>`,
		meta.String(),
		html.EscapeString(lang),
		html.EscapeString(lang),
		html.EscapeString(ec.Code),
	)
}

func assistantCodeExecutionResultHTML(cer *genai.CodeExecutionResult) string {
	if cer == nil {
		return ""
	}
	outcome := string(cer.Outcome)
	var idLine string
	if cer.ID != "" {
		idLine = `<div class="mb-1 text-xs opacity-70">Executable ID <code>` + html.EscapeString(cer.ID) + `</code></div>`
	}
	return fmt.Sprintf(
		`<div class="assistant-part assistant-part-code-result rounded-box border border-base-300 p-2 text-sm">%s<div class="mb-2"><span class="rounded bg-neutral px-2 py-0.5 text-xs font-mono">%s</span></div><pre class="max-h-64 overflow-auto whitespace-pre-wrap rounded bg-base-300/50 p-2 text-xs">%s</pre></div>`,
		idLine,
		html.EscapeString(outcome),
		html.EscapeString(cer.Output),
	)
}

func assistantFileDataHTML(fd *genai.FileData) string {
	if fd == nil {
		return ""
	}
	var head strings.Builder
	head.WriteString(`<div class="assistant-part assistant-part-file rounded-box border border-base-300 p-2 text-xs">`)
	if fd.DisplayName != "" {
		head.WriteString(`<div class="font-medium">` + html.EscapeString(fd.DisplayName) + `</div>`)
	}
	head.WriteString(`<div class="opacity-70">` + html.EscapeString(fd.MIMEType) + `</div>`)
	head.WriteString(`<div class="mt-1 break-all">` + assistantSafeURIHTML(fd.FileURI) + `</div></div>`)
	return head.String()
}

func assistantInlineBlobHTML(blob *genai.Blob) string {
	if blob == nil {
		return ""
	}
	n := len(blob.Data)
	label := blob.MIMEType
	if blob.DisplayName != "" {
		label = blob.DisplayName + " · " + label
	}
	return fmt.Sprintf(
		`<div class="assistant-part assistant-part-inline rounded-box border border-base-300 p-2 text-xs"><div class="font-medium">Inline data</div><div>%s</div><div class="mt-1 opacity-70">%s</div></div>`,
		html.EscapeString(label),
		html.EscapeString(strconv.Itoa(n)+" bytes"),
	)
}

func assistantVideoMetadataHTML(vm *genai.VideoMetadata) string {
	if vm == nil {
		return ""
	}
	var bits []string
	if vm.StartOffset != 0 {
		bits = append(bits, "start "+vm.StartOffset.String())
	}
	if vm.EndOffset != 0 {
		bits = append(bits, "end "+vm.EndOffset.String())
	}
	if vm.FPS != nil {
		bits = append(bits, fmt.Sprintf("fps %g", *vm.FPS))
	}
	if len(bits) == 0 {
		return ""
	}
	return `<div class="mt-2 text-xs opacity-70">Video: ` + html.EscapeString(strings.Join(bits, ", ")) + `</div>`
}

func assistantSafeURIHTML(uri string) string {
	if uri == "" {
		return `<span class="opacity-50">(no URI)</span>`
	}
	esc := html.EscapeString(uri)
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		return fmt.Sprintf(`<a href="%s" class="link" target="_blank" rel="noopener noreferrer">%s</a>`, esc, esc)
	}
	return `<code>` + esc + `</code>`
}

func assistantMapHTML(m map[string]any) string {
	if len(m) == 0 {
		return `<span class="opacity-50">{}</span>`
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(`<dl class="assistant-kv grid gap-x-3 gap-y-1 text-sm" style="grid-template-columns:auto 1fr">`)
	for _, k := range keys {
		b.WriteString(`<dt class="text-xs font-medium opacity-70">`)
		b.WriteString(html.EscapeString(k))
		b.WriteString(`</dt><dd class="min-w-0">`)
		b.WriteString(assistantAnyHTML(m[k]))
		b.WriteString(`</dd>`)
	}
	b.WriteString(`</dl>`)
	return b.String()
}

func assistantAnyHTML(v any) string {
	if v == nil {
		return `<span class="opacity-50 italic">null</span>`
	}
	switch t := v.(type) {
	case map[string]any:
		return assistantMapHTML(t)
	case []any:
		return assistantSliceHTML(t)
	case string:
		return `<span class="whitespace-pre-wrap">` + html.EscapeString(t) + `</span>`
	case json.Number:
		return `<span class="font-mono">` + html.EscapeString(t.String()) + `</span>`
	case float64:
		return fmt.Sprintf(`<span class="font-mono">%g</span>`, t)
	case float32:
		return fmt.Sprintf(`<span class="font-mono">%g</span>`, t)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf(`<span class="font-mono">%d</span>`, t)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf(`<span class="font-mono">%d</span>`, t)
	case bool:
		if t {
			return `<span class="font-mono">true</span>`
		}
		return `<span class="font-mono">false</span>`
	default:
		return `<span class="font-mono text-xs">` + html.EscapeString(fmt.Sprint(v)) + `</span>`
	}
}

func assistantSliceHTML(a []any) string {
	if len(a) == 0 {
		return `<span class="opacity-50">[]</span>`
	}
	var b strings.Builder
	b.WriteString(`<ul class="assistant-list list-disc space-y-1 pl-4 text-sm">`)
	for _, item := range a {
		b.WriteString(`<li class="min-w-0">`)
		b.WriteString(assistantAnyHTML(item))
		b.WriteString(`</li>`)
	}
	b.WriteString(`</ul>`)
	return b.String()
}
