// Package p_llm_assistant implements an interactive LLM chat assistant powered by Gemini.
// It supports chat history storage, custom tool calling (like search engines, database inspections, or code execution),
// user prompt templates (skills), and WebSocket-based streaming chat feeds.
//
// # Registrations and Features Added
//
// # Configurations
//
//   - "p_llm_assistant" -> p_llm_assistant.AssistantPluginConfig
//         Configures the custom Google CSE search credentials (cseApiKey, cseCx) and the active Gemini chat model (chatModel, default gemini-2.5-flash).
//
// # Database Models
//
//   - p_llm_assistant.ChatSession: Represents a unique conversation thread, wrapping user references.
//   - p_llm_assistant.ChatMessage: Stores chat messages, prompt contents, role indicators (user/model/system/tool), and execution duration.
//   - p_llm_assistant.SkillNode: Defines custom prompt templates / system instructions that modify the assistant's behavior.
//
// # Pages
//
//   - "assistant.ChatPage" -> components.PageInterface
//         The main interactive chat UI containing message logs, prompt fields, and active session listings.
//   - "assistant.HistoryPage" -> components.PageInterface
//         Detail panel listing previous chat session histories.
//   - "assistant.SkillsPage" -> components.PageInterface
//         Interface for creating and managing custom prompt skill templates.
//
// # Routes
//
// Registers HTTP ServeMux path mappings:
//
//   - "/llm-assistant/" -> p_llm_assistant.ChatView
//   - "/llm-assistant/history/" -> p_llm_assistant.HistoryListView
//   - "/llm-assistant/skills/" -> p_llm_assistant.SkillsView
//   - "/llm-assistant/ws/" -> p_llm_assistant.WebSocketChatRoute (WebSocket streaming feed endpoint)
//
// # Views
//
//   - "assistant.ChatView": Coordinates active conversation screens.
//   - "assistant.HistoryListView": Lists previous chat sessions.
//   - "assistant.SkillsView": Processes CRUD configurations for custom prompt template nodes.
package p_llm_assistant
