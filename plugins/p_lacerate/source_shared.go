package p_lacerate

import (
	"context"
	"fmt"
	"gorm.io/gorm"
)

type SourcePageData struct {
	Source       Source
	Reddit       *RedditSource
	Twitter      *TwitterSource
	Website      *WebsiteSource
	GoogleSearch *GoogleSearchSource
	Websearch    *WebsearchSource
	DirectMedia  *DirectMediaSource
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
	case sourceKindGoogleSearch:
		var row GoogleSearchSource
		if err := db.WithContext(ctx).Where("source_id = ?", sourceID).First(&row).Error; err != nil {
			return data, err
		}
		row.Source = data.Source
		data.GoogleSearch = &row
	case sourceKindWebsearch:
		var row WebsearchSource
		if err := db.WithContext(ctx).Where("source_id = ?", sourceID).First(&row).Error; err != nil {
			return data, err
		}
		row.Source = data.Source
		data.Websearch = &row
	case sourceKindDirectMedia:
		var row DirectMediaSource
		if err := db.WithContext(ctx).Where("source_id = ?", sourceID).First(&row).Error; err != nil {
			return data, err
		}
		row.Source = data.Source
		data.DirectMedia = &row
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

	var googleSearchRows []GoogleSearchSource
	if err := db.WithContext(ctx).Where("source_id IN ?", sourceIDs).Find(&googleSearchRows).Error; err != nil {
		return nil, err
	}
	googleSearchBySourceID := make(map[uint]GoogleSearchSource, len(googleSearchRows))
	for _, row := range googleSearchRows {
		googleSearchBySourceID[row.SourceID] = row
	}
	var websearchRows []WebsearchSource
	if err := db.WithContext(ctx).Where("source_id IN ?", sourceIDs).Find(&websearchRows).Error; err != nil {
		return nil, err
	}
	websearchBySourceID := make(map[uint]WebsearchSource, len(websearchRows))
	for _, row := range websearchRows {
		websearchBySourceID[row.SourceID] = row
	}

	var directMediaRows []DirectMediaSource
	if err := db.WithContext(ctx).Where("source_id IN ?", sourceIDs).Find(&directMediaRows).Error; err != nil {
		return nil, err
	}
	directMediaBySourceID := make(map[uint]DirectMediaSource, len(directMediaRows))
	for _, row := range directMediaRows {
		directMediaBySourceID[row.SourceID] = row
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
		if row, ok := googleSearchBySourceID[source.ID]; ok {
			row.Source = source
			item.GoogleSearch = &row
		}
		if row, ok := websearchBySourceID[source.ID]; ok {
			row.Source = source
			item.Websearch = &row
		}
		if row, ok := directMediaBySourceID[source.ID]; ok {
			row.Source = source
			item.DirectMedia = &row
		}
		items = append(items, item)
	}
	return items, nil
}
