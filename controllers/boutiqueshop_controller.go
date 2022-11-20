/*
Copyright 2022.

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
	"bytes"
	"context"
	"io"

	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	demov1alpha1 "github.com/saas-patterns/boutique-shop-operator/api/v1alpha1"
)

type NewComponentFn func(context.Context, *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error)

type component struct {
	name   string
	reason string
	fn     NewComponentFn
}

// BoutiqueShopReconciler reconciles a BoutiqueShop object
type BoutiqueShopReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=demo.openshift.com,resources=boutiqueshops,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demo.openshift.com,resources=boutiqueshops/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demo.openshift.com,resources=boutiqueshops/finalizers,verbs=update

func (r *BoutiqueShopReconciler) components() []component {
	return []component{
		{"AdService", "", r.newAdService},
		{"CartService", "", r.newCartService},
		{"CatalogService", "", r.newCatalogService},
		{"CheckoutService", "", r.newCheckoutService},
		{"CurrencyService", "", r.newCurrencyService},
		{"EmailDeployment", "", r.newEmailDeployment},
		{"EmailService", "", r.newEmailService},
		{"FrontendService", "", r.newFrontendService},
		{"PaymentService", "", r.newPaymentService},
		{"RecommendationService", "", r.newRecommendationService},
		{"RedisService", "", r.newRedisService},
		{"ShippingService", "", r.newShippingService},
	}
}

// WriteManifests generates a yaml manifest for the whole application
// corresponding to what's defined on the BoutiqueShop instance. The yaml is
// then written to the provided Writer.
func (r *BoutiqueShopReconciler) WriteManifests(instance *demov1alpha1.BoutiqueShop, out io.Writer) error {
	ctx := context.TODO()
	log := ctrllog.FromContext(ctx)

	buf := bytes.Buffer{}
	for i, component := range r.components() {
		if i > 0 {
			buf.Write([]byte("\n---\n"))
		}
		obj, mutateFn, err := component.fn(ctx, instance)
		if err != nil {
			log.Error(err, "Failed to mutate resource", "Kind", component.name)
			return err
		}
		mutateFn()
		// we don't want owner refs since the BoutiqueShop resource won't
		// actually exist
		obj.SetOwnerReferences(nil)
		// convert to Unstructured to do further changes
		u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}
		// we don't want status fields
		delete(u, "status")
		// creationTimestamp field for some reason renders with a nil value, but
		// we want to remove it
		delete(u["metadata"].(map[string]interface{}), "creationTimestamp")

		b, err := yaml.Marshal(u)
		if err != nil {
			return err
		}
		buf.Write(b)
	}
	_, err := buf.WriteTo(out)
	return err
}

func (r *BoutiqueShopReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	instance := &demov1alpha1.BoutiqueShop{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	for _, component := range r.components() {
		obj, mutateFn, err := component.fn(ctx, instance)
		if err != nil {
			log.Error(err, "Failed to mutate resource", "Kind", component.name)
			return ctrl.Result{}, err
		}

		result, err := controllerutil.CreateOrUpdate(ctx, r.Client, obj, mutateFn)
		if err != nil {
			log.Error(err, "Failed to create or update", "Kind", component.name)
			return ctrl.Result{}, err
		}
		switch result {
		case controllerutil.OperationResultCreated:
			log.Info("Created " + component.name)
		case controllerutil.OperationResultUpdated:
			log.Info("Updated " + component.name)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BoutiqueShopReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1alpha1.BoutiqueShop{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
