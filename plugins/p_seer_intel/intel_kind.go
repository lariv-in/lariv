package p_seer_intel

// IntelKind is implemented by source-side types that contribute raw text used to build an [Intel]
// row (title, summary, embedding). For example, a future RedditPost model may implement IntelKind
// and hold a 1:1 association to Intel; ingest would call Content(), then generate embeddings and
// summaries from that string. Nothing in this package binds IntelKind to the database yet.
type IntelKind interface {
	Content() string
}
