package p_seer_reddit

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/views"
)

// redditRunnerEnrichSourceIDsLayer sets [RedditRunner.RedditSourceIDs] on GET for the worker
// edit form. [components.InputManyToMany] prefers $in.RedditSourceIDs over the field Getter;
// [getters.MapFromStruct] always includes that field as an empty slice, so without this step
// assigned sources never show as chips.
type redditRunnerEnrichSourceIDsLayer struct{}

func (redditRunnerEnrichSourceIDsLayer) Next(_ views.View, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		redditRunner, ok := ctx.Value("redditRunner").(RedditRunner)
		if !ok || redditRunner.ID == 0 {
			next.ServeHTTP(w, r)
			return
		}
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			next.ServeHTTP(w, r)
			return
		}
		var ids []uint
		if err := db.Model(&RedditSource{}).Where("reddit_runner_id = ?", redditRunner.ID).Order("id DESC").Pluck("id", &ids).Error; err != nil {
			slog.Error("p_seer_reddit: load runner reddit source ids", "error", err, "runner_id", redditRunner.ID)
			next.ServeHTTP(w, r)
			return
		}
		redditRunner.RedditSourceIDs = ids
		ctx = context.WithValue(ctx, "redditRunner", redditRunner)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
