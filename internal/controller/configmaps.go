package controller

import (
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/capabilitiesgenerator"
	"github.com/pdok/mapproxy-operator/internal/controller/mapproxy"
	smoothoperatorutils "github.com/pdok/smooth-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// capabilities generator config map
	capabilitiesGeneratorInputFileName = "input.yaml"
	// mapproxy config map
	mapproxyIncludeFileName  = "include.conf"
	mapproxyConfigFileName   = "mapproxy.yaml"
	mapproxyResponseFileName = "response.lua"
)

func mutateConfigMapCapabilitiesGenerator(r *WMTSReconciler, obj *pdoknlv2.WMTS, configMap *corev1.ConfigMap) error {
	reconcilerClient := r.Client

	labels := addCommonLabels(obj, smoothoperatorutils.CloneOrEmptyMap(obj.GetLabels()))
	if err := smoothoperatorutils.SetImmutableLabels(reconcilerClient, configMap, labels); err != nil {
		return err
	}

	if len(configMap.Data) == 0 {
		input, err := capabilitiesgenerator.GetInput(obj)
		if err != nil {
			return err
		}
		configMap.Data = map[string]string{capabilitiesGeneratorInputFileName: input}
	}
	configMap.Immutable = smoothoperatorutils.Pointer(true)

	if err := smoothoperatorutils.EnsureSetGVK(reconcilerClient, configMap, configMap); err != nil {
		return err
	}
	if err := ctrl.SetControllerReference(obj, configMap, r.Scheme); err != nil {
		return err
	}
	return smoothoperatorutils.AddHashSuffix(configMap)
}

func mutateConfigMapMapProxy(r *WMTSReconciler, obj *pdoknlv2.WMTS, configMap *corev1.ConfigMap) error {
	reconcilerClient := r.Client

	labels := addCommonLabels(obj, smoothoperatorutils.CloneOrEmptyMap(obj.GetLabels()))
	if err := smoothoperatorutils.SetImmutableLabels(reconcilerClient, configMap, labels); err != nil {
		return err
	}

	if len(configMap.Data) == 0 {
		configMap.Data = map[string]string{}
		includeConfig, err := mapproxy.GetInclude(obj)
		if err != nil {
			return err
		}
		configMap.Data[mapproxyIncludeFileName] = includeConfig
		mapproxyConfig, err := mapproxy.GetMapproxyConfig(obj)
		if err != nil {
			return err
		}
		configMap.Data[mapproxyConfigFileName] = mapproxyConfig
		responseConfig, err := mapproxy.GetResponse(obj)
		if err != nil {
			return err
		}
		configMap.Data[mapproxyResponseFileName] = responseConfig
	}
	configMap.Immutable = smoothoperatorutils.Pointer(true)

	if err := smoothoperatorutils.EnsureSetGVK(reconcilerClient, configMap, configMap); err != nil {
		return err
	}
	if err := ctrl.SetControllerReference(obj, configMap, r.Scheme); err != nil {
		return err
	}
	return smoothoperatorutils.AddHashSuffix(configMap)
}

func getBareConfigMap(obj *pdoknlv2.WMTS, name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getSuffixedName(obj, name),
			Namespace: obj.GetNamespace(),
		},
	}
}
