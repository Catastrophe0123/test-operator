/*
Copyright 2021.

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
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cachev1alpha1 "github.com/example/test-operator/api/v1alpha1"
	"github.com/example/test-operator/controllers"
	"github.com/gorilla/mux"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(cachev1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func createDeployment(mgr manager.Manager) {
	// deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: " -depl"} ,Spec: appsv1.DeploymentSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": }} }}
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "worker-depl",
			Namespace: "test-operator-system",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "worker-depl"},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "worker-depl"},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:  "worker-depl",
						Image: "heroku/nodejs-hello-world",
						Ports: []v1.ContainerPort{{
							Name:          "nodejs-port",
							ContainerPort: 5000,
						}},
					}},
				},
			},
		},
	}

	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "worker-srv",
			Namespace: "test-operator-system",
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": "worker-depl",
			},
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{{
				Port:       5000,
				NodePort:   31534,
				TargetPort: intstr.FromInt(5000),
			}},
		},
	}

	setupLog.Info("deployment " + deployment.GetName())
	setupLog.Info("service " + service.GetName())

	// fmt.Print("aoijda", deployment)
	// fmt.Print("aoijda", service)

	err := mgr.GetClient().Create(context.TODO(), deployment)
	if err != nil {
		setupLog.Error(err, "Failed to create new Deployment")
	}

	setupLog.Info("created deployment " + deployment.GetName())

	err = mgr.GetClient().Create(context.TODO(), service)
	if err != nil {
		setupLog.Error(err, "Failed to create new service")
	}

	setupLog.Info("created service " + service.GetName())

}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "e4fccaf4.example.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.CloudBaseMainReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CloudBaseMain")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	go func() {
		setupLog.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}
	}()

	// create http server
	setupLog.Info("Creating HTTP server")

	handler := func(w http.ResponseWriter, r *http.Request) {

		createDeployment(mgr)

		w.Write([]byte("Deployed the heroku hello world deployment !\n"))
	}

	router := mux.NewRouter()

	router.HandleFunc("/", handler)
	// http.Handle("/", router)
	setupLog.Info("Before listen and serve")

	log.Fatal(http.ListenAndServe(":8000", router))

}
