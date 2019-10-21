package activemqartemis

import (
	"context"
	brokerv2alpha1 "github.com/rh-messaging/activemq-artemis-operator/pkg/apis/broker/v2alpha1"
	"github.com/rh-messaging/activemq-artemis-operator/pkg/logs"
	"github.com/rh-messaging/activemq-artemis-operator/version"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	buildv1 "github.com/openshift/api/build/v1"
	oimagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

var log = logf.Log.WithName("controller_activemqartemis")
var log2 = logs.GetLogger("kieapp.controller")

var namespacedNameToFSM = make(map[types.NamespacedName]*ActiveMQArtemisFSM)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ActiveMQArtemis Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileActiveMQArtemis{client: mgr.GetClient(), scheme: mgr.GetScheme(), result: reconcile.Result{Requeue: false}}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("activemqartemis-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ActiveMQArtemis
	err = c.Watch(&source.Kind{Type: &brokerv2alpha1.ActiveMQArtemis{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ActiveMQArtemis
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &brokerv2alpha1.ActiveMQArtemis{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner ActiveMQArtemis
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &brokerv2alpha1.ActiveMQArtemis{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileActiveMQArtemis{}

// ReconcileActiveMQArtemis reconciles a ActiveMQArtemis object
type ReconcileActiveMQArtemis struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	result reconcile.Result
}

// Reconcile reads that state of the cluster for a ActiveMQArtemis object and makes changes based on the state read
// and what is in the ActiveMQArtemis.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileActiveMQArtemis) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	// Log where we are and what we're doing
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ActiveMQArtemis")

	var err error = nil
	var namespacedNameFSM *ActiveMQArtemisFSM = nil
	var amqbfsm *ActiveMQArtemisFSM = nil

	instance := &brokerv2alpha1.ActiveMQArtemis{}
	namespacedName := types.NamespacedName{
		Name:      request.Name,
		Namespace: request.Namespace,
	}

	// Fetch the ActiveMQArtemis instance
	// When first creating this will have err == nil
	// When deleting after creation this will have err NotFound
	// When deleting before creation reconcile won't be called
	if err = r.client.Get(context.TODO(), request.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("ActiveMQArtemis Controller Reconcile encountered a IsNotFound, checking to see if we should delete namespacedName tracking for request NamespacedName " + request.NamespacedName.String())

			// See if we have been tracking this NamespacedName
			if namespacedNameFSM = namespacedNameToFSM[namespacedName]; namespacedNameFSM != nil {
				reqLogger.Info("Removing namespacedName tracking for " + namespacedName.String())
				// If so we should no longer track it
				amqbfsm = namespacedNameToFSM[namespacedName]
				amqbfsm.Exit()
				delete(namespacedNameToFSM, namespacedName)
				amqbfsm = nil
			}

			// Setting err to nil to prevent requeue
			err = nil
		} else {
			reqLogger.Error(err, "ActiveMQArtemis Controller Reconcile errored thats not IsNotFound, requeuing request", "Request Namespace", request.Namespace, "Request Name", request.Name)
			// Leaving err as !nil causes requeue
		}

		// Add error detail for use later
		return r.result, err
	}

	reqLogger.Info("Reconciling ActiveMQArtemis", "Operator version", version.Version, "ActiveMQArtemis release", instance.Spec.Version)

	minor, micro, err := checkProductUpgrade(instance)
	if err != nil {
		//return  err
	}

	minorVersion := getMinorImageVersion(instance.Spec.Version)
	latestMinorVersion := getMinorImageVersion(CurrentVersion)
	if (micro && minorVersion == latestMinorVersion) ||
		(minor && minorVersion != latestMinorVersion) {
		/*if err := getConfigVersionDiffs(customResource.Spec.Version, CurrentVersion, service); err != nil {
			return api.Environment{}, err
		}*/
		// reset current annotations and update CR use to latest product version
		instance.SetAnnotations(map[string]string{})
		instance.Spec.Version = CurrentVersion
	}

	/*//Create a list of objects that should be deployed
		requestedResources := resource.KubernetesResource()
		for index := range requestedResources {
			requestedResources[index].SetNamespace(instance.Namespace)
		}
*/

	//Obtain a list of objects that are actually deployed
	deployed, err := r.getDeployedResources(instance)
	if err != nil {

		return reconcile.Result{}, err
	}
	reqLogger.Info("Reconciling ActiveMQArtemis", "Operator version", version.Version, "Deployed Resouce length", len(deployed))


	// Do lookup to see if we have a fsm for the incoming name in the incoming namespace
	// if not, create it
	// for the given fsm, do an update
	// - update first level sets? what if the operator has gone away and come back? stateless?
	if namespacedNameFSM = namespacedNameToFSM[namespacedName]; namespacedNameFSM == nil {
		// TODO: Fix multiple fsm's post ENTMQBR-2875
		if len(namespacedNameToFSM) > 0 {
			reqLogger.Info("ActiveMQArtemis Controller Reconcile does not yet support more than one custom resource instance per namespace!")
			return r.result, nil
		}
		amqbfsm = NewActiveMQArtemisFSM(instance, namespacedName, r)
		namespacedNameToFSM[namespacedName] = amqbfsm

		// Enter the first state; atm CreatingK8sResourcesState
		amqbfsm.Enter(CreatingK8sResourcesID)
	} else {
		amqbfsm = namespacedNameFSM
		*amqbfsm.customResource = *instance
		err, _ = amqbfsm.Update()
	}

	// Single exit, return the result and error condition
	return r.result, err
}

func (reconciler *ReconcileActiveMQArtemis) getDeployedResources(instance *brokerv2alpha1.ActiveMQArtemis) (map[reflect.Type][]resource.KubernetesResource, error) {
	log := log2.With("kind", instance.Kind, "name", instance.Name, "namespace", instance.Namespace)

	resourceMap := make(map[reflect.Type][]resource.KubernetesResource)

	listOps := &client.ListOptions{Namespace: instance.Namespace}

	statefulSetList := &appsv1.StatefulSetList{}
	err := reconciler.client.List(context.TODO(), listOps, statefulSetList)
	if err != nil {
		log.Warn("Failed to list statefulSets. ", err)
		return nil, err
	}
	var statefulSets []resource.KubernetesResource
	for index := range statefulSetList.Items {
		statefulSet := statefulSetList.Items[index]
		for _, ownerRef := range statefulSet.GetOwnerReferences() {
			if ownerRef.UID == instance.UID {
				statefulSets = append(statefulSets, &statefulSet)
				break
			}
		}
	}
	resourceMap[reflect.TypeOf(appsv1.StatefulSet{})] = statefulSets

	pvcList := &corev1.PersistentVolumeClaimList{}
	err = reconciler.client.List(context.TODO(), listOps, pvcList)
	if err != nil {
		log.Warn("Failed to list PersistentVolumeClaims. ", err)
		return nil, err
	}
	var pvcs []resource.KubernetesResource
	for index := range pvcList.Items {
		pvc := pvcList.Items[index]
		for _, ownerRef := range pvc.GetOwnerReferences() {
			if ownerRef.UID == instance.UID {
				pvcs = append(pvcs, &pvc)
				break
			}
		}
	}
	resourceMap[reflect.TypeOf(corev1.PersistentVolumeClaim{})] = pvcs

	saList := &corev1.ServiceAccountList{}
	err = reconciler.client.List(context.TODO(), listOps, saList)
	if err != nil {
		log.Warn("Failed to list ServiceAccounts. ", err)
		return nil, err
	}
	var sas []resource.KubernetesResource
	for index := range saList.Items {
		sa := saList.Items[index]
		for _, ownerRef := range sa.GetOwnerReferences() {
			if ownerRef.UID == instance.UID {
				sas = append(sas, &sa)
				break
			}
		}
	}
	resourceMap[reflect.TypeOf(corev1.ServiceAccount{})] = sas

	//secretList := &corev1.SecretList{}
	var secrets []resource.KubernetesResource
	//err = reconciler.Service.List(context.TODO(), listOps, secretList) //TODO: can't list secrets due bug:
	// https://github.com/kubernetes-sigs/controller-runtime/issues/362
	// multiple group-version-kinds associated with type *api.SecretList, refusing to guess at one
	// Will work around by loading known secrets instead

	for _, res := range statefulSets {
		dc := res.(*appsv1.StatefulSet)
		for _, volume := range dc.Spec.Template.Spec.Volumes {
			if volume.Secret != nil {
				name := volume.Secret.SecretName
				secret := &corev1.Secret{}
				err := reconciler.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: instance.GetNamespace()}, secret)
				if err != nil && !errors.IsNotFound(err) {
					log.Warn("Failed to load Secret", err)
					return nil, err
				}
				secrets = append(secrets, secret)
			}
		}
	}
	//if err != nil {
	//	log.Info("Failed to list Secrets. ", err)
	//	return nil, err
	//}
	//for index := range secretList.Items {
	//	secret := secretList.Items[index]
	//	for _, ownerRef := range secret.GetOwnerReferences() {
	//		if ownerRef.UID == instance.UID {
	//			secrets = append(secrets, &secret)
	//			break
	//		}
	//	}
	//}
	resourceMap[reflect.TypeOf(corev1.Secret{})] = secrets

	roleList := &rbacv1.RoleList{}
	err = reconciler.client.List(context.TODO(), listOps, roleList)
	if err != nil {
		log.Warn("Failed to list roles. ", err)
		return nil, err
	}
	var roles []resource.KubernetesResource
	for index := range roleList.Items {
		role := roleList.Items[index]
		for _, ownerRef := range role.GetOwnerReferences() {
			if ownerRef.UID == instance.UID {
				roles = append(roles, &role)
				break
			}
		}
	}
	resourceMap[reflect.TypeOf(rbacv1.Role{})] = roles

	roleBindingList := &rbacv1.RoleBindingList{}
	err = reconciler.client.List(context.TODO(), listOps, roleBindingList)
	if err != nil {
		log.Warn("Failed to list roleBindings. ", err)
		return nil, err
	}
	var roleBindings []resource.KubernetesResource
	for index := range roleBindingList.Items {
		roleBinding := roleBindingList.Items[index]
		for _, ownerRef := range roleBinding.GetOwnerReferences() {
			if ownerRef.UID == instance.UID {
				roleBindings = append(roleBindings, &roleBinding)
				break
			}
		}
	}
	resourceMap[reflect.TypeOf(rbacv1.RoleBinding{})] = roleBindings

	serviceList := &corev1.ServiceList{}
	err = reconciler.client.List(context.TODO(), listOps, serviceList)
	if err != nil {
		log.Warn("Failed to list services. ", err)
		return nil, err
	}
	var services []resource.KubernetesResource
	for index := range serviceList.Items {
		service := serviceList.Items[index]
		for _, ownerRef := range service.GetOwnerReferences() {
			if ownerRef.UID == instance.UID {
				services = append(services, &service)
				break
			}
		}
	}
	resourceMap[reflect.TypeOf(corev1.Service{})] = services

	RouteList := &routev1.RouteList{}
	err = reconciler.client.List(context.TODO(), listOps, RouteList)
	if err != nil {
		log.Warn("Failed to list routes. ", err)
		return nil, err
	}
	var routes []resource.KubernetesResource
	for index := range RouteList.Items {
		route := RouteList.Items[index]
		for _, ownerRef := range route.GetOwnerReferences() {
			if ownerRef.UID == instance.UID {
				routes = append(routes, &route)
				break
			}
		}
	}
	resourceMap[reflect.TypeOf(routev1.Route{})] = routes

	imageStreamList := &oimagev1.ImageStreamList{}
	err = reconciler.client.List(context.TODO(), listOps, imageStreamList)
	if err != nil {
		log.Warn("Failed to list imageStreams. ", err)
		return nil, err
	}
	var imageStreams []resource.KubernetesResource
	for index := range imageStreamList.Items {
		imageStream := imageStreamList.Items[index]
		for _, ownerRef := range imageStream.GetOwnerReferences() {
			if ownerRef.UID == instance.UID {
				imageStreams = append(imageStreams, &imageStream)
				break
			}
		}
	}
	resourceMap[reflect.TypeOf(oimagev1.ImageStream{})] = imageStreams

	buildConfigList := &buildv1.BuildConfigList{}
	err = reconciler.client.List(context.TODO(), listOps, buildConfigList)
	if err != nil {
		log.Warn("Failed to list buildConfigs. ", err)
		return nil, err
	}
	var buildConfigs []resource.KubernetesResource
	for index := range buildConfigList.Items {
		buildConfig := buildConfigList.Items[index]
		for _, ownerRef := range buildConfig.GetOwnerReferences() {
			if ownerRef.UID == instance.UID {
				buildConfigs = append(buildConfigs, &buildConfig)
				break
			}
		}
	}
	resourceMap[reflect.TypeOf(buildv1.BuildConfig{})] = buildConfigs

	return resourceMap, nil
}
