package dto

type FieldErrorDTO struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FieldErrorsFromMap(errors map[string]error) []FieldErrorDTO {
	res := make([]FieldErrorDTO, 0, len(errors))
	for key, err := range errors {
		res = append(res, FieldErrorDTO{
			Field:   key,
			Message: err.Error(),
		})
	}
	return res
}

type ErrorDetailsDTO struct {
	Status  int             `json:"status"`
	Error   string          `json:"error"`
	Details []FieldErrorDTO `json:"details,omitempty"`
}
