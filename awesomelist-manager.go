package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"sync"
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

func (alm *AwesomeListManager) EnrichListDry() error {
	alm.EnrichedList = make(enrichedAwesomelist, len(alm.RawList))
	for i, baseCat := range alm.RawList {
		alm.EnrichedList[i] = *baseEnrichCategory(baseCat)
	}
	return nil
}

func (alm *AwesomeListManager) EnrichList() error {
	alm.EnrichedList = make(enrichedAwesomelist, len(alm.RawList))
	for i, baseCat := range alm.RawList {
		alm.EnrichedList[i] = *baseEnrichCategory(baseCat)
	}

	var wg sync.WaitGroup
	const maxConcurrency = 100
	sem := make(chan struct{}, maxConcurrency)
	errCh := make(chan error, len(alm.EnrichedList))

	for i := range alm.EnrichedList {
		wg.Add(1)
		go func(index int) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()
			if err := remoteEnrichCategory(&alm.EnrichedList[index], sem); err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			log.Printf("Error during remote enrichment: %v", err)
		}
	}
	return nil
}

func baseEnrichCategory(baseCategory BaseCategory) *EnrichedCategory {
	enrichedCat := &EnrichedCategory{
		Title:       baseCategory.Title,
		Description: baseCategory.Description,
	}
	enrichedCat.Slug = slugifiy(enrichedCat.Title)

	enrichedCat.Links = make([]EnrichedLink, len(baseCategory.Links))
	for i, link := range baseCategory.Links {
		enrichedCat.Links[i] = *baseEnrichLink(link)
	}

	enrichedCat.Subcategories = make([]EnrichedCategory, len(baseCategory.Subcategories))
	for i, subCat := range baseCategory.Subcategories {
		enrichedCat.Subcategories[i] = *baseEnrichCategory(subCat)
	}

	return enrichedCat
}

func baseEnrichLink(baseLink BaseLink) *EnrichedLink {
	return &EnrichedLink{
		Title:       baseLink.Title,
		Description: baseLink.Description,
		Url:         baseLink.Url,
	}
}

func remoteEnrichCategory(enrichedCat *EnrichedCategory, sem chan struct{}) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(enrichedCat.Links)+len(enrichedCat.Subcategories))

	for i := range enrichedCat.Links {
		wg.Add(1)
		go func(index int) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()
			if err := remoteEnrichLink(&enrichedCat.Links[index]); err != nil {
				errCh <- err
			}
		}(i)
	}

	for i := range enrichedCat.Subcategories {
		wg.Add(1)
		go func(index int) {
			sem <- struct{}{}
			defer func() {
				<-sem
				wg.Done()
			}()
			if err := remoteEnrichCategory(&enrichedCat.Subcategories[index], sem); err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func remoteEnrichLink(enrichedLink *EnrichedLink) error {
	repo := NewRemoteRepo(enrichedLink.Url)
	if repo == nil {
		return nil
	}
	enrichedLink.IsRepo = true
	err := repo.Enrich()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch repo data for %s: %v\n", enrichedLink.Url, err)
		return nil
	}
	enrichedLink.Stars = repo.Stars()
	enrichedLink.LastUpdate = repo.LastUpdate()
	// enrichedLink.IsArchived = repo.IsArchived()
	return nil
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
