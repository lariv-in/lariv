package p_seer_reddit

import (
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_seer_intel"
	"github.com/lariv-in/lago/plugins/p_seer_runners"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RedditSource configures which subreddits to fetch and how. Optional [p_seer_runners.RunnerID] is for
// scheduling metadata only—this package does not couple ingestion to runners or Intel generation.
type RedditSource struct {
	gorm.Model

	Subreddits    datatypes.JSON `gorm:"type:json"`
	SearchQuery   string
	MaxFreshPosts uint `gorm:"not null;default:25"`

	RunnerID *uint
	Runner   *p_seer_runners.Runner `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

func (RedditSource) TableName() string {
	return "seer_reddit_sources"
}

// RedditPost stores raw Reddit listing data. [IntelID] is set later by the Intel generation pipeline
// (not by Reddit fetch); [Content] supports [p_seer_intel.IntelKind] for that step.
type RedditPost struct {
	gorm.Model

	RedditSourceID uint         `gorm:"not null;uniqueIndex:reddit_post_per_source"`
	RedditSource   RedditSource `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	IntelID *uint                  `gorm:"uniqueIndex"`
	Intel   *p_seer_intel.Intel `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`

	// PostID is Reddit's short id (t3 `id` field). Unique per source.
	PostID string `gorm:"size:32;not null;uniqueIndex:reddit_post_per_source"`
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
	Score          int       `gorm:"not null;default:0"`
	NumComments    int       `gorm:"not null;default:0"`
	IsSelf         bool      `gorm:"not null;default:false"`
}

func (RedditPost) TableName() string {
	return "seer_reddit_posts"
}

func init() {
	lago.OnDBInit("p_seer_reddit.models", func(db *gorm.DB) *gorm.DB {
		lago.RegisterModel[RedditSource](db)
		lago.RegisterModel[RedditPost](db)
		return db
	})
}
