/*
Copyright 2020.

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

package controllers

import (
	"context"

	"net"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"

	errs "k8s.io/apimachinery/pkg/api/errors"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	//corev1 "k8s.io/api/core/v1"
	//networking "istio.io/api/networking/v1alpha3"
	networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

// ServiceEntryReconciler reconciles a ServiceEntry object
type ServiceEntryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

var (
	ipRegex, _ = regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	logger     logr.Logger
)

func getInstance() *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "networking.istio.io",
		Kind:    "ServiceEntry",
		Version: "v1alpha3",
	})
	return u
}

// +kubebuilder:rbac:groups=com.example.example.com,resources=serviceentries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=com.example.example.com,resources=serviceentries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=com.example.example.com,resources=serviceentries/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServiceEntry object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *ServiceEntryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger = r.Log.WithValues("serviceentry", req.NamespacedName)

	se := &networking.ServiceEntry{}

	err := r.Get(context.TODO(), req.NamespacedName, se)

	if err != nil {
		if errs.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	logger.Info("", "hosts", se.Spec.Hosts)

	ips := lookupIps(se.Spec.Hosts)

	logger.Info("", "ips", ips)

	return reconcile.Result{
		RequeueAfter: time.Second * 30,
	}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceEntryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		//For(getInstance()).
		For(&networking.ServiceEntry{}).
		Complete(r)
}

func isIpv4Net(ip string) bool {
	return ipRegex.MatchString(strings.Trim(ip, " "))
}

func lookupIps(hosts []string) []string {
	var addresses []string
	for _, host := range hosts {
		addr, err := net.LookupIP(host)
		if err != nil {
			logger.Error(err, "Unknown Host", "host", host)
		} else {
			for _, ip := range addr {
				if isIpv4Net(ip.String()) {
					addresses = append(addresses, ip.String()+"/32")
				}
			}

		}
	}
	return addresses
}
