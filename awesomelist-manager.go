package main

import (
	"fmt"
	"log"
	"slices"
	"sync"
	"time"
)

type AwesomeListManager struct {
	RawList      baseAwesomelist
	EnrichedList enrichedAwesomelist
}

func NewAwesomeListManager(raw baseAwesomelist) *AwesomeListManager {
	return &AwesomeListManager{
		RawList: raw,
	}
}

func (alm *AwesomeListManager) EnrichList() error {
	alm.EnrichedList = make(enrichedAwesomelist, len(alm.RawList))
	var wg sync.WaitGroup
	enrichedCh := make(chan struct {
		enrichedCat *EnrichedCategory
		index       int
	}, len(alm.RawList))

	for i, baseCat := range alm.RawList {
		wg.Add(1)
		go func(index int, category BaseCategory) {
			defer wg.Done()
			enrichedCat, err := enrichCategory(category)
			if err != nil {
				log.Printf("Error enriching category '%s': %v", category.Title, err)
				return
			}
			enrichedCh <- struct {
				enrichedCat *EnrichedCategory
				index       int
			}{enrichedCat: enrichedCat, index: index}
		}(i, baseCat)
	}

	wg.Wait()
	close(enrichedCh)

	for enriched := range enrichedCh {
		alm.EnrichedList[enriched.index] = *enriched.enrichedCat
	}

	return nil
}

func enrichCategory(baseCategory BaseCategory) (*EnrichedCategory, error) {
	enrichedCat := &EnrichedCategory{
		Title:       baseCategory.Title,
		Description: baseCategory.Description,
	}

	enrichedCat.Slug = slugifiy(enrichedCat.Title)

	var wg sync.WaitGroup
	linkCh := make(chan struct {
		link  *EnrichedLink
		index int
	}, len(baseCategory.Links))

	for i, baseLink := range baseCategory.Links {
		wg.Add(1)
		go func(index int, link BaseLink) {
			defer wg.Done()
			enrichedLink, err := enrichLink(link)
			if err != nil {
				log.Printf("error enriching link '%s' in category '%s': %v", link.Title, baseCategory.Title, err)
				return
			}
			linkCh <- struct {
				link  *EnrichedLink
				index int
			}{link: enrichedLink, index: index}
		}(i, baseLink)
	}

	subCatCh := make(chan struct {
		subCat *EnrichedCategory
		index  int
	}, len(baseCategory.Subcategories))

	for i, baseSubCat := range baseCategory.Subcategories {
		wg.Add(1)
		go func(index int, subCat BaseCategory) {
			defer wg.Done()
			enrichedSubCat, err := enrichCategory(subCat)
			if err != nil {
				log.Printf("error enriching subcategory '%s' in category '%s': %v", subCat.Title, baseCategory.Title, err)
				return
			}
			subCatCh <- struct {
				subCat *EnrichedCategory
				index  int
			}{subCat: enrichedSubCat, index: index}
		}(i, baseSubCat)
	}

	wg.Wait()
	close(linkCh)
	close(subCatCh)

	enrichedCat.Links = make([]EnrichedLink, len(baseCategory.Links))
	for result := range linkCh {
		if result.link != nil {
			enrichedCat.Links[result.index] = *result.link
		}
	}

	enrichedCat.Subcategories = make([]EnrichedCategory, len(baseCategory.Subcategories))
	for result := range subCatCh {
		if result.subCat != nil {
			enrichedCat.Subcategories[result.index] = *result.subCat
		}
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

func (alm *AwesomeListManager) AddLink(newLink BaseLink, pathToLink []string) error {
	if len(pathToLink) == 0 {
		return fmt.Errorf("link path cannot be empty")
	}

	rawListPtr := (*[]BaseCategory)(&alm.RawList)

	return addLinkRecursive(rawListPtr, newLink, pathToLink)
}

func (alm *AwesomeListManager) AddCategory(newCategory BaseCategory, pathToLink []string) error {
	rawListPtr := (*[]BaseCategory)(&alm.RawList)

	if len(pathToLink) == 0 {
		var foundCatTitle string
		index := slices.IndexFunc(alm.RawList, func(cat BaseCategory) bool {
			foundCatTitle = cat.Title
			return slugifiy(cat.Title) == slugifiy(newCategory.Title)
		})

		if index != -1 {
			return fmt.Errorf("category with title %q already exist", foundCatTitle)
		}

		*rawListPtr = append(*rawListPtr, newCategory)
		return nil
	}

	return addCategoryRecursive(rawListPtr, newCategory, pathToLink)
}

func addLinkRecursive(categories *[]BaseCategory, newLink BaseLink, pathToLink []string) error {
	titleSlug := slugifiy(pathToLink[0])

	index := slices.IndexFunc(*categories, func(cat BaseCategory) bool {
		return slugifiy(cat.Title) == titleSlug
	})

	if index == -1 {
		return fmt.Errorf("category %q not found", pathToLink[0])
	}

	if len(pathToLink) == 1 {
		var foundCatTitle string
		foundIndex := slices.IndexFunc((*categories)[index].Links, func(link BaseLink) bool {
			foundCatTitle = link.Title
			return slugifiy(link.Title) == slugifiy(newLink.Title)
		})

		if foundIndex != -1 {
			return fmt.Errorf("link with title %q already exist", foundCatTitle)
		}

		(*categories)[index].Links = append((*categories)[index].Links, newLink)
		return nil
	}

	return addLinkRecursive(&(*categories)[index].Subcategories, newLink, pathToLink[1:])
}

func addCategoryRecursive(categories *[]BaseCategory, newCategory BaseCategory, pathToLink []string) error {
	titleSlug := slugifiy(pathToLink[0])

	index := slices.IndexFunc(*categories, func(cat BaseCategory) bool {
		return slugifiy(cat.Title) == titleSlug
	})

	if index == -1 {
		return fmt.Errorf("category %q not found", pathToLink[0])
	}

	if len(pathToLink) == 1 {
		var foundCatTitle string
		foundIndex := slices.IndexFunc((*categories)[index].Subcategories, func(cat BaseCategory) bool {
			foundCatTitle = cat.Title
			return slugifiy(cat.Title) == slugifiy(newCategory.Title)
		})

		if foundIndex != -1 {
			return fmt.Errorf("category with title %q already exist", foundCatTitle)
		}

		(*categories)[index].Subcategories = append((*categories)[index].Subcategories, newCategory)
		return nil
	}

	return addCategoryRecursive(&(*categories)[index].Subcategories, newCategory, pathToLink[1:])
}
