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

func getBareIngressRoute(obj *pdoknlv2.WMTS, suffix string) *traefikiov1alpha1.IngressRoute {
	return &traefikiov1alpha1.IngressRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getSuffixedName(obj, suffix),
			Namespace: obj.GetNamespace(),
		},
	}
}

func mutateDirectIngressRoute(r *WMTSReconciler, obj *pdoknlv2.WMTS, ingressRoute *traefikiov1alpha1.IngressRoute) error {
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

	mapproxyService := getTraefixService(obj, constants.MapserverPortNr)

	middlewareRef := traefikiov1alpha1.MiddlewareRef{
		Name: getBareCorsHeadersMiddleware(obj).GetName(),
	}

	ingressRoute.Spec.Routes = []traefikiov1alpha1.Route{}
	for _, ingressRouteURL := range obj.GetIngressRouteUrls() {
		ingressRoute.Spec.Routes = append(ingressRoute.Spec.Routes, makeRoute(getExactMatchRule(ingressRouteURL), mapproxyService, middlewareRef))
	}

	if err := smoothoperatorutils.EnsureSetGVK(reconcilerClient, ingressRoute, ingressRoute); err != nil {
		return err
	}
	return ctrl.SetControllerReference(obj, ingressRoute, r.Scheme)
}

func mutateRestfulIngressRoute(r *WMTSReconciler, obj *pdoknlv2.WMTS, ingressRoute *traefikiov1alpha1.IngressRoute) error {
	reconcilerClient := r.Client

	labels := smoothoperatorutils.CloneOrEmptyMap(obj.GetLabels())
	if err := smoothoperatorutils.SetImmutableLabels(reconcilerClient, ingressRoute, labels); err != nil {
		return err
	}

	// restful ingress should not be considered for uptime
	ingressRoute.Annotations["uptime.pdok.nl/ignore"] = "-"

	mapproxyService := getTraefixService(obj, constants.MapproxyPortNumber)

	middlewareRef := traefikiov1alpha1.MiddlewareRef{
		Name: getBareCorsHeadersMiddleware(obj).GetName(),
	}

	ingressRoute.Spec.Routes = []traefikiov1alpha1.Route{}
	for _, ingressRouteURL := range obj.GetIngressRouteUrls() {
		ingressRoute.Spec.Routes = append(ingressRoute.Spec.Routes, makeRoute(getPrefixMatchRule(ingressRouteURL), mapproxyService, middlewareRef))
	}

	if err := smoothoperatorutils.EnsureSetGVK(reconcilerClient, ingressRoute, ingressRoute); err != nil {
		return err
	}
	return ctrl.SetControllerReference(obj, ingressRoute, r.Scheme)
}

func getTraefixService(obj *pdoknlv2.WMTS, port int32) traefikiov1alpha1.Service {
	return traefikiov1alpha1.Service{
		LoadBalancerSpec: traefikiov1alpha1.LoadBalancerSpec{
			Name: getBareService(obj).GetName(),
			Kind: "Service",
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: port,
			},
		},
	}
}

func makeRoute(match string, service traefikiov1alpha1.Service, middlewareRef traefikiov1alpha1.MiddlewareRef) traefikiov1alpha1.Route {
	return traefikiov1alpha1.Route{
		Kind:        "Rule",
		Match:       match,
		Services:    []traefikiov1alpha1.Service{service},
		Middlewares: []traefikiov1alpha1.MiddlewareRef{middlewareRef},
	}
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

func getExactMatchRule(url smoothoperatormodel.URL) string {
	host := url.Hostname()
	if strings.Contains(host, "localhost") {
		return "Host(`localhost`) && Path(`" + url.Path + "`)"
	}

	return "(Host(`localhost`) || Host(`" + host + "`)) && Path(`" + url.Path + "`)"
}

func getPrefixMatchRule(url smoothoperatormodel.URL) string {
	host := url.Hostname()
	if strings.Contains(host, "localhost") {
		return "Host(`localhost`) && PathPrefix(`" + url.Path + "`)"
	}

	return "(Host(`localhost`) || Host(`" + host + "`)) && PathPrefix(`" + url.Path + "`)"
}
