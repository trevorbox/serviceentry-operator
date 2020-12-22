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
	networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceEntryReconciler reconciles a ServiceEntry object
type ServiceEntryReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}
type void struct{}

const updateAnnotation = "updateme"

var (
	ipRegex, _ = regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	log        logr.Logger
	member     void
)

// +kubebuilder:rbac:groups=networking.istio.io,resources=serviceentries,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=networking.istio.io,resources=serviceentries/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get

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
	log = r.Log.WithValues("ServiceEntry", req.NamespacedName)

	log.Info("Reconcile...")

	se := &networking.ServiceEntry{}

	err := r.Get(ctx, req.NamespacedName, se)

	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("ServiceEntry resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get ServiceEntry")
		return ctrl.Result{}, err
	}

	// Not configured to track, don't requeue it
	if se.GetAnnotations()[updateAnnotation] != "true" {
		return ctrl.Result{}, nil
	}

	newIps := lookupIps(se.Spec.Hosts)

	if !arrEqual(se.Spec.Addresses, newIps) {
		se.Spec.Addresses = newIps
		err = r.Update(ctx, se)
		if err != nil {
			log.Error(err, "Failed to update ServiceEntry")
		} else {
			log.Info("Updated ServiceEntry", "se.Spec.Addresses", se.Spec.Addresses)
		}
	}

	return ctrl.Result{
		RequeueAfter: time.Second * 5,
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
		if !strings.HasPrefix(host, "*") {
			addr, err := net.LookupIP(host)
			if err != nil {
				log.Error(err, "Unknown Host", "host", host)
			} else {
				for _, ip := range addr {
					if isIpv4Net(ip.String()) {
						addresses = append(addresses, ip.String())
					}

				}

			}
		}
	}
	return addresses
}

// use a set to determine if arrays have the same values
func arrEqual(a, b []string) bool {

	if len(a) != len(b) {
		return false
	}
	set := make(map[string]void)

	for _, v := range a {
		set[v] = member
	}

	for _, v := range b {
		_, exists := set[v]
		if !exists {
			return false
		}
	}

	return true
}
