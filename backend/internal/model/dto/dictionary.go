package dto

type DictionaryItem struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type DictionaryResponse struct {
	MediaTypes     []DictionaryItem `json:"media_types,omitempty"`
	RequestStatuses []DictionaryItem `json:"request_statuses,omitempty"`
	SourceStatuses []DictionaryItem `json:"source_statuses,omitempty"`
}
