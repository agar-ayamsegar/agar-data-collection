package domain

type SuccessResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	TraceID string `json:"trace_id"`
}
