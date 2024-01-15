package endpoint

import (
	"github.com/kotalco/core-api/k8s/ingressroute"
	"github.com/kotalco/core-api/k8s/middleware"
	"github.com/kotalco/core-api/k8s/secret"
	restErrors "github.com/kotalco/core-api/pkg/errors"
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
	ingressRouteCreateFunc func(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr)
	ingressRouteListFunc   func(ns string, labels map[string]string) (*traefikv1alpha1.IngressRouteList, restErrors.IRestErr)
	ingressRouteGetFunc    func(name string, namespace string) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr)
	ingressRouteDeleteFunc func(name string, namespace string) restErrors.IRestErr
	ingressRouteUpdateFunc func(record *traefikv1alpha1.IngressRoute) restErrors.IRestErr
)

type ingressRouteServiceMock struct{}

func (i ingressRouteServiceMock) Create(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
	return ingressRouteCreateFunc(dto)
}

func (i ingressRouteServiceMock) List(ns string, labels map[string]string) (*traefikv1alpha1.IngressRouteList, restErrors.IRestErr) {
	return ingressRouteListFunc(ns, labels)
}

func (i ingressRouteServiceMock) Get(name string, namespace string) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
	return ingressRouteGetFunc(name, namespace)
}

func (i ingressRouteServiceMock) Delete(name string, namespace string) restErrors.IRestErr {
	return ingressRouteDeleteFunc(name, namespace)
}
func (i ingressRouteServiceMock) Update(record *traefikv1alpha1.IngressRoute) restErrors.IRestErr {
	return ingressRouteUpdateFunc(record)
}

type k8MiddlewareServiceMock struct{}

var (
	k8middlewareCreateFunc func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr
	k8middlewareGetFunc    func(name string, namespace string) (*traefikv1alpha1.Middleware, restErrors.IRestErr)
)

func (k k8MiddlewareServiceMock) Create(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
	return k8middlewareCreateFunc(dto)
}
func (k k8MiddlewareServiceMock) Get(name string, namespace string) (*traefikv1alpha1.Middleware, restErrors.IRestErr) {
	return k8middlewareGetFunc(name, namespace)
}

type secretServiceMock struct{}

var (
	secretCreateFunc func(dto *secret.CreateSecretDto) restErrors.IRestErr
	secretGetFunc    func(name string, namespace string) (*corev1.Secret, restErrors.IRestErr)
)

func (s secretServiceMock) Create(dto *secret.CreateSecretDto) restErrors.IRestErr {
	return secretCreateFunc(dto)
}
func (s secretServiceMock) Get(name string, namespace string) (*corev1.Secret, restErrors.IRestErr) {
	return secretGetFunc(name, namespace)
}

func TestMain(m *testing.M) {
	ingressRoutesService = &ingressRouteServiceMock{}
	k8MiddlewareService = &k8MiddlewareServiceMock{}
	secretService = &secretServiceMock{}
	endpointService = NewService()
	code := m.Run()
	os.Exit(code)
}

func TestService_Create(t *testing.T) {
	t.Run("create endpoint should pass", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       traefikv1alpha1.IngressRouteSpec{},
			}, nil
		}
		k8middlewareCreateFunc = func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
			return nil
		}

		k8middlewareGetFunc = func(name string, namespace string) (*traefikv1alpha1.Middleware, restErrors.IRestErr) {
			return new(traefikv1alpha1.Middleware), nil
		}

		createDto := &CreateEndpointDto{}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Name: "api"}},
		}}

		err := endpointService.Create(createDto, svc)
		assert.Nil(t, err)
	})
	t.Run("create endpoint should pass with basic auth", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       traefikv1alpha1.IngressRouteSpec{},
			}, nil
		}
		k8middlewareCreateFunc = func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
			return nil
		}

		createDto := &CreateEndpointDto{
			UseBasicAuth: true,
		}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{}},
		}}

		secretCreateFunc = func(dto *secret.CreateSecretDto) restErrors.IRestErr {
			return nil
		}
		err := endpointService.Create(createDto, svc)
		assert.Nil(t, err)
	})
	t.Run("create endpoint should throw if can't create secret with basic auth", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       traefikv1alpha1.IngressRouteSpec{},
			}, nil
		}
		k8middlewareCreateFunc = func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
			return nil
		}

		createDto := &CreateEndpointDto{
			UseBasicAuth: true,
		}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{}},
		}}

		secretCreateFunc = func(dto *secret.CreateSecretDto) restErrors.IRestErr {
			return restErrors.NewInternalServerError("can't create secret")
		}
		ingressRouteDeleteFunc = func(name string, namespace string) restErrors.IRestErr {
			return nil
		}
		err := endpointService.Create(createDto, svc)
		assert.EqualValues(t, "can't create secret", err.Error())
	})
	t.Run("create endpoint should throw if can't create secret with basic auth and delete ingressRoute failed", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       traefikv1alpha1.IngressRouteSpec{},
			}, nil
		}
		k8middlewareCreateFunc = func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
			return nil
		}

		createDto := &CreateEndpointDto{
			UseBasicAuth: true,
		}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{}},
		}}

		secretCreateFunc = func(dto *secret.CreateSecretDto) restErrors.IRestErr {
			return restErrors.NewInternalServerError("can't create secret")
		}
		ingressRouteDeleteFunc = func(name string, namespace string) restErrors.IRestErr {
			return restErrors.NewInternalServerError("can't delete ingress route")
		}
		err := endpointService.Create(createDto, svc)
		assert.EqualValues(t, "can't delete ingress route", err.Error())
	})
	t.Run("create endpoint should throw if can't can't roll back after creating failed", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       traefikv1alpha1.IngressRouteSpec{},
			}, nil
		}
		k8middlewareCreateFunc = func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
			return nil
		}

		secretCreateFunc = func(dto *secret.CreateSecretDto) restErrors.IRestErr {
			return nil
		}
		k8middlewareCreateFunc = func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
			return restErrors.NewInternalServerError("can't create basic auth middleware")
		}
		ingressRouteDeleteFunc = func(name string, namespace string) restErrors.IRestErr {
			return nil
		}

		createDto := &CreateEndpointDto{UseBasicAuth: true}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Name: "api"}},
		}}

		err := endpointService.Create(createDto, svc)
		assert.EqualValues(t, "can't create basic auth middleware", err.Error())
	})

	t.Run("create endpoint should throw if ingressRoute.create throws", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}

		createDto := &CreateEndpointDto{}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{}},
		}}
		err := endpointService.Create(createDto, svc)
		assert.EqualValues(t, http.StatusInternalServerError, err.StatusCode())
		assert.EqualValues(t, "something went wrong", err.Error())
	})
	t.Run("create endpoint should throw if k8middleware service throws", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       traefikv1alpha1.IngressRouteSpec{},
			}, nil
		}
		ingressRouteDeleteFunc = func(name string, namespace string) restErrors.IRestErr {
			return nil
		}
		k8middlewareCreateFunc = func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		createDto := &CreateEndpointDto{}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{}},
		}}
		err := endpointService.Create(createDto, svc)
		assert.EqualValues(t, "something went wrong", err.Error())
	})
	t.Run("create endpoint should throw if can't create prefix middleware and delete fails", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       traefikv1alpha1.IngressRouteSpec{},
			}, nil
		}
		ingressRouteDeleteFunc = func(name string, namespace string) restErrors.IRestErr {
			return nil
		}
		k8middlewareCreateFunc = func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		ingressRouteDeleteFunc = func(name string, namespace string) restErrors.IRestErr {
			return restErrors.NewInternalServerError("can't delete ingress route")
		}
		createDto := &CreateEndpointDto{}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{}},
		}}
		err := endpointService.Create(createDto, svc)
		assert.EqualValues(t, "can't delete ingress route", err.Error())
	})
	t.Run("create endpoint should throw if can't find crossover middleware and it throws error rather than notfound", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       traefikv1alpha1.IngressRouteSpec{},
			}, nil
		}
		k8middlewareCreateFunc = func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
			return nil
		}

		createDto := &CreateEndpointDto{}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{}},
		}}

		k8middlewareGetFunc = func(name string, namespace string) (*traefikv1alpha1.Middleware, restErrors.IRestErr) {
			return nil, restErrors.NewInternalServerError("can't find crossover internal err")
		}

		ingressRouteDeleteFunc = func(name string, namespace string) restErrors.IRestErr {
			return nil
		}

		err := endpointService.Create(createDto, svc)
		assert.EqualValues(t, "can't find crossover internal err", err.Error())
	})
	t.Run("create endpoint should throw if can't find crossover middleware and can't create new one", func(t *testing.T) {
		ingressRouteCreateFunc = func(dto *ingressroute.IngressRouteDto) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       traefikv1alpha1.IngressRouteSpec{},
			}, nil
		}
		k8middlewareCreateFunc = func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
			return nil
		}

		createDto := &CreateEndpointDto{}
		svc := &corev1.Service{Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{}},
		}}

		k8middlewareGetFunc = func(name string, namespace string) (*traefikv1alpha1.Middleware, restErrors.IRestErr) {
			return nil, restErrors.NewNotFoundError("no such record")
		}

		k8middlewareCreateFunc = func(dto *middleware.CreateMiddlewareDto) restErrors.IRestErr {
			return restErrors.NewInternalServerError("can't create middleware")
		}
		ingressRouteDeleteFunc = func(name string, namespace string) restErrors.IRestErr {
			return nil
		}
		err := endpointService.Create(createDto, svc)
		assert.EqualValues(t, "can't create middleware", err.Error())
	})

}

func TestService_List(t *testing.T) {
	t.Run("list endpoints should pass", func(t *testing.T) {
		ingressRouteListFunc = func(ns string, labels map[string]string) (*traefikv1alpha1.IngressRouteList, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRouteList{
				TypeMeta: metav1.TypeMeta{},
				ListMeta: metav1.ListMeta{},
				Items: []traefikv1alpha1.IngressRoute{{
					Spec: traefikv1alpha1.IngressRouteSpec{
						Routes: []traefikv1alpha1.Route{{
							Services: []traefikv1alpha1.Service{{}},
						}},
					},
				}},
			}, nil
		}
		secretGetFunc = func(name string, namespace string) (*corev1.Secret, restErrors.IRestErr) {
			return &corev1.Secret{}, nil
		}

		list, err := endpointService.List("namespace", map[string]string{"app.kubernetes.io/created-by": "kotal-api"})

		assert.Nil(t, err)
		assert.NotNil(t, list)
	})

	t.Run("list endpoint should throw if ingressroutesService.list throws", func(t *testing.T) {
		ingressRouteListFunc = func(ns string, labels map[string]string) (*traefikv1alpha1.IngressRouteList, restErrors.IRestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}
		list, err := endpointService.List("namespace", map[string]string{})
		assert.Nil(t, list)
		assert.EqualValues(t, http.StatusInternalServerError, err.StatusCode())
		assert.EqualValues(t, "something went wrong", err.Error())
	})
}

func TestService_Get(t *testing.T) {
	t.Run("get endpoint should pass", func(t *testing.T) {
		ingressRouteGetFunc = func(name string, namespace string) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRoute{
				Spec: traefikv1alpha1.IngressRouteSpec{
					Routes: []traefikv1alpha1.Route{{
						Services: []traefikv1alpha1.Service{{}},
					}},
				},
			}, nil
		}
		secretGetFunc = func(name string, namespace string) (*corev1.Secret, restErrors.IRestErr) {
			return &corev1.Secret{}, nil
		}
		record, err := endpointService.Get("name", "namespace")
		assert.Nil(t, err)
		assert.NotNil(t, record)

	})

	t.Run("get endpoint should throw if ingressroute.get throws", func(t *testing.T) {
		ingressRouteGetFunc = func(name string, namespace string) (*traefikv1alpha1.IngressRoute, restErrors.IRestErr) {
			return nil, restErrors.NewNotFoundError("no such record")
		}
		record, err := endpointService.Get("name", "namespace")
		assert.Nil(t, record)
		assert.EqualValues(t, http.StatusNotFound, err.StatusCode())
		assert.EqualValues(t, "no such record", err.Error())
	})

}

func TestService_Delete(t *testing.T) {
	t.Run("delete endpoint should pass", func(t *testing.T) {
		ingressRouteDeleteFunc = func(name string, namespace string) restErrors.IRestErr {
			return nil
		}

		err := endpointService.Delete("name", "namespace")
		assert.Nil(t, err)
	})

	t.Run("delete ednpoint should throw if ingressrouteService.delete throws", func(t *testing.T) {
		ingressRouteDeleteFunc = func(name string, namespace string) restErrors.IRestErr {
			return restErrors.NewInternalServerError("something went wrong")
		}
		err := endpointService.Delete("name", "namespace")
		assert.EqualValues(t, http.StatusInternalServerError, err.StatusCode())
		assert.EqualValues(t, "something went wrong", err.Error())
	})
}

func TestService_Count(t *testing.T) {
	t.Run("count endpoints should pass", func(t *testing.T) {
		ingressRouteListFunc = func(ns string, labels map[string]string) (*traefikv1alpha1.IngressRouteList, restErrors.IRestErr) {
			return &traefikv1alpha1.IngressRouteList{Items: []traefikv1alpha1.IngressRoute{{}}}, nil
		}
		count, err := endpointService.Count("default", map[string]string{})
		assert.Nil(t, err)
		assert.EqualValues(t, 1, count)
	})

	t.Run("count endpoint should throw ingressRouteService throws", func(t *testing.T) {
		ingressRouteListFunc = func(ns string, labels map[string]string) (*traefikv1alpha1.IngressRouteList, restErrors.IRestErr) {
			return nil, restErrors.NewInternalServerError("something went wrong")
		}
		count, err := endpointService.Count("default", map[string]string{})
		assert.EqualValues(t, 0, count)
		assert.EqualValues(t, "something went wrong", err.Error())
	})
}
