package endpoint

import (
	"github.com/kotalco/cloud-api/pkg/k8s/ingressroute"
	restErrors "github.com/kotalco/community-api/pkg/errors"
	"github.com/stretchr/testify/assert"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
	"testing"
)

var (
	endpointService        IService
	ingressRouteCreateFunc func(dto *ingressroute.IngressRouteDto) *restErrors.RestErr
	ingressRouteListFunc   func(namesapce string) (*traefikv1alpha1.IngressRouteList, *restErrors.RestErr)
	ingressRouteGetFunc    func(name string, namespace string) (*traefikv1alpha1.IngressRoute, *restErrors.RestErr)
	ingressRouteDeleteFunc func(name string, namespace string) *restErrors.RestErr
)

type ingressRouteServiceMock struct{}

func (i ingressRouteServiceMock) Create(dto *ingressroute.IngressRouteDto) *restErrors.RestErr {
	return ingressRouteCreateFunc(dto)
}

func (i ingressRouteServiceMock) List(namesapce string) (*traefikv1alpha1.IngressRouteList, *restErrors.RestErr) {
	return ingressRouteListFunc(namesapce)
}

func (i ingressRouteServiceMock) Get(name string, namespace string) (*traefikv1alpha1.IngressRoute, *restErrors.RestErr) {
	return ingressRouteGetFunc(name, namespace)
}

func (i ingressRouteServiceMock) Delete(name string, namespace string) *restErrors.RestErr {
	return ingressRouteDeleteFunc(name, namespace)
}

func TestMain(m *testing.M) {
	ingressRoutesService = &ingressRouteServiceMock{}
	endpointService = NewService()
	code := m.Run()
	os.Exit(code)
}

func TestService_Create(t *testing.T) {
	t.Run("create endpoint should pass", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) *restErrors.RestErr {
			return nil
		}

		createDto := &CreateEndpointDto{}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{}},
		}}
		err := endpointService.Create(createDto, svc, "")
		assert.Nil(t, err)
	})
	t.Run("create endpoint should throw if ingressRoute.create throws", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}

		createDto := &CreateEndpointDto{}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{}},
		}}
		err := endpointService.Create(createDto, svc, "")
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
		assert.EqualValues(t, "something went wrong", err.Message)
	})
	t.Run("create endpoint should throw if service has no invalid ports", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) *restErrors.RestErr {
			return nil
		}

		createDto := &CreateEndpointDto{}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{},
		}}
		err := endpointService.Create(createDto, svc, "")
		assert.EqualValues(t, http.StatusBadRequest, err.Status)
		assert.EqualValues(t, "service invalid ports", err.Message)
	})
}

func TestService_List(t *testing.T) {
	t.Run("list endpoints should pass", func(t *testing.T) {
		ingressRouteListFunc = func(namesapce string) (*traefikv1alpha1.IngressRouteList, *restErrors.RestErr) {
			return &traefikv1alpha1.IngressRouteList{
				TypeMeta: metav1.TypeMeta{},
				ListMeta: metav1.ListMeta{},
				Items:    []traefikv1alpha1.IngressRoute{{}},
			}, nil
		}

		list, err := endpointService.List("namespace")
		assert.Nil(t, err)
		assert.NotNil(t, list)
	})

	t.Run("list endpoint should throw if ingressroutesService.list throws", func(t *testing.T) {
		ingressRouteListFunc = func(namesapce string) (*traefikv1alpha1.IngressRouteList, *restErrors.RestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}
		list, err := endpointService.List("namespace")
		assert.Nil(t, list)
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
		assert.EqualValues(t, "something went wrong", err.Message)
	})
}

func TestService_Get(t *testing.T) {
	t.Run("get endpoint should pass", func(t *testing.T) {
		ingressRouteGetFunc = func(name string, namespace string) (*traefikv1alpha1.IngressRoute, *restErrors.RestErr) {
			return &traefikv1alpha1.IngressRoute{}, nil
		}
		record, err := endpointService.Get("name", "namespace")
		assert.Nil(t, err)
		assert.NotNil(t, record)

	})

	t.Run("get endpoint should throw if ingressroute.get throws", func(t *testing.T) {
		ingressRouteGetFunc = func(name string, namespace string) (*traefikv1alpha1.IngressRoute, *restErrors.RestErr) {
			return nil, restErrors.NewNotFoundError("no such record")
		}
		record, err := endpointService.Get("name", "namespace")
		assert.Nil(t, record)
		assert.EqualValues(t, http.StatusNotFound, err.Status)
		assert.EqualValues(t, "no such record", err.Message)
	})

}

func TestService_Delete(t *testing.T) {
	t.Run("delete endpoint should pass", func(t *testing.T) {
		ingressRouteDeleteFunc = func(name string, namespace string) *restErrors.RestErr {
			return nil
		}

		err := endpointService.Delete("name", "namespace")
		assert.Nil(t, err)
	})

	t.Run("delete ednpoint should throw if ingressrouteService.delete throws", func(t *testing.T) {
		ingressRouteDeleteFunc = func(name string, namespace string) *restErrors.RestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		err := endpointService.Delete("name", "namespace")
		assert.EqualValues(t, http.StatusInternalServerError, err.Status)
		assert.EqualValues(t, "something went wrong", err.Message)
	})
}
