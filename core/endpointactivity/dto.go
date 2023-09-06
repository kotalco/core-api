package endpointactivity

import (
	"github.com/go-playground/validator/v10"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"regexp"
)

type ActivityDto struct {
	Name     string      `json:"name"`
	Protocol string      `json:"protocol"`
	Routes   []*RouteDto `json:"routes"`
}

type RouteDto struct {
	Name       string `json:"name"`
	EndpointId string `json:"-"`
	Hits       int    `json:"hits"`
}

type CreateEndpointActivityDto struct {
	RequestId string `json:"request_id" validate:"required"`
}

func Validate(dto interface{}) restErrors.IRestErr {
	newValidator := validator.New()
	err := newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "RequestId":
				fields["request_id"] = "invalid request id"
				break
			}
		}
		if len(fields) > 0 {
			return restErrors.NewValidationError(fields)
		}
	}

	return nil
}

func (dto *ActivityDto) Marshall(dtoIngressRoute *traefikv1alpha1.IngressRoute) *ActivityDto {
	dto.Name = dtoIngressRoute.Name
	dto.Routes = make([]*RouteDto, 0)
	dto.Protocol = dtoIngressRoute.Labels["kotal.io/protocol"]
	for _, route := range dtoIngressRoute.Spec.Routes {

		dto.Routes = append(dto.Routes, &RouteDto{
			Name:       route.Services[0].Port.StrVal,
			EndpointId: getEndpointId(route.Match),
		})
	}
	return dto
}

func getEndpointId(path string) string {
	// Compile the regular expression
	re := regexp.MustCompile("([0-9a-z]{8}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{12})")
	// Find the first match of the pattern in the URL Path
	match := re.FindStringSubmatch(path)

	if len(match) == 0 {
		return ""
	}
	return match[0]
}
