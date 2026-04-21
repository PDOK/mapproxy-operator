package controller

import (
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	"github.com/pdok/mapproxy-operator/internal/controller/mapperutils"
	smoothoperatorutils "github.com/pdok/smooth-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func mutateHorizontalPodAutoscaler(r *WMTSReconciler, obj *pdoknlv2.WMTS, autoscaler *autoscalingv2.HorizontalPodAutoscaler) error {
	reconcilerClient := r.Client
	labels := smoothoperatorutils.CloneOrEmptyMap(obj.GetLabels())
	if err := smoothoperatorutils.SetImmutableLabels(reconcilerClient, autoscaler, labels); err != nil {
		return err
	}

	autoscaler.Spec.MaxReplicas = 30
	autoscaler.Spec.MinReplicas = smoothoperatorutils.Pointer(int32(2))
	autoscaler.Spec.ScaleTargetRef = autoscalingv2.CrossVersionObjectReference{
		APIVersion: appsv1.SchemeGroupVersion.String(),
		Kind:       "Deployment",
		Name:       getSuffixedName(obj, constants.MapproxyName),
	}

	var averageCPU int32 = 90
	if cpu := mapperutils.GetContainerResourceRequest(obj, constants.MapproxyName, corev1.ResourceCPU); cpu != nil {
		averageCPU = 80
	}
	autoscaler.Spec.Metrics = []autoscalingv2.MetricSpec{{
		Type: autoscalingv2.ResourceMetricSourceType,
		Resource: &autoscalingv2.ResourceMetricSource{
			Name: corev1.ResourceCPU,
			Target: autoscalingv2.MetricTarget{
				Type:               autoscalingv2.UtilizationMetricType,
				AverageUtilization: &averageCPU,
			},
		},
	}}

	var behaviourStabilizationWindowSeconds int32

	autoscaler.Spec.Behavior = &autoscalingv2.HorizontalPodAutoscalerBehavior{
		ScaleUp: &autoscalingv2.HPAScalingRules{
			StabilizationWindowSeconds: &behaviourStabilizationWindowSeconds,
			Policies: []autoscalingv2.HPAScalingPolicy{{
				Type:          autoscalingv2.PodsScalingPolicy,
				Value:         20,
				PeriodSeconds: 60,
			}},
			SelectPolicy: smoothoperatorutils.Pointer(autoscalingv2.MaxChangePolicySelect),
		},
		ScaleDown: &autoscalingv2.HPAScalingRules{
			StabilizationWindowSeconds: smoothoperatorutils.Pointer(int32(3600)),
			Policies: []autoscalingv2.HPAScalingPolicy{
				{
					Type:          autoscalingv2.PercentScalingPolicy,
					Value:         10,
					PeriodSeconds: 600,
				},
				{
					Type:          autoscalingv2.PodsScalingPolicy,
					Value:         1,
					PeriodSeconds: 600,
				},
			},
			SelectPolicy: smoothoperatorutils.Pointer(autoscalingv2.MaxChangePolicySelect),
		},
	}
	if obj.Spec.HorizontalPodAutoscalerPatch != nil {
		patchedSpec, err := smoothoperatorutils.StrategicMergePatch(&autoscaler.Spec, obj.Spec.HorizontalPodAutoscalerPatch)
		if err != nil {
			return err
		}
		autoscaler.Spec = *patchedSpec
	}
	if err := smoothoperatorutils.EnsureSetGVK(r.Client, autoscaler, autoscaler); err != nil {
		return err
	}
	return ctrl.SetControllerReference(obj, autoscaler, r.Scheme)
}

func getBareHorizontalPodAutoScaler(obj *pdoknlv2.WMTS) *autoscalingv2.HorizontalPodAutoscaler {
	return &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getSuffixedName(obj, constants.MapproxyName),
			Namespace: obj.GetNamespace(),
		},
	}
}
