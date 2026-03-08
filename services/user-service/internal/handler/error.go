package handler

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorDetails struct {
	Type    string       `json:"type"`
	Status  int          `json:"status"`
	Error   string       `json:"error"`
	Details []FieldError `json:"details"`
}
