package controller

import (
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	smoothoperatorutils "github.com/pdok/smooth-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	mapserverWebserviceProxyPortNr = 9111
	metricPortName                 = "metric"
)

func getBareService(obj *pdoknlv2.WMTS) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getSuffixedName(obj, constants.MapproxyName),
			Namespace: obj.GetNamespace(),
		},
	}
}

func mutateService(r *WMTSReconciler, obj *pdoknlv2.WMTS, service *corev1.Service) error {
	reconcilerClient := r.Client

	labels := addCommonLabels(obj, smoothoperatorutils.CloneOrEmptyMap(obj.GetLabels()))
	selector := smoothoperatorutils.CloneOrEmptyMap(labels)
	if err := smoothoperatorutils.SetImmutableLabels(reconcilerClient, service, labels); err != nil {
		return err
	}

	ports := []corev1.ServicePort{
		{
			Name:       constants.MapproxyName,
			Port:       constants.MapserverPortNr,
			TargetPort: intstr.FromInt32(constants.MapserverPortNr),
			Protocol:   corev1.ProtocolTCP,
		},
	}

	// Add port here to get the same port order as the odl ansible operator
	ports = append(ports, corev1.ServicePort{
		Name:       metricPortName,
		Port:       constants.ApachePortNr,
		TargetPort: intstr.FromInt32(constants.ApachePortNr),
		Protocol:   corev1.ProtocolTCP,
	})

	service.Spec = corev1.ServiceSpec{
		Type:                  corev1.ServiceTypeClusterIP,
		ClusterIP:             service.Spec.ClusterIP,
		ClusterIPs:            service.Spec.ClusterIPs,
		IPFamilyPolicy:        service.Spec.IPFamilyPolicy,
		IPFamilies:            service.Spec.IPFamilies,
		SessionAffinity:       corev1.ServiceAffinityNone,
		InternalTrafficPolicy: smoothoperatorutils.Pointer(corev1.ServiceInternalTrafficPolicyCluster),
		Ports:                 ports,
		Selector:              selector,
	}
	if err := smoothoperatorutils.EnsureSetGVK(reconcilerClient, service, service); err != nil {
		return err
	}
	return ctrl.SetControllerReference(obj, service, reconcilerClient.Scheme())
}
