package list

import (
	"slices"
	"strings"
	"testing"

	"github.com/0xmzn/awelist/internal/types"
)

func mockAwesomeList() types.AwesomeList {
	return types.AwesomeList{
		{
			Title: "Category A",
			Links: []*types.Link{
				{Title: "Link A1", URL: "http://example.com/a1"},
			},
			Subcategories: []*types.Category{
				{
					Title: "Subcategory A1",
					Links: []*types.Link{
						{Title: "Link A1.1", URL: "http://example.com/a1-1"},
					},
				},
			},
		},
		{
			Title: "Category B",
			Links: []*types.Link{
				{Title: "Link B1", URL: "http://example.com/b1"},
			},
		},
	}
}

func areLinksEqual(a, b []*types.Link) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Title != b[i].Title || a[i].URL != b[i].URL {
			return false
		}
	}
	return true
}

func TestService_AddLink(t *testing.T) {
	tests := []struct {
		name          string
		path          []string
		newLink       *types.Link
		expectedErr   string
		expectedLinks []*types.Link
	}{
		{
			name: "Add link to a top-level category",
			path: []string{"Category A"},
			newLink: &types.Link{
				Title: "New Link A",
				URL:   "http://example.com/newa",
			},
			expectedErr: "",
			expectedLinks: []*types.Link{
				{Title: "Link A1", URL: "http://example.com/a1"},
				{Title: "New Link A", URL: "http://example.com/newa"},
			},
		},
		{
			name:        "Add link that already exists with different title but same url",
			path:        []string{"Category A"},
			newLink:     &types.Link{Title: "Link A2", URL: "http://example.com/a1"},
			expectedErr: "link with url \"http://example.com/a1\" already exists in \"Category A\"",
			expectedLinks: []*types.Link{
				{Title: "Link A1", URL: "http://example.com/a1"},
			},
		},
		{
			name: "Add link to a subcategory",
			path: []string{"Category A", "Subcategory A1"},
			newLink: &types.Link{
				Title: "New Sublink A1",
				URL:   "http://example.com/newsuba1",
			},
			expectedErr: "",
			expectedLinks: []*types.Link{
				{Title: "Link A1.1", URL: "http://example.com/a1-1"},
				{Title: "New Sublink A1", URL: "http://example.com/newsuba1"},
			},
		},
		{
			name: "Add link to non-existent top-level category",
			path: []string{"Nonexistent Category"},
			newLink: &types.Link{
				Title: "New Link",
				URL:   "http://example.com/new",
			},
			expectedErr: "category \"Nonexistent Category\" not found",
		},
		{
			name: "Add link to non-existent subcategory",
			path: []string{"Category B", "Nonexistent Subcategory"},
			newLink: &types.Link{
				Title: "New Link",
				URL:   "http://example.com/new",
			},
			expectedErr: "category \"Nonexistent Subcategory\" not found",
		},
		{
			name: "Add link with empty path",
			path: []string{},
			newLink: &types.Link{
				Title: "Empty Path Link",
				URL:   "http://example.com/empty",
			},
			expectedErr: "path cannot be empty",
		},
		{
			name: "Add link with a duplicate title to top-level category",
			path: []string{"Category A"},
			newLink: &types.Link{
				Title: "Link A1",
				URL:   "http://example.com/a1",
			},
			expectedErr: "link with title \"Link A1\" already exists in \"Category A\"",
			expectedLinks: []*types.Link{
				{Title: "Link A1", URL: "http://example.com/a1"},
			},
		},
		{
			name: "Add link with a duplicate title to subcategory",
			path: []string{"Category A", "Subcategory A1"},
			newLink: &types.Link{
				Title: "Link A1.1",
				URL:   "http://example.com/a1-1",
			},
			expectedErr: "link with title \"Link A1.1\" already exists in \"Subcategory A1\"",
			expectedLinks: []*types.Link{
				{Title: "Link A1.1", URL: "http://example.com/a1-1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := mockAwesomeList()
			mngr := NewManager()

			err := mngr.AddLink(list, tt.newLink, tt.path)

			if tt.expectedErr == "" {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				var targetLinks []*types.Link
				if len(tt.path) == 1 {
					targetLinks = list[0].Links
				} else if len(tt.path) == 2 {
					targetLinks = list[0].Subcategories[0].Links
				}

				if !areLinksEqual(targetLinks, tt.expectedLinks) {
					t.Errorf("Links mismatch.\nGot: %+v\nExpected: %+v", targetLinks, tt.expectedLinks)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error: %q, but got none", tt.expectedErr)
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Error mismatch.\nGot: %q\nExpected content: %q", err.Error(), tt.expectedErr)
				}
			}
		})
	}
}

func TestService_AddCategory(t *testing.T) {
	tests := []struct {
		name                string
		path                []string
		newCategory         *types.Category
		expectedErr         string
		expectedCategoryLen int
	}{
		{
			name:                "Add top-level category",
			path:                []string{},
			newCategory:         &types.Category{Title: "New Category"},
			expectedErr:         "",
			expectedCategoryLen: 3,
		},
		{
			name:                "Add subcategory to a top-level category",
			path:                []string{"Category A"},
			newCategory:         &types.Category{Title: "New Subcategory"},
			expectedErr:         "",
			expectedCategoryLen: 2,
		},
		{
			name:        "Add subcategory to non-existent parent",
			path:        []string{"Nonexistent Parent"},
			newCategory: &types.Category{Title: "New Category"},
			expectedErr: "category \"Nonexistent Parent\" not found",
		},
		{
			name:        "Add top-level category with a duplicate title",
			path:        []string{},
			newCategory: &types.Category{Title: "Category A"},
			expectedErr: "category \"Category A\" already exists at root",
		},
		{
			name:        "Add subcategory with a duplicate title",
			path:        []string{"Category A"},
			newCategory: &types.Category{Title: "Subcategory A1"},
			expectedErr: "subcategory \"Subcategory A1\" already exists under \"Category A\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := mockAwesomeList()
			mngr := NewManager()

			err := mngr.AddCategory(&list, tt.newCategory, tt.path)

			if tt.expectedErr == "" {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if len(tt.path) == 0 {
					if len(list) != tt.expectedCategoryLen {
						t.Errorf("Expected %d top-level categories, got %d", tt.expectedCategoryLen, len(list))
					}
					exists := slices.ContainsFunc(list, func(c *types.Category) bool {
						return c.Title == tt.newCategory.Title
					})
					if !exists {
						t.Errorf("New category %q not found in the list", tt.newCategory.Title)
					}
				} else {
					parentCat := list[0]
					if len(parentCat.Subcategories) != tt.expectedCategoryLen {
						t.Errorf("Expected %d subcategories, got %d", tt.expectedCategoryLen, len(parentCat.Subcategories))
					}
					exists := slices.ContainsFunc(parentCat.Subcategories, func(c *types.Category) bool {
						return c.Title == tt.newCategory.Title
					})
					if !exists {
						t.Errorf("New subcategory %q not found in the list", tt.newCategory.Title)
					}
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error: %q, but got none", tt.expectedErr)
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Error mismatch.\nGot:      %q\nExpected content: %q", err.Error(), tt.expectedErr)
				}
			}
		})
	}
}
