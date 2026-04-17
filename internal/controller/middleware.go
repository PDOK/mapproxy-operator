package controller

import (
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	smoothoperatorutils "github.com/pdok/smooth-operator/pkg/util"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefikiov1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const corsHeadersName = "mapproxy-headers"

func getBareCorsHeadersMiddleware(obj *pdoknlv2.WMTS) *traefikiov1alpha1.Middleware {
	return &traefikiov1alpha1.Middleware{
		ObjectMeta: metav1.ObjectMeta{
			Name: getSuffixedName(obj, corsHeadersName),
			// name might become too long. not handling here. will just fail on apply.
			Namespace: obj.GetNamespace(),
			UID:       obj.GetUID(),
		},
	}
}

func mutateCorsHeadersMiddleware(r *WMTSReconciler, obj *pdoknlv2.WMTS, middleware *traefikiov1alpha1.Middleware) error {
	reconcilerClient := r.Client

	labels := smoothoperatorutils.CloneOrEmptyMap(obj.GetLabels())
	if err := smoothoperatorutils.SetImmutableLabels(reconcilerClient, middleware, labels); err != nil {
		return err
	}
	middleware.Spec = traefikiov1alpha1.MiddlewareSpec{
		Headers: &dynamic.Headers{
			CustomResponseHeaders: map[string]string{
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Method":  "GET, POST, OPTIONS",
				"Access-Control-Allow-Origin":  "*",
				"Cache-Control":                "public, max-age=3600, no-transform",
			},
		},
	}

	if err := smoothoperatorutils.EnsureSetGVK(reconcilerClient, middleware, middleware); err != nil {
		return err
	}

	return ctrl.SetControllerReference(obj, middleware, r.Scheme)
}
