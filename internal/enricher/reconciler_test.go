package enricher

import (
	"slices"
	"testing"
	"time"

	"github.com/0xmzn/awelist/internal/types"
)

func TestReconciler_Reconcile(t *testing.T) {
	reconciler := NewReconciler()

	now := time.Now()

	tests := []struct {
		name       string
		yamlList   types.AwesomeList
		jsonList   types.AwesomeList
		wantURLs   []string
		checkOrder bool
	}{
		{
			name: "All new links (empty lock file)",
			yamlList: types.AwesomeList{
				{Links: []*types.Link{{URL: "https://a.com"}, {URL: "https://b.com"}}},
			},
			jsonList: nil,
			wantURLs: []string{"https://a.com", "https://b.com"},
		},
		{
			name: "All fresh links (no action needed)",
			yamlList: types.AwesomeList{
				{Links: []*types.Link{{URL: "https://a.com"}}},
			},
			jsonList: types.AwesomeList{
				{Links: []*types.Link{{
					URL:          "https://a.com",
					RepoMetadata: &types.GitRepoMetadata{EnrichedAt: now.Add(-1 * time.Hour)},
				}}},
			},
			wantURLs: []string{},
		},
		{
			name: "Stale link (older than 24h)",
			yamlList: types.AwesomeList{
				{Links: []*types.Link{{URL: "https://old.com"}}},
			},
			jsonList: types.AwesomeList{
				{Links: []*types.Link{{
					URL:          "https://old.com",
					RepoMetadata: &types.GitRepoMetadata{EnrichedAt: now.Add(-25 * time.Hour)},
				}}},
			},
			wantURLs: []string{"https://old.com"},
		},
		{
			name: "Mixed: New, New2, Fresh, and Stale",
			yamlList: types.AwesomeList{
				{Links: []*types.Link{
					{URL: "https://new.com"},
					{URL: "https://new2.com"},
					{URL: "https://fresh.com"},
					{URL: "https://stale.com"},
				}},
			},
			jsonList: types.AwesomeList{
				{Links: []*types.Link{
					{
						URL:          "https://fresh.com",
						RepoMetadata: &types.GitRepoMetadata{EnrichedAt: now.Add(-1 * time.Hour)},
					},
					{
						URL:          "https://stale.com",
						RepoMetadata: &types.GitRepoMetadata{EnrichedAt: now.Add(-48 * time.Hour)},
					},
				}},
			},
			wantURLs:   []string{"https://new.com", "https://new2.com", "https://stale.com"},
			checkOrder: true,
		},
		{
			name: "Link removed from YAML",
			yamlList: types.AwesomeList{
				{Links: []*types.Link{{URL: "https://keep.com"}}},
			},
			jsonList: types.AwesomeList{
				{Links: []*types.Link{
					{URL: "https://keep.com", RepoMetadata: &types.GitRepoMetadata{EnrichedAt: now}},
					{URL: "https://deleted.com", RepoMetadata: &types.GitRepoMetadata{EnrichedAt: now.Add(-100 * time.Hour)}},
				}},
			},
			wantURLs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinks := reconciler.Reconcile(tt.yamlList, tt.jsonList)

			var gotURLs []string
			for _, l := range gotLinks {
				gotURLs = append(gotURLs, l.URL)
			}

			if len(gotURLs) != len(tt.wantURLs) {
				t.Errorf("Reconcile() returned %d items, want %d", len(gotURLs), len(tt.wantURLs))
			}

			for _, want := range tt.wantURLs {
				if !slices.Contains(gotURLs, want) {
					t.Errorf("Reconcile() missing expected URL: %s", want)
				}
			}

			if tt.checkOrder {
				newIdx := slices.Index(gotURLs, "https://new.com")
				new2Idx := slices.Index(gotURLs, "https://new2.com")
				staleIdx := slices.Index(gotURLs, "https://stale.com")

				if newIdx == -1 || staleIdx == -1 {
					t.Fatal("Missing expected links for order check")
				}

				if !(newIdx == 0 && new2Idx == 1 && staleIdx == 2) {
					t.Errorf("Expected links to be ordered.")
				}

				if newIdx > staleIdx {
					t.Errorf("Expected New links before Stale links. Got order: %v", gotURLs)
				}
			}
		})
	}
}
