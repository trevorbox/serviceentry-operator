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
	logger := r.Log.WithValues("serviceentry", req.NamespacedName)

	// Using a unstructured object.
	//u := getInstance()

	se := &networking.ServiceEntry{}

	err := r.Get(ctx, req.NamespacedName, se)

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

	//uc := u.UnstructuredContent()

	//hosts := uc["spec"].(map[string]interface{})["hosts"]

	//u.SetUnstructuredContent()

	//logger.Info("UnstructuredContent", "hosts", hosts)
	//logger.Info("UnstructuredContent", "apiVersion", u.GetAPIVersion())
	//logger.Info("UnstructuredContent", "annotations", u.GetAnnotations())

	m := make(map[string]string)
	m["yes"] = "yes"
	//u.SetAnnotations(m)

	//err = r.Update(ctx, u)

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
