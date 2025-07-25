package main

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
