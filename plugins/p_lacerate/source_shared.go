package p_lacerate

import (
	"context"
	"fmt"
	"gorm.io/gorm"
)

type SourcePageData struct {
	Source  Source
	Reddit  *RedditSource
	Twitter *TwitterSource
	Website *WebsiteSource
}

func loadSourcePageData(ctx context.Context, db *gorm.DB, sourceID uint) (SourcePageData, error) {
	var data SourcePageData
	if err := db.WithContext(ctx).First(&data.Source, sourceID).Error; err != nil {
		return data, err
	}
	switch data.Source.Kind {
	case "reddit":
		var row RedditSource
		if err := db.WithContext(ctx).Where("source_id = ?", sourceID).First(&row).Error; err != nil {
			return data, err
		}
		row.Source = data.Source
		data.Reddit = &row
	case "twitter":
		var row TwitterSource
		if err := db.WithContext(ctx).Where("source_id = ?", sourceID).First(&row).Error; err != nil {
			return data, err
		}
		row.Source = data.Source
		data.Twitter = &row
	case "website":
		var row WebsiteSource
		if err := db.WithContext(ctx).Where("source_id = ?", sourceID).First(&row).Error; err != nil {
			return data, err
		}
		row.Source = data.Source
		data.Website = &row
	case "":
		return data, fmt.Errorf("source %d has empty kind", sourceID)
	default:
		return data, fmt.Errorf("unsupported source kind %q", data.Source.Kind)
	}
	return data, nil
}

func loadSourcePageDataList(ctx context.Context, db *gorm.DB, sources []Source) ([]SourcePageData, error) {
	if len(sources) == 0 {
		return nil, nil
	}
	sourceIDs := make([]uint, 0, len(sources))
	for _, source := range sources {
		sourceIDs = append(sourceIDs, source.ID)
	}

	var redditRows []RedditSource
	if err := db.WithContext(ctx).Where("source_id IN ?", sourceIDs).Find(&redditRows).Error; err != nil {
		return nil, err
	}
	redditBySourceID := make(map[uint]RedditSource, len(redditRows))
	for _, row := range redditRows {
		redditBySourceID[row.SourceID] = row
	}

	var twitterRows []TwitterSource
	if err := db.WithContext(ctx).Where("source_id IN ?", sourceIDs).Find(&twitterRows).Error; err != nil {
		return nil, err
	}
	twitterBySourceID := make(map[uint]TwitterSource, len(twitterRows))
	for _, row := range twitterRows {
		twitterBySourceID[row.SourceID] = row
	}

	var websiteRows []WebsiteSource
	if err := db.WithContext(ctx).Where("source_id IN ?", sourceIDs).Find(&websiteRows).Error; err != nil {
		return nil, err
	}
	websiteBySourceID := make(map[uint]WebsiteSource, len(websiteRows))
	for _, row := range websiteRows {
		websiteBySourceID[row.SourceID] = row
	}

	items := make([]SourcePageData, 0, len(sources))
	for _, source := range sources {
		item := SourcePageData{Source: source}
		if row, ok := redditBySourceID[source.ID]; ok {
			row.Source = source
			item.Reddit = &row
		}
		if row, ok := twitterBySourceID[source.ID]; ok {
			row.Source = source
			item.Twitter = &row
		}
		if row, ok := websiteBySourceID[source.ID]; ok {
			row.Source = source
			item.Website = &row
		}
		items = append(items, item)
	}
	return items, nil
}
