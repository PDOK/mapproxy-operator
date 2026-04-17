package v2

import (
	"context"

	v2 "github.com/pdok/mapproxy-operator/api/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var wmtsLog = logf.Log.WithName("wmts-resource")

func SetupWMTSWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &v2.WMTS{}).WithValidator(&WMTSCustomValidator{mgr.GetClient()}).Complete()
}

type WMTSCustomValidator struct {
	Client client.Client
}

func (W *WMTSCustomValidator) ValidateCreate(ctx context.Context, obj *v2.WMTS) (warnings admission.Warnings, err error) {
	wmtsLog.Info("Validation for WMTS upon creation", "name", obj.Name)
	return obj.ValidateCreate(W.Client)
}

func (W *WMTSCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *v2.WMTS) (warnings admission.Warnings, err error) {
	wmtsLog.Info("Validation for WMTS upon update", "name", newObj.Name)
	return newObj.ValidateUpdate(W.Client, oldObj)
}

func (W *WMTSCustomValidator) ValidateDelete(ctx context.Context, obj *v2.WMTS) (warnings admission.Warnings, err error) {
	wmtsLog.Info("Validation for WMTS upon deletion", "name", obj.Name)
	// No validation as of now
	return nil, nil
}
