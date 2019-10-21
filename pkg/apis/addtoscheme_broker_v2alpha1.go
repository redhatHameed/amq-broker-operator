package apis

import (
	oappsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha1"
	//operatorsv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		v2alpha1.SchemeBuilder.AddToScheme,
		routev1.SchemeBuilder.AddToScheme,
		rbacv1.SchemeBuilder.AddToScheme,
		oappsv1.SchemeBuilder.AddToScheme,
		oimagev1.SchemeBuilder.AddToScheme,
		buildv1.SchemeBuilder.AddToScheme,
		//operatorsv1alpha1.SchemeBuilder.AddToScheme,

	)
}
