// main.go

package main

import (
	"flag"
	"os"

	appv1 "your-domain.com/app-operator/api/v1"
	"your-domain.com/app-operator/controllers"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	flag.Parse()

	// 设置 Manager
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "your-domain.com-app-operator-lock",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// 设置 Scheme，用于反序列化和序列化对象
	if err := appv1.AddToScheme(mgr.GetScheme()); err != nil {
		setupLog.Error(err, "unable to add Appv1 to scheme")
		os.Exit(1)
	}

	// 设置控制器
	if err := controllers.NewAppReconciler(mgr.GetClient(), mgr.GetScheme()).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "App")
		os.Exit(1)
	}

	// 启动 Manager
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
