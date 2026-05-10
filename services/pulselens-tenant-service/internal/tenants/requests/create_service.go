package requests

type CreateServiceRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Environment string                 `json:"environment" validate:"required"`
	Tags        map[string]interface{} `json:"tags"`
}
