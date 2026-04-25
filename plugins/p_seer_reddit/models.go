package p_seer_reddit

import (
	"time"

	"github.com/lariv-in/lago/lago"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Join table for [RedditSource] ↔ [RedditPost] (many-to-many). GORM migrates this from the association tags.
const (
	RedditSourcePostsJoinTable = "seer_reddit_source_posts"
	RedditPostsTable           = "seer_reddit_posts"
	RedditRunnersTable         = "seer_reddit_runners"
)

// RedditRunner is a cadence or scheduling bucket; [RedditPost.RedditRunnerID] is optional.
type RedditRunner struct {
	gorm.Model

	Name            string        `gorm:"size:64;not null;uniqueIndex"`
	Duration        time.Duration `gorm:"not null"`
	RedditSourceIDs []uint        `gorm:"-"`
}

func (RedditRunner) TableName() string {
	return RedditRunnersTable
}

// redditRunnerFromUpdateDest reads the struct GORM is updating/creating, including
// [RedditRunner.RedditSourceIDs] (gorm:"-"). The hook receiver may not carry ignored fields.
func redditRunnerFromUpdateDest(tx *gorm.DB) *RedditRunner {
	if tx == nil || tx.Statement == nil {
		return nil
	}
	switch d := tx.Statement.Dest.(type) {
	case RedditRunner:
		return &d
	case *RedditRunner:
		if d == nil {
			return nil
		}
		return d
	default:
		return nil
	}
}

func (r *RedditRunner) AfterSave(tx *gorm.DB) error {
	runner := redditRunnerFromUpdateDest(tx)
	if runner == nil {
		runner = r
	}
	if runner == nil || runner.ID == 0 {
		return nil
	}
	// Form gives the full desired set: detach this worker from all sources, then re-attach the submitted ids.
	if err := tx.Model(&RedditSource{}).Where("reddit_runner_id = ?", runner.ID).Update("reddit_runner_id", nil).Error; err != nil {
		return err
	}
	ids := runner.RedditSourceIDs
	if len(ids) == 0 {
		return nil
	}
	return tx.Model(&RedditSource{}).Where("id IN ?", ids).Update("reddit_runner_id", runner.ID).Error
}

// RedditSource configures which subreddits to fetch and how.
type RedditSource struct {
	gorm.Model

	RedditRunnerID *uint         `gorm:"index"`
	RedditRunner   *RedditRunner `gorm:"foreignKey:RedditRunnerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`

	Subreddits    datatypes.JSON `gorm:"type:json"`
	SearchQuery   string
	MaxFreshPosts uint `gorm:"not null;default:25"`
	// LoadWebsites when true: discovered http(s) URLs from fetched posts are sent to [p_seer_websites.WebsiteScrapeURLQueue].
	LoadWebsites bool `gorm:"not null;default:false"`

	RedditPosts []RedditPost `gorm:"many2many:seer_reddit_source_posts;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (RedditSource) TableName() string {
	return "seer_reddit_sources"
}

// RedditPost stores raw Reddit listing data from Reddit JSON.
//
// [PostID] is Reddit’s t3 id and is unique in this table; sources link through join table [RedditSourcePostsJoinTable] (see [RedditSource.RedditPosts]).
type RedditPost struct {
	gorm.Model

	RedditRunnerID *uint         `gorm:"index"`
	RedditRunner   *RedditRunner `gorm:"foreignKey:RedditRunnerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`

	PostID string `gorm:"size:32;not null;uniqueIndex"`
	Title  string `gorm:"not null;default:''"`
	// Selftext is the post body for text posts.
	Selftext string `gorm:"type:text;not null;default:''"`
	Author   string `gorm:"not null;default:''"`
	// Subreddit name without "r/".
	Subreddit string `gorm:"not null;default:'';index"`
	Permalink string `gorm:"not null;default:''"`
	URL       string `gorm:"not null;default:''"`
	// CreatedUTCUnix is the post's created_utc from Reddit JSON.
	CreatedUTCUnix float64 `gorm:"not null"`
	Score          int     `gorm:"not null;default:0"`
	NumComments    int     `gorm:"not null;default:0"`
	IsSelf         bool    `gorm:"not null;default:false"`
}

func (RedditPost) TableName() string {
	return "seer_reddit_posts"
}

func init() {
	lago.OnDBInit("p_seer_reddit.models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[RedditRunner](db)
		lago.RegisterModel[RedditSource](db)
		lago.RegisterModel[RedditPost](db)
		return db
	})
}
