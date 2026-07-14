// Package p_google_genai implements Google Gemini GenAI client wrappers and dynamic schema generators.
// It integrates the official genai SDK client context settings with GORM or custom structures.
//
// # Registrations and Features Added
//
// # Configurations
//
//   - "p_google_genai" -> p_google_genai.Config
//         Configures the client-specific Google Gemini API key (apiKey) mapped from the config.toml file.
//
// # Exported Utility API Handlers
//
// The plugin exposes helper functions to construct AI query clients and JSON schema representations:
//
//   - p_google_genai.NewClient(ctx):
//         Instantiates a new *genai.Client instance referencing the configured apiKey value. Falls back to GOOGLE_API_KEY/GEMINI_API_KEY environment variables if no apiKey was supplied.
//   - p_google_genai.NewSchema[T]():
//         Compiles type properties of any Go data struct T into a *genai.Schema object utilizing reflection. Automatically resolves struct fields, slices, booleans, integer values, and timestamps.
package p_google_genai
