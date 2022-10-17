package v1alpha1

//
//import (
//	"github.com/kotalco/community-api/pkg/k8s"
//	"reflect"
//
//	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
//	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
//	apierrors "k8s.io/apimachinery/pkg/api/errors"
//	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//)
//
//const (
//	CRDPlural   string = "sslconfigs"
//	CRDGroup    string = "blog.velotio.com"
//	CRDVersion  string = "v1alpha1"
//	FullCRDName string = CRDPlural + "." + CRDGroup
//)
//
//func CreateCRD(clientset apiextension.Interface) error {
//	crd := &apiextensionv1beta1.CustomResourceDefinition{
//		ObjectMeta: meta_v1.ObjectMeta{Name: FullCRDName},
//		Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
//			Group:   CRDGroup,
//			Version: CRDVersion,
//			Scope:   apiextensionv1beta1.NamespaceScoped,
//			Names: apiextensionv1beta1.CustomResourceDefinitionNames{
//				Plural: CRDPlural,
//				Kind:   reflect.TypeOf(SslConfig{}).Name(),
//			},
//		},
//	}
//
//	_, err := k8s.Clientset().ExtensionsV1beta1().Ingresses("default").Create()
//	if err != nil && apierrors.IsAlreadyExists(err) {
//		return nil
//	}
//	return err
//}
