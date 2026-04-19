package p_seer_intel

import "context"

// IntelKind is implemented by source-side types that contribute raw text used to build an [Intel]
// row (title, summary, embedding). For example, RedditPost in package p_seer_reddit implements IntelKind;
// ingest calls [IntelKind.Content] for generation, [IntelKind.Kind] and [IntelKind.IntelID] for
// persisted [Intel.Kind] / [Intel.KindID]. Source plugins register a [IntelKindLoader] on
// [RegistryIntelKind] under the same string [IntelKind.Kind] returns.
type IntelKind interface {
	Content() string
	// Kind returns a stable, short source-family label (e.g. "reddit") stored on [Intel.Kind].
	Kind() string
	// IntelID returns the source model primary key stored on [Intel.KindID].
	IntelID() uint
	// IntelDetail returns the app path to this kind's source detail page (e.g. Reddit post detail for "reddit").
	IntelDetail(ctx context.Context) (string, error)
}
