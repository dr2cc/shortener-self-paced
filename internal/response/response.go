package response

// CreateBatchResponse (ex. ShorteningBatchResult) — результат сокращения при пакетном вводе
type CreateBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
