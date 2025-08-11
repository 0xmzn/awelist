package main

import (
	"reflect"
	"slices"
	"testing"
)

func mockAwesomeListManager() *AwesomeListManager {
	initialList := baseAwesomelist{
		{
			Title: "Category A",
			Links: []BaseLink{
				{Title: "Link A1", Url: "http://example.com/a1"},
			},
			Subcategories: []BaseCategory{
				{
					Title: "Subcategory A1",
					Links: []BaseLink{
						{Title: "Link A1.1", Url: "http://example.com/a1-1"},
					},
				},
			},
		},
		{
			Title: "Category B",
			Links: []BaseLink{
				{Title: "Link B1", Url: "http://example.com/b1"},
			},
		},
	}
	return NewAwesomeListManager(initialList)
}

func areBaseLinksEqual(a, b []BaseLink) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func TestAwesomeListManager_AddLink(t *testing.T) {
	tests := []struct {
		name          string
		pathToLink    []string
		newLink       BaseLink
		expectedErr   string
		expectedLinks []BaseLink
	}{
		{
			name:        "Add link to a top-level category",
			pathToLink:  []string{"Category A"},
			newLink:     BaseLink{Title: "New Link A", Url: "http://example.com/newa"},
			expectedErr: "",
			expectedLinks: []BaseLink{
				{Title: "Link A1", Url: "http://example.com/a1"},
				{Title: "New Link A", Url: "http://example.com/newa"},
			},
		},
		{
			name:        "Add link to a subcategory",
			pathToLink:  []string{"Category A", "Subcategory A1"},
			newLink:     BaseLink{Title: "New Sublink A1", Url: "http://example.com/newsuba1"},
			expectedErr: "",
			expectedLinks: []BaseLink{
				{Title: "Link A1.1", Url: "http://example.com/a1-1"},
				{Title: "New Sublink A1", Url: "http://example.com/newsuba1"},
			},
		},
		{
			name:        "Add link to non-existent top-level category",
			pathToLink:  []string{"Nonexistent Category"},
			newLink:     BaseLink{Title: "New Link", Url: "http://example.com/new"},
			expectedErr: "category \"Nonexistent Category\" not found",
		},
		{
			name:        "Add link to non-existent subcategory",
			pathToLink:  []string{"Category B", "Nonexistent Subcategory"},
			newLink:     BaseLink{Title: "New Link", Url: "http://example.com/new"},
			expectedErr: "category \"Nonexistent Subcategory\" not found",
		},
		{
			name:        "Add link with empty path",
			pathToLink:  []string{},
			newLink:     BaseLink{Title: "Empty Path Link", Url: "http://example.com/empty"},
			expectedErr: "link path cannot be empty",
		},
		{
			name:        "Add link with a duplicate title to top-level category",
			pathToLink:  []string{"Category A"},
			newLink:     BaseLink{Title: "Link A1", Url: "http://example.com/a1"},
			expectedErr: "link with title \"Link A1\" already exist",
			expectedLinks: []BaseLink{
				{Title: "Link A1", Url: "http://example.com/a1"},
			},
		},
		{
			name:        "Add link with a duplicate title to subcategory",
			pathToLink:  []string{"Category A", "Subcategory A1"},
			newLink:     BaseLink{Title: "Link A1.1", Url: "http://example.com/a1-1"},
			expectedErr: "link with title \"Link A1.1\" already exist",
			expectedLinks: []BaseLink{
				{Title: "Link A1.1", Url: "http://example.com/a1-1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := mockAwesomeListManager()
			err := manager.AddLink(tt.newLink, tt.pathToLink)

			if tt.expectedErr == "" {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				var targetLinks []BaseLink
				if len(tt.pathToLink) == 1 {
					targetLinks = manager.RawList[0].Links
				} else if len(tt.pathToLink) == 2 {
					targetLinks = manager.RawList[0].Subcategories[0].Links
				}

				if !areBaseLinksEqual(targetLinks, tt.expectedLinks) {
					t.Errorf("Links mismatch. Got: %v, Expected: %v", targetLinks, tt.expectedLinks)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error: %q, but got none", tt.expectedErr)
				}
				if err.Error() != tt.expectedErr {
					t.Errorf("Error mismatch. Got: %q, Expected: %q", err.Error(), tt.expectedErr)
				}
			}
		})
	}
}

func TestAwesomeListManager_AddCategory(t *testing.T) {
	tests := []struct {
		name                string
		pathToCategory      []string
		newCategory         BaseCategory
		expectedErr         string
		expectedCategoryLen int
	}{
		{
			name:                "Add top-level category",
			pathToCategory:      []string{},
			newCategory:         BaseCategory{Title: "New Category"},
			expectedErr:         "",
			expectedCategoryLen: 3,
		},
		{
			name:                "Add subcategory to a top-level category",
			pathToCategory:      []string{"Category A"},
			newCategory:         BaseCategory{Title: "New Subcategory"},
			expectedErr:         "",
			expectedCategoryLen: 2,
		},
		{
			name:           "Add subcategory to non-existent parent",
			pathToCategory: []string{"Nonexistent Parent"},
			newCategory:    BaseCategory{Title: "New Category"},
			expectedErr:    "category \"Nonexistent Parent\" not found",
		},
		{
			name:           "Add top-level category with a duplicate title",
			pathToCategory: []string{},
			newCategory:    BaseCategory{Title: "Category A"},
			expectedErr:    "category with title \"Category A\" already exist",
		},
		{
			name:           "Add subcategory with a duplicate title",
			pathToCategory: []string{"Category A"},
			newCategory:    BaseCategory{Title: "Subcategory A1"},
			expectedErr:    "category with title \"Subcategory A1\" already exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := mockAwesomeListManager()
			err := manager.AddCategory(tt.newCategory, tt.pathToCategory)

			if tt.expectedErr == "" {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if len(tt.pathToCategory) == 0 {
					if len(manager.RawList) != tt.expectedCategoryLen {
						t.Errorf("Expected %d top-level categories, got %d", tt.expectedCategoryLen, len(manager.RawList))
					}
					if !slices.ContainsFunc(manager.RawList, func(c BaseCategory) bool {
						return c.Title == tt.newCategory.Title
					}) {
						t.Errorf("New category %q not found in the list", tt.newCategory.Title)
					}
				} else {
					parentCat := manager.RawList[0]
					if len(parentCat.Subcategories) != tt.expectedCategoryLen {
						t.Errorf("Expected %d subcategories, got %d", tt.expectedCategoryLen, len(parentCat.Subcategories))
					}
					if !slices.ContainsFunc(parentCat.Subcategories, func(c BaseCategory) bool {
						return c.Title == tt.newCategory.Title
					}) {
						t.Errorf("New subcategory %q not found in the list", tt.newCategory.Title)
					}
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error: %q, but got none", tt.expectedErr)
				}
				if err.Error() != tt.expectedErr {
					t.Errorf("Error mismatch. Got: %q, Expected: %q", err.Error(), tt.expectedErr)
				}
			}
		})
	}
}
