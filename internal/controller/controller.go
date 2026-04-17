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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type WMTSReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Images types.Images
}

func (r *WMTSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	lgr := log.FromContext(ctx)
	lgr.Info("Starting reconcile for WMTS resource", "name", req.NamespacedName)

	// Fetch the WFS instance
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
	lgr.Info("finished creating resources for wfs", "wfs", wmts.Name)
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
	if obj.Options().IncludeIngress {
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

func createOrUpdateConfigMaps(ctx context.Context, r *WMTSReconciler, obj *pdoknlv2.WMTS) (hashedConfigMapNames types.HashedConfigMapNames, operationResults map[string]controllerutil.OperationResult, err error) {
	operationResults, configMaps := make(map[string]controllerutil.OperationResult), make(map[string]func(*WMTSReconciler, *pdoknlv2.WMTS, *corev1.ConfigMap) error)
	configMaps[constants.MapserverName] = mutateConfigMap
	if obj.Mapfile() == nil {
		configMaps[constants.MapfileGeneratorName] = func(r *WMTSReconciler, o *pdoknlv2.WMTS, cm *corev1.ConfigMap) error {
			return mutateConfigMapMapfileGenerator(r, o, cm)
		}
	}
	configMaps[constants.CapabilitiesGeneratorName] = func(r *WMTSReconciler, o *pdoknlv2.WMTS, cm *corev1.ConfigMap) error {
		return mutateConfigMapCapabilitiesGenerator(r, o, cm)
	}

	for cmName, mutate := range configMaps {
		cm, or, err := createOrUpdateConfigMap(ctx, obj, r, cmName, func(r *WMTSReconciler, o *pdoknlv2.WMTS, cm *corev1.ConfigMap) error {
			return mutate(r, o, cm)
		})
		if or != nil {
			operationResults[smoothoperatorutils.GetObjectFullName(r.Client, cm)] = *or
		}
		if err != nil {
			return hashedConfigMapNames, operationResults, err
		}
		switch cmName {
		case constants.MapserverName:
			hashedConfigMapNames.Mapserver = cm.Name
		case constants.MapfileGeneratorName:
			hashedConfigMapNames.MapfileGenerator = cm.Name
		case constants.CapabilitiesGeneratorName:
			hashedConfigMapNames.CapabilitiesGenerator = cm.Name
		case constants.InitScriptsName:
			hashedConfigMapNames.InitScripts = cm.Name
		case constants.LegendGeneratorName:
			hashedConfigMapNames.LegendGenerator = cm.Name
		case constants.FeatureinfoGeneratorName:
			hashedConfigMapNames.FeatureInfoGenerator = cm.Name
		case constants.OgcWebserviceProxyName:
			hashedConfigMapNames.OgcWebserviceProxy = cm.Name
		}
	}

	return hashedConfigMapNames, operationResults, err
}

func mutateConfigMap(r *WMTSReconciler, obj *pdoknlv2.WMTS, configMap *corev1.ConfigMap) error {
	reconcilerClient := r.Client
	labels := smoothoperatorutils.CloneOrEmptyMap(obj.GetLabels())
	if err := smoothoperatorutils.SetImmutableLabels(reconcilerClient, configMap, labels); err != nil {
		return err
	}

	configMap.Immutable = smoothoperatorutils.Pointer(true)
	configMap.Data = map[string]string{}

	updateConfigMapWithStaticFiles(configMap, obj)

	if err := smoothoperatorutils.EnsureSetGVK(reconcilerClient, configMap, configMap); err != nil {
		return err
	}
	if err := ctrl.SetControllerReference(obj, configMap, getReconcilerScheme(r)); err != nil {
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

func getBareConfigMap(obj *pdoknlv2.WMTS, name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getSuffixedName(obj, name),
			Namespace: obj.GetNamespace(),
		},
	}
}
