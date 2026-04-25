package p_seer_assistant

import (
	"github.com/lariv-in/lago/plugins/p_google_genai"
	"google.golang.org/genai"
)

// Gemini function names and JSON parameter shapes must match [assistantEnvFromFunctionCall]
// unmarshaling into [assistantToolEnvelope].

type intelSearchArgs struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type googleSearchArgs struct {
	Query string `json:"query"`
}

type redditAddSourceArgs struct {
	RedditRunnerID *uint    `json:"reddit_runner_id,omitempty"`
	Subreddits     []string `json:"subreddits,omitempty"`
	SearchQuery    string   `json:"search_query,omitempty"`
	MaxFreshPosts  uint     `json:"max_fresh_posts,omitempty"`
	LoadWebsites   bool     `json:"load_websites,omitempty"`
}

type redditEditSourceArgs struct {
	RedditSourceID uint     `json:"reddit_source_id"`
	RedditRunnerID *uint    `json:"reddit_runner_id,omitempty"`
	Subreddits     []string `json:"subreddits,omitempty"`
	SearchQuery    string   `json:"search_query,omitempty"`
	MaxFreshPosts  uint     `json:"max_fresh_posts,omitempty"`
	LoadWebsites   bool     `json:"load_websites,omitempty"`
}

type redditEditWorkerArgs struct {
	RedditRunnerID *uint  `json:"reddit_runner_id"`
	WorkerName     string `json:"worker_name"`
	WorkerDuration string `json:"worker_duration"`
}

type redditAddWorkerArgs struct {
	WorkerName     string `json:"worker_name"`
	WorkerDuration string `json:"worker_duration"`
}

type websiteListArgs struct{}

type websiteAddSourceArgs struct {
	SeedURL         string `json:"seed_url"`
	WebsiteDepth    uint   `json:"website_depth,omitempty"`
	WebsiteRunnerFK *uint  `json:"website_runner_id,omitempty"`
}

type websiteEditSourceArgs struct {
	WebsiteSourceID uint   `json:"website_source_id"`
	SeedURL         string `json:"seed_url"`
	WebsiteDepth    uint   `json:"website_depth,omitempty"`
	WebsiteRunnerFK *uint  `json:"website_runner_id,omitempty"`
}

type websiteAddWorkerArgs struct {
	WorkerName     string `json:"worker_name"`
	WorkerDuration string `json:"worker_duration"`
}

type websiteEditWorkerArgs struct {
	WebsiteRunnerFK uint   `json:"website_runner_id"`
	WorkerName      string `json:"worker_name"`
	WorkerDuration  string `json:"worker_duration"`
}

// assistantGeminiTools returns function declarations for Seer Assistant (native tool calling).
func assistantGeminiTools() []*genai.Tool {
	decls := []*genai.FunctionDeclaration{
		{
			Name:        "intel_search",
			Description: "Vector search over the Intel database (summaries of scraped pages and ingested content). Returns id, title, summary, kind.",
			Parameters:  p_google_genai.NewSchema[intelSearchArgs](),
		},
		{
			Name:        "google_search",
			Description: "Search the public web via Google Custom Search (configured in Lago). Use before adding Reddit sources when you need to discover or verify names on the web.",
			Parameters:  p_google_genai.NewSchema[googleSearchArgs](),
		},
		{
			Name:        "reddit_add_source",
			Description: "Create a Reddit ingestion source (subreddits and/or search query, optional runner).",
			Parameters:  p_google_genai.NewSchema[redditAddSourceArgs](),
		},
		{
			Name:        "reddit_edit_source",
			Description: "Update an existing Reddit source by reddit_source_id.",
			Parameters:  p_google_genai.NewSchema[redditEditSourceArgs](),
		},
		{
			Name:        "reddit_add_worker",
			Description: "Create a Reddit runner worker schedule (worker_name, worker_duration as Go duration string e.g. 1h, 30m).",
			Parameters:  p_google_genai.NewSchema[redditAddWorkerArgs](),
		},
		{
			Name:        "reddit_edit_worker",
			Description: "Update a Reddit runner worker schedule (reddit_runner_id, worker name, Go duration string e.g. 1h, 30m).",
			Parameters:  p_google_genai.NewSchema[redditEditWorkerArgs](),
		},
		{
			Name:        "website_list_sources",
			Description: "List Seer Websites crawl sources (JSON).",
			Parameters:  p_google_genai.NewSchema[websiteListArgs](),
		},
		{
			Name:        "website_list_workers",
			Description: "List Seer Websites crawl workers/runners (JSON).",
			Parameters:  p_google_genai.NewSchema[websiteListArgs](),
		},
		{
			Name:        "website_add_source",
			Description: "Add a website crawl source (seed_url, optional depth and website_runner_id).",
			Parameters:  p_google_genai.NewSchema[websiteAddSourceArgs](),
		},
		{
			Name:        "website_edit_source",
			Description: "Edit a website source by website_source_id (seed_url, optional depth and runner).",
			Parameters:  p_google_genai.NewSchema[websiteEditSourceArgs](),
		},
		{
			Name:        "website_add_worker",
			Description: "Add a website crawl worker (worker_name, worker_duration as Go duration string).",
			Parameters:  p_google_genai.NewSchema[websiteAddWorkerArgs](),
		},
		{
			Name:        "website_edit_worker",
			Description: "Edit a website worker by website_runner_id (name and duration).",
			Parameters:  p_google_genai.NewSchema[websiteEditWorkerArgs](),
		},
	}
	return []*genai.Tool{{FunctionDeclarations: decls}}
}
