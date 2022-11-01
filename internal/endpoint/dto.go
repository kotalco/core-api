package endpoint

import (
	"github.com/go-playground/validator/v10"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"strings"
)

type CreateEndpointDto struct {
	Name        string `json:"name" validate:"required"`
	ServiceName string `json:"service_name" validate:"required"`
}

type EndpointDto struct {
	Name   string   `json:"name"`
	Routes []string `json:"routes"`
}

func Validate(dto interface{}) *restErrors.RestErr {
	newValidator := validator.New()
	err := newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Name":
				fields["name"] = "invalid name"
				break
			case "ServiceName":
				fields["service_name"] = "invalid service_name"
				break
			}
		}
		if len(fields) > 0 {
			return restErrors.NewValidationError(fields)
		}
	}

	return nil
}

func (endpoint *EndpointDto) Marshall(dtoIngressRoute *traefikv1alpha1.IngressRoute) *EndpointDto {
	endpoint.Name = dtoIngressRoute.Name
	for _, route := range dtoIngressRoute.Spec.Routes {
		str := strings.ReplaceAll(route.Match, "Host(`", "")
		str = strings.ReplaceAll(str, "`)", "")
		str = strings.ReplaceAll(str, " && ", "")
		str = strings.ReplaceAll(str, "PathPrefix(`", "")
		endpoint.Routes = append(endpoint.Routes, str)
	}

	return endpoint
}
