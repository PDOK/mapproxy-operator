package controller

import (
	"regexp"
	"strings"

	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller/constants"
	smoothoperatormodel "github.com/pdok/smooth-operator/model"
	uptimeutils "github.com/pdok/smooth-operator/pkg/uptime-utils"
	smoothoperatorutils "github.com/pdok/smooth-operator/pkg/util"
	traefikiov1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
)

var setUptimeOperatorAnnotations = true

func SetUptimeOperatorAnnotations(set bool) {
	setUptimeOperatorAnnotations = set
}

func getBareIngressRoute(obj *pdoknlv2.WMTS) *traefikiov1alpha1.IngressRoute {
	return &traefikiov1alpha1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getSuffixedName(obj, constants.MapproxyName),
			Namespace: obj.GetNamespace(),
		},
	}
}

func mutateIngressRoute(r *WMTSReconciler, obj *pdoknlv2.WMTS, ingressRoute *traefikiov1alpha1.IngressRoute) error {
	reconcilerClient := r.Client

	labels := smoothoperatorutils.CloneOrEmptyMap(obj.GetLabels())
	if err := smoothoperatorutils.SetImmutableLabels(reconcilerClient, ingressRoute, labels); err != nil {
		return err
	}

	if setUptimeOperatorAnnotations {

		queryString := obj.Spec.HealthCheck.Querystring
		ingressRoute.Annotations = uptimeutils.GetUptimeAnnotations(
			obj.GetAnnotations(),
			obj.TypedName(),
			getUptimeName(obj),
			obj.Spec.Service.BaseURL.String()+"?"+queryString,
			obj.GetLabels(),
		)
	}

	mapserverService := traefikiov1alpha1.Service{
		LoadBalancerSpec: traefikiov1alpha1.LoadBalancerSpec{
			Name: getBareService(obj).GetName(),
			Kind: "Service",
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: constants.MapserverPortNr,
			},
		},
	}

	middlewareRef := traefikiov1alpha1.MiddlewareRef{
		Name: getBareCorsHeadersMiddleware(obj).GetName(),
	}

	makeRoute := func(match string, service traefikiov1alpha1.Service, middlewareRef traefikiov1alpha1.MiddlewareRef) traefikiov1alpha1.Route {
		return traefikiov1alpha1.Route{
			Kind:        "Rule",
			Match:       match,
			Services:    []traefikiov1alpha1.Service{service},
			Middlewares: []traefikiov1alpha1.MiddlewareRef{middlewareRef},
		}
	}

	ingressRoute.Spec.Routes = []traefikiov1alpha1.Route{}
	for _, ingressRouteURL := range obj.Spec.IngressRouteURLs {
		ingressRoute.Spec.Routes = append(ingressRoute.Spec.Routes, makeRoute(getMatchRule(ingressRouteURL.URL), mapserverService, middlewareRef))
	}

	if err := smoothoperatorutils.EnsureSetGVK(reconcilerClient, ingressRoute, ingressRoute); err != nil {
		return err
	}
	return ctrl.SetControllerReference(obj, ingressRoute, r.Scheme)
}

// getUptimeName transforms the CR name into a uptime.pdok.nl/name value
// owner-dataset-v1-0 -> OWNER dataset v1_0 [INSPIRE] [WMS|WFS]
func getUptimeName(obj *pdoknlv2.WMTS) string {
	// Extract the version from the CR name, owner-dataset-v1-0 -> owner-dataset + v1-0
	versionMatcher := regexp.MustCompile("^(.*)(?:-(v?[1-9](?:-[0-9])?))?$")
	match := versionMatcher.FindStringSubmatch(obj.GetName())

	nameParts := strings.Split(match[1], "-")
	nameParts[0] = strings.ToUpper(nameParts[0])

	// Add service version if found
	if len(match) > 2 && len(match[2]) > 0 {
		nameParts = append(nameParts, strings.ReplaceAll(match[2], "-", "_"))
	}

	return strings.Join(append(nameParts, "wmts"), " ")
}

func getMatchRule(url smoothoperatormodel.URL) string {
	host := url.Hostname()
	if strings.Contains(host, "localhost") {
		return "Host(`localhost`) && Path(`" + url.Path + "`)"
	}

	return "(Host(`localhost`) || Host(`" + host + "`)) && Path(`" + url.Path + "`)"
}

func getLegendMatchRule(url smoothoperatormodel.URL) string {
	host := url.Hostname()
	if strings.Contains(host, "localhost") {
		return "Host(`localhost`) && PathPrefix(`" + url.Path + "/legend`)"
	}

	return "(Host(`localhost`) || Host(`" + host + "`)) && PathPrefix(`" + url.Path + "/legend`)"
}
