package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	"github.com/pdok/mapproxy-operator/internal/controller/types"
	"github.com/pdok/smooth-operator/pkg/status"
	smoothoperatorutils "github.com/pdok/smooth-operator/pkg/util"
	"github.com/pkg/errors"
	traefikiov1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type WMTSReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Images types.Images
}

// +kubebuilder:rbac:groups=pdok.nl,resources=wmts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pdok.nl,resources=wmts/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=pdok.nl,resources=wmts/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configmaps;services,verbs=watch;create;get;update;list;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=watch;list;get
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=watch;create;get;update;list;delete
// +kubebuilder:rbac:groups=traefik.io,resources=ingressroutes;middlewares,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=create;update;delete;list;watch
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets/status,verbs=get;update
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets/finalizers,verbs=update
func (r *WMTSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	lgr := log.FromContext(ctx)
	lgr.Info("Starting reconcile for WMTS resource", "name", req.NamespacedName)

	// Fetch the WMTS instance
	wmts := &pdoknlv2.WMTS{}
	if err = r.Get(ctx, req.NamespacedName, wmts); err != nil {
		if apierrors.IsNotFound(err) {
			lgr.Info("WMTS resource not found", "name", req.NamespacedName)
		} else {
			lgr.Error(err, "unable to fetch WMTS resource", "error", err)
		}
		return result, client.IgnoreNotFound(err)
	}

	// Recover from a panic so we can add the error to the status of the Atom
	defer func() {
		if rec := recover(); rec != nil {
			err = recoveredPanicToError(rec)
			status.LogAndUpdateStatusError(ctx, r.Client, wmts, err)
		}
	}()

	// Check TTL, delete if expired
	if ttlExpired(wmts) {
		err = r.Delete(ctx, wmts)

		return result, err
	}

	ensureLabel(wmts, "pdok.nl/service-type", "wmts")

	lgr.Info("creating resources for wmts", "wmts", wmts.Name)
	operationResults, err := createOrUpdateAllForWMTS(ctx, r, wmts)
	if err != nil {
		lgr.Info("failed creating resources for wmts", "wmts", wmts.Name)
		status.LogAndUpdateStatusError(ctx, r.Client, wmts, err)
		return result, err
	}
	lgr.Info("finished creating resources for wmts", "wmts", wmts.Name)
	status.LogAndUpdateStatusFinished(ctx, r.Client, wmts, operationResults)

	return result, err
}

func recoveredPanicToError(rec any) (err error) {
	switch x := rec.(type) {
	case string:
		err = errors.New(x)
	case error:
		err = x
	default:
		err = errors.New("unknown panic")
	}

	// Add stack
	// TODO - this doesn't seem to work, see if there is a better method to add the stack
	err = errors.WithStack(err)

	return
}

func ttlExpired(obj *pdoknlv2.WMTS) bool {
	lifecycle := obj.Spec.Lifecycle
	if lifecycle != nil && lifecycle.TTLInDays != nil {
		expiresAt := obj.GetCreationTimestamp().Add(time.Duration(*lifecycle.TTLInDays) * 24 * time.Hour)

		return expiresAt.Before(time.Now())
	}

	return false
}

func ensureLabel(obj *pdoknlv2.WMTS, key, value string) {
	labels := obj.GetLabels()
	if _, ok := labels[key]; !ok {
		labels[key] = value
	}

	obj.SetLabels(labels)
}

func createOrUpdateAllForWMTS(ctx context.Context, r *WMTSReconciler, obj *pdoknlv2.WMTS) (operationResults map[string]controllerutil.OperationResult, err error) {
	reconcilerClient := r.Client

	hashedConfigMapNames, operationResults, err := createOrUpdateConfigMaps(ctx, r, obj)
	if err != nil {
		return operationResults, err
	}

	// region Deployment
	{
		deployment := getBareDeployment(obj)
		operationResults[smoothoperatorutils.GetObjectFullName(reconcilerClient, deployment)], err = controllerutil.CreateOrUpdate(ctx, reconcilerClient, deployment, func() error {
			return mutateDeployment(r, obj, deployment, hashedConfigMapNames)
		})
		if err != nil && !strings.Contains(err.Error(), "the object has been modified; please apply your changes to the latest version and try again") {
			return operationResults, fmt.Errorf("unable to create/update resource %s: %w", smoothoperatorutils.GetObjectFullName(reconcilerClient, deployment), err)
		}
	}
	// end region Deployment

	// region TraefikMiddleware
	if obj.Spec.Options.IncludeIngress {
		middleware := getBareCorsHeadersMiddleware(obj)
		operationResults[smoothoperatorutils.GetObjectFullName(reconcilerClient, middleware)], err = controllerutil.CreateOrUpdate(ctx, reconcilerClient, middleware, func() error {
			return mutateCorsHeadersMiddleware(r, obj, middleware)
		})
		if err != nil {
			return operationResults, fmt.Errorf("unable to create/update resource %s: %w", smoothoperatorutils.GetObjectFullName(reconcilerClient, middleware), err)
		}
	}
	// end region TraefikMiddleware

	// region PodDisruptionBudget
	{
		err = createOrUpdateOrDeletePodDisruptionBudget(ctx, r, obj, operationResults)
		if err != nil {
			return operationResults, err
		}
	}
	// end region PodDisruptionBudget

	// region HorizontalAutoScaler
	{
		autoscaler := getBareHorizontalPodAutoScaler(obj)
		operationResults[smoothoperatorutils.GetObjectFullName(reconcilerClient, autoscaler)], err = controllerutil.CreateOrUpdate(ctx, reconcilerClient, autoscaler, func() error {
			return mutateHorizontalPodAutoscaler(r, obj, autoscaler)
		})
		if err != nil {
			return operationResults, fmt.Errorf("unable to create/update resource %s: %w", smoothoperatorutils.GetObjectFullName(reconcilerClient, autoscaler), err)
		}
	}
	// end region HorizontalAutoScaler

	// region IngressRoute
	if obj.Spec.Options.IncludeIngress {
		ingress := getBareIngressRoute(obj)
		operationResults[smoothoperatorutils.GetObjectFullName(reconcilerClient, ingress)], err = controllerutil.CreateOrUpdate(ctx, reconcilerClient, ingress, func() error {
			return mutateIngressRoute(r, obj, ingress)
		})
		if err != nil {
			return operationResults, fmt.Errorf("unable to create/update resource %s: %w", smoothoperatorutils.GetObjectFullName(reconcilerClient, ingress), err)
		}
	}
	// end region IngressRoute

	// region Service
	{
		service := getBareService(obj)
		operationResults[smoothoperatorutils.GetObjectFullName(reconcilerClient, service)], err = controllerutil.CreateOrUpdate(ctx, reconcilerClient, service, func() error {
			return mutateService(r, obj, service)
		})
		if err != nil {
			return operationResults, fmt.Errorf("unable to create/update resource %s: %w", smoothoperatorutils.GetObjectFullName(reconcilerClient, service), err)
		}
	}
	// end region Service

	return operationResults, nil
}

func createOrUpdateConfigMaps(ctx context.Context, r *WMTSReconciler, obj *pdoknlv2.WMTS) (hashedConfigMapNames types.HashedConfigMapNames, operationResults map[string]controllerutil.OperationResult, err error) { //nolint:revive
	operationResults, configMaps := make(map[string]controllerutil.OperationResult), make(map[string]func(*WMTSReconciler, *pdoknlv2.WMTS, *corev1.ConfigMap) error)
	configMaps[constants.CapabilitiesGeneratorName] = func(r *WMTSReconciler, o *pdoknlv2.WMTS, cm *corev1.ConfigMap) error {
		return mutateConfigMapCapabilitiesGenerator(r, o, cm)
	}

	configMaps[constants.MapproxyName] = func(reconciler *WMTSReconciler, o *pdoknlv2.WMTS, cm *corev1.ConfigMap) error {
		return mutateConfigMapMapProxy(r, o, cm)
	}

	for cmName, mutate := range configMaps {
		cm, or, err := createOrUpdateConfigMap(ctx, r, obj, cmName, func(r *WMTSReconciler, o *pdoknlv2.WMTS, cm *corev1.ConfigMap) error {
			return mutate(r, o, cm)
		})
		if or != nil {
			operationResults[smoothoperatorutils.GetObjectFullName(r.Client, cm)] = *or
		}
		if err != nil {
			return hashedConfigMapNames, operationResults, err
		}
		switch cmName {
		case constants.CapabilitiesGeneratorName:
			hashedConfigMapNames.CapabilitiesGenerator = cm.Name
		case constants.MapproxyName:
			hashedConfigMapNames.Mapproxy = cm.Name
		}
	}

	return hashedConfigMapNames, operationResults, err
}

func mutateConfigMap(r *WMTSReconciler, obj *pdoknlv2.WMTS, configMap *corev1.ConfigMap) error { //nolint:unused
	reconcilerClient := r.Client
	labels := smoothoperatorutils.CloneOrEmptyMap(obj.GetLabels())
	if err := smoothoperatorutils.SetImmutableLabels(reconcilerClient, configMap, labels); err != nil {
		return err
	}

	configMap.Immutable = smoothoperatorutils.Pointer(true)
	configMap.Data = map[string]string{}
	//nolint:gocritic
	//updateConfigMapWithStaticFiles(configMap, obj)

	if err := smoothoperatorutils.EnsureSetGVK(reconcilerClient, configMap, configMap); err != nil {
		return err
	}
	if err := ctrl.SetControllerReference(obj, configMap, r.Scheme); err != nil {
		return err
	}
	return smoothoperatorutils.AddHashSuffix(configMap)
}

func getBareDeployment(obj *pdoknlv2.WMTS) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: getSuffixedName(obj, constants.MapproxyName),
			// name might become too long. not handling here. will just fail on apply.
			Namespace: obj.GetNamespace(),
		},
	}
}

func getSuffixedName(obj *pdoknlv2.WMTS, suffix string) string {
	return obj.TypedName() + "-" + suffix
}

func createOrUpdateConfigMap(ctx context.Context, reconciler *WMTSReconciler, obj *pdoknlv2.WMTS, name string, mutate func(*WMTSReconciler, *pdoknlv2.WMTS, *corev1.ConfigMap) error) (*corev1.ConfigMap, *controllerutil.OperationResult, error) {
	reconcilerClient := reconciler.Client
	cm := getBareConfigMap(obj, name)
	if err := mutate(reconciler, obj, cm); err != nil {
		return cm, nil, err
	}
	or, err := controllerutil.CreateOrUpdate(ctx, reconcilerClient, cm, func() error {
		return mutate(reconciler, obj, cm)
	})
	if err != nil {
		return cm, &or, fmt.Errorf("unable to create/update resource %s: %w", smoothoperatorutils.GetObjectFullName(reconcilerClient, cm), err)
	}
	return cm, &or, nil
}

func createOrUpdateOrDeletePodDisruptionBudget(ctx context.Context, reconciler *WMTSReconciler, obj *pdoknlv2.WMTS, operationResults map[string]controllerutil.OperationResult) (err error) {
	reconcilerClient := reconciler.Client
	podDisruptionBudget := getBarePodDisruptionBudget(obj)
	autoscalerPatch := obj.Spec.HorizontalPodAutoscalerPatch
	if autoscalerPatch != nil && autoscalerPatch.MinReplicas != nil && autoscalerPatch.MaxReplicas != nil &&
		*autoscalerPatch.MinReplicas == 1 && *autoscalerPatch.MaxReplicas == 1 {
		err = reconcilerClient.Delete(ctx, podDisruptionBudget)
		if err == nil {
			operationResults[smoothoperatorutils.GetObjectFullName(reconcilerClient, podDisruptionBudget)] = "deleted"
		}
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("unable to delete resource %s: %w", smoothoperatorutils.GetObjectFullName(reconcilerClient, podDisruptionBudget), err)
		}
	} else {
		operationResults[smoothoperatorutils.GetObjectFullName(reconcilerClient, podDisruptionBudget)], err = controllerutil.CreateOrUpdate(ctx, reconcilerClient, podDisruptionBudget, func() error {
			return mutatePodDisruptionBudget(reconciler, obj, podDisruptionBudget)
		})
		if err != nil {
			return fmt.Errorf("unable to create/update resource %s: %w", smoothoperatorutils.GetObjectFullName(reconcilerClient, podDisruptionBudget), err)
		}
	}
	return nil
}

func addCommonLabels(obj *pdoknlv2.WMTS, labels map[string]string) map[string]string { //nolint:revive
	return labels
}

// SetupWithManager sets up the controller with the Manager.
func (r *WMTSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return createControllerManager(mgr, &pdoknlv2.WMTS{}).Complete(r)
}

func createControllerManager(mgr ctrl.Manager, obj client.Object) *builder.TypedBuilder[reconcile.Request] {
	kind := "WMTS"

	controllerMgr := ctrl.NewControllerManagedBy(mgr).For(obj).Named(strings.ToLower(kind))
	controllerMgr.Owns(&corev1.ConfigMap{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&appsv1.Deployment{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.Service{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&traefikiov1alpha1.Middleware{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&traefikiov1alpha1.IngressRoute{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&policyv1.PodDisruptionBudget{}, builder.WithPredicates(predicate.GenerationChangedPredicate{}))

	return controllerMgr.Watches(&appsv1.ReplicaSet{}, status.GetReplicaSetEventHandlerForObj(mgr, kind))
}
