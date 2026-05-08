package controller

import (
	"fmt"
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
			Name:      obj.Name + "-wmts-mapproxy" + suffix,
			Namespace: obj.GetNamespace(),
		},
	}
}

func mutateDirectIngressRoute(r *WMTSReconciler, obj *pdoknlv2.WMTS, ingressRoute *traefikiov1alpha1.IngressRoute) error {
	reconcilerClient := r.Client

	labels := addCommonLabels(obj, smoothoperatorutils.CloneOrEmptyMap(obj.GetLabels()))
	if err := smoothoperatorutils.SetImmutableLabels(reconcilerClient, ingressRoute, labels); err != nil {
		return err
	}

	if setUptimeOperatorAnnotations {

		queryString := obj.Spec.HealthCheck.Querystring
		ingressRoute.Annotations = uptimeutils.GetUptimeAnnotations(
			obj.GetAnnotations(),
			obj.TypedName(),
			getUptimeName(obj),
			obj.Spec.Service.BaseURL.String()+"/"+queryString,
			obj.GetLabels(),
		)

		ingressRoute.Annotations["uptime.pdok.nl/tags"] = "public-stats,wmts"
	}

	mapproxyService := getTraefixService(obj, constants.MapproxyPortNumber)

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

	labels := addCommonLabels(obj, smoothoperatorutils.CloneOrEmptyMap(obj.GetLabels()))
	if err := smoothoperatorutils.SetImmutableLabels(reconcilerClient, ingressRoute, labels); err != nil {
		return err
	}

	// restful ingress should not be considered for uptime
	if ingressRoute.Annotations == nil {
		ingressRoute.Annotations = make(map[string]string)
	}

	ingressRoute.Annotations["uptime.pdok.nl/ignore"] = "-"

	mapproxyService := getTraefixService(obj, constants.MapserverPortNr)

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
func getUptimeName(obj *pdoknlv2.WMTS) string {
	url := obj.URL()
	path := url.Path
	split := strings.Split(path, "/")
	owner := "owner"
	dataset := "dataset"
	version := "v1_0"
	if len(split) > 1 {
		owner = split[1]
		owner = strings.Replace(owner, "-", " ", 99)
		owner = strings.ToUpper(owner)
	}

	if len(split) > 2 {
		dataset = split[2]
		dataset = strings.Replace(dataset, "-", " ", 99)
	}

	if len(split) > 4 {
		version = split[4]
	}

	return fmt.Sprintf("%s %s %s WMTS", owner, dataset, version)
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
	path := url.Path
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	if strings.Contains(host, "localhost") {
		return "Host(`localhost`) && PathPrefix(`" + path + "`)"
	}

	return "(Host(`localhost`) || Host(`" + host + "`)) && PathPrefix(`" + path + "`)"
}
