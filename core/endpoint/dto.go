package endpoint

import (
	"fmt"
	validate "github.com/go-playground/validator/v10"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/kotalco/community-api/pkg/logger"
	"github.com/kotalco/community-api/pkg/shared"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"regexp"
	"strings"
)

type CreateEndpointDto struct {
	Name         string `json:"name" validate:"regexp,lt=64"`
	ServiceName  string `json:"service_name" validate:"required"`
	UseBasicAuth bool   `json:"use_basic_auth"`
}

type EndpointMetaDto struct {
	Name      string `json:"name"`
	Protocol  string `json:"protocol"`
	CreatedAt string `json:"created_at"`
}

type EndpointDto struct {
	EndpointMetaDto
	Routes    []*RouteDto   `json:"routes"`
	BasicAuth *BasicAuthDto `json:"basic_auth,omitempty"`
}

type RouteDto struct {
	Name       string   `json:"name"`
	Route      string   `json:"route"`
	Example    string   `json:"example"`
	References []string `json:"references"`
}

type BasicAuthDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Validate(dto interface{}) restErrors.IRestErr {
	newValidator := validate.New()
	err := newValidator.RegisterValidation("regexp", func(fl validate.FieldLevel) bool {
		re := regexp.MustCompile("^([a-z]|[0-9])([a-z]|[0-9]|-)+([a-z]|[0-9])$")
		return re.MatchString(fl.Field().String())
	})
	if err != nil {
		logger.Warn("ENDPOINT_DTO_VALIDATION", err)
		return restErrors.NewInternalServerError("something went wrong!")
	}
	err = newValidator.Struct(dto)

	if err != nil {
		fields := map[string]string{}
		for _, err := range err.(validate.ValidationErrors) {
			switch err.Field() {
			case "Name":
				fields["name"] = "name must start and end with an alphanumeric, and contains no more than 64 alphanumeric characters and - in total."
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

func (endpoint *EndpointMetaDto) Marshall(dtoIngressRoute *traefikv1alpha1.IngressRoute) *EndpointMetaDto {
	endpoint.Name = dtoIngressRoute.Name
	endpoint.Protocol = dtoIngressRoute.Labels["kotal.io/protocol"]
	endpoint.CreatedAt = dtoIngressRoute.CreationTimestamp.UTC().Format(shared.JavascriptISOString)
	return endpoint
}

func (endpoint *EndpointDto) Marshall(dtoIngressRoute *traefikv1alpha1.IngressRoute, secret *corev1.Secret) *EndpointDto {
	endpoint.Name = dtoIngressRoute.Name
	endpoint.Routes = make([]*RouteDto, 0)
	endpoint.Protocol = dtoIngressRoute.Labels["kotal.io/protocol"]
	endpoint.CreatedAt = dtoIngressRoute.CreationTimestamp.UTC().Format(shared.JavascriptISOString)
	for _, route := range dtoIngressRoute.Spec.Routes {
		str := strings.ReplaceAll(route.Match, "Host(`", "")
		str = strings.ReplaceAll(str, "`)", "")
		str = strings.ReplaceAll(str, " && ", "")
		str = strings.ReplaceAll(str, "PathPrefix(`", "")

		if secret != nil {
			str = fmt.Sprintf("%s:%s@%s", secret.Data["username"], secret.Data["password"], str)
		}
		if route.Services[0].Port.StrVal == "ws" {
			str = fmt.Sprintf("wss://%s", str)
		} else {
			str = fmt.Sprintf("https://%s", str)
		}

		protocol := protocol(endpoint.Protocol)
		kind := kind(dtoIngressRoute.Labels["kotal.io/kind"])
		port := port(route.Services[0].Port.StrVal)
		routeExample := strings.ReplaceAll(examples[protocol][kind][port], `${route}`, str)
		routeRefs := references[protocol][kind][port]
		if routeRefs == nil {
			routeRefs = []string{}
		}
		endpoint.Routes = append(endpoint.Routes, &RouteDto{
			Name:       route.Services[0].Port.StrVal,
			Route:      str,
			Example:    routeExample,
			References: routeRefs,
		})
	}

	return endpoint
}
