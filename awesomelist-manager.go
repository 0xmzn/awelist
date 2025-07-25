package main

import (
	"fmt"
	"log"
	"time"
)

type AwesomeListManager struct {
	RawList      baseAwesomelist
	EnrichedList enrichedAwesomelist
}

func NewAwesomeDataManager(raw baseAwesomelist, enriched enrichedAwesomelist) *AwesomeListManager {
	return &AwesomeListManager{
		RawList:      raw,
		EnrichedList: enriched,
	}
}

func (alm *AwesomeListManager) EnrichList() error {
	alm.EnrichedList = make(enrichedAwesomelist, len(alm.RawList))
	for i, baseCat := range alm.RawList {
		enrichedCat, err := enrichCategory(baseCat)
		if err != nil {
			log.Printf("Error enriching category '%s': %v", baseCat.Title, err)
			continue
		}
		alm.EnrichedList[i] = *enrichedCat
	}
	return nil
}

func enrichCategory(baseCategory BaseCategory) (*EnrichedCategory, error) {
	enrichedCat := &EnrichedCategory{
		Title:       baseCategory.Title,
		Description: baseCategory.Description,
	}

	sluggifier := NewSlugifier(enrichedCat.Title)
	sluggifier.Enrich()
	enrichedCat.Slug = sluggifier.Slug()

	enrichedCat.Links = make([]EnrichedLink, len(baseCategory.Links))
	for i, baseLink := range baseCategory.Links {
		enrichedLink, err := enrichLink(baseLink)
		if err != nil {
			log.Printf("Error enriching link '%s' in category '%s': %v", baseLink.Title, baseCategory.Title, err)
			continue
		}
		enrichedCat.Links[i] = *enrichedLink
	}

	enrichedCat.Subcategories = make([]EnrichedCategory, len(baseCategory.Subcategories))
	for i, baseSubCat := range baseCategory.Subcategories {
		enrichedSubCat, err := enrichCategory(baseSubCat)
		if err != nil {
			log.Printf("Error enriching subcategory '%s' in category '%s': %v", baseSubCat.Title, baseCategory.Title, err)
			continue
		}
		enrichedCat.Subcategories[i] = *enrichedSubCat
	}

	return enrichedCat, nil
}

func enrichLink(baseLink BaseLink) (*EnrichedLink, error) {
	enrichedLink := &EnrichedLink{
		Title:       baseLink.Title,
		Description: baseLink.Description,
		Url:         baseLink.Url,

		IsRepo:     false,
		Stars:      0,
		LastUpdate: time.Time{},
		IsArchived: false,
	}

	repo := NewRemoteRepo(enrichedLink.Url)
	if repo == nil {
		return enrichedLink, nil
	}

	enrichedLink.IsRepo = true
	err := repo.Enrich()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repo data for %s: %w", baseLink.Url, err)
	}
	enrichedLink.Stars = repo.Stars()
	enrichedLink.LastUpdate = repo.LastUpdate()
	// enrichedLink.IsArchived = repo.IsArchived()

	return enrichedLink, nil
}
