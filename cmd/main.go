/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"crypto/tls"
	"flag"
	"os"

	"github.com/go-logr/zapr"
	pdoknlv2 "github.com/pdok/mapproxy-operator/api/v2"
	"github.com/pdok/mapproxy-operator/internal/controller"
	"github.com/pdok/mapproxy-operator/internal/controller/types"
	v2 "github.com/pdok/mapproxy-operator/internal/webhook/v2"
	smoothoperatorv1 "github.com/pdok/smooth-operator/api/v1"
	"github.com/pdok/smooth-operator/pkg/integrations/logging"
	"github.com/peterbourgon/ff"
	"github.com/pkg/errors"
	traefikiov1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"go.uber.org/zap/zapcore"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	// +kubebuilder:scaffold:imports
)

const (
	EnvFalse = "false"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(traefikiov1alpha1.AddToScheme(scheme))
	utilruntime.Must(smoothoperatorv1.AddToScheme(scheme))
	utilruntime.Must(pdoknlv2.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// nolint:gocyclo,funlen
func main() {
	var metricsAddr string
	var metricsCertPath, metricsCertName, metricsCertKey string
	var webhookCertPath, webhookCertName, webhookCertKey string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var tlsOpts []func(*tls.Config)

	var multitoolImage, capabilitiesGeneratorImage, kvpToRestfulImage, mapproxyImage, apacheExporterImage string
	var slackWebhookURL string
	var logLevel int
	var setUptimeOperatorAnnotations bool
	var storageClassName string

	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.StringVar(&webhookCertPath, "webhook-cert-path", "", "The directory that contains the webhook certificate.")
	flag.StringVar(&webhookCertName, "webhook-cert-name", "tls.crt", "The name of the webhook certificate file.")
	flag.StringVar(&webhookCertKey, "webhook-cert-key", "tls.key", "The name of the webhook key file.")
	flag.StringVar(&metricsCertPath, "metrics-cert-path", "",
		"The directory that contains the metrics server certificate.")
	flag.StringVar(&metricsCertName, "metrics-cert-name", "tls.crt", "The name of the metrics server certificate file.")
	flag.StringVar(&metricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	flag.StringVar(&apacheExporterImage, "apache-exporter-image", "v0.7.0", "The image to use in the apache-exporter container.")
	flag.StringVar(&capabilitiesGeneratorImage, "capabilities-generator-image", "1.0.1", "The image to use in the capabilities generator init-container.")
	flag.StringVar(&mapproxyImage, "mapproxy-image", "3.1.3", "The image to use in the mapproxy container.")
	flag.StringVar(&kvpToRestfulImage, "kvp-to-restful-image", "1.2.8", "The image to use in the kvp-to-restful init-container.")
	flag.StringVar(&multitoolImage, "multitool-image", "0.9.1", "The image to use in the blob download init-container.")
	flag.BoolVar(&setUptimeOperatorAnnotations, "set-uptime-operator-annotations", true, "When enabled IngressRoutes get annotations that are used by the pdok/uptime-operator.")
	flag.StringVar(&storageClassName, "storage-class-name", "", "The name of the storage class to use when using an ephemeral volume.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)

	if err := ff.Parse(flag.CommandLine, os.Args[1:], ff.WithEnvVarNoPrefix()); err != nil {
		setupLog.Error(err, "unable to parse flags")
		os.Exit(1)
	}

	//nolint:gosec
	levelEnabler := zapcore.Level(logLevel)
	zapLogger, _ := logging.SetupLogger("mapproxy-operator", slackWebhookURL, levelEnabler)
	logrLogger := zapr.NewLogger(zapLogger)
	ctrl.SetLogger(logrLogger)

	reqFlags := make(map[string]string)
	reqFlags["apache-exporter-image"] = apacheExporterImage
	reqFlags["capabilities-generator-image"] = capabilitiesGeneratorImage
	reqFlags["mapproxy-image"] = mapproxyImage
	reqFlags["kvp-to-restful-image"] = kvpToRestfulImage
	reqFlags["multitool-image"] = multitoolImage

	for reqFlag, val := range reqFlags {
		if val == "" {
			setupLog.Error(errors.New(reqFlag+" is a required flag"), "A value for "+reqFlag+" must be specified.")
			os.Exit(1)
		}
	}

	controller.SetUptimeOperatorAnnotations(setUptimeOperatorAnnotations)
	controller.SetStorageClassName(storageClassName)

	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("Disabling HTTP/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookTLSOpts := tlsOpts
	webhookServerOptions := webhook.Options{
		TLSOpts: webhookTLSOpts,
	}

	if len(webhookCertPath) > 0 {
		setupLog.Info("Initializing webhook certificate watcher using provided certificates",
			"webhook-cert-path", webhookCertPath, "webhook-cert-name", webhookCertName, "webhook-cert-key", webhookCertKey)

		webhookServerOptions.CertDir = webhookCertPath
		webhookServerOptions.CertName = webhookCertName
		webhookServerOptions.KeyName = webhookCertKey
	}

	webhookServer := webhook.NewServer(webhookServerOptions)

	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics {
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	if len(metricsCertPath) > 0 {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", metricsCertPath, "metrics-cert-name", metricsCertName, "metrics-cert-key", metricsCertKey)

		metricsServerOptions.CertDir = metricsCertPath
		metricsServerOptions.CertName = metricsCertName
		metricsServerOptions.KeyName = metricsCertKey
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "8726aa76.pdok.nl",
	})
	if err != nil {
		setupLog.Error(err, "Failed to start manager")
		os.Exit(1)
	}

	if err = (&controller.WMTSReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Images: types.Images{
			ApacheExporterImage:        apacheExporterImage,
			CapabilitiesGeneratorImage: capabilitiesGeneratorImage,
			KvpToRestfulImage:          kvpToRestfulImage,
			MapproxyImage:              mapproxyImage,
			MultiToolImage:             multitoolImage,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "WMTS")
		os.Exit(1)
	}

	if os.Getenv("ENABLE_WEBHOOKS") != EnvFalse {
		if err = v2.SetupWMTSWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "WMTS")
			os.Exit(1)
		}
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Failed to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Failed to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Failed to run manager")
		os.Exit(1)
	}
}
