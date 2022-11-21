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

	routev1 "github.com/openshift/api/route/v1"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	demov1alpha1 "github.com/saas-patterns/boutique-shop-operator/api/v1alpha1"
)

type NewComponentFn func(context.Context, *demov1alpha1.BoutiqueShop) (*appResource, error)

type appResource struct {
	object      client.Object
	mutateFn    controllerutil.MutateFn
	shouldExist bool
}

type component struct {
	name   string
	reason string
	fn     NewComponentFn
}

type ExternalAccess string

var (
	ExternalAccessIngress         ExternalAccess = "Ingress"
	ExternalAccessNone            ExternalAccess = "None"
	ExternalAccessRoute           ExternalAccess = "Route"
	ExternalAccessServiceNodePort ExternalAccess = "ServiceNodePort"
)

// BoutiqueShopReconciler reconciles a BoutiqueShop object
type BoutiqueShopReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	ExternalAccess ExternalAccess
	RouteAvailable bool
}

//+kubebuilder:rbac:groups=demo.openshift.com,resources=boutiqueshops,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demo.openshift.com,resources=boutiqueshops/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demo.openshift.com,resources=boutiqueshops/finalizers,verbs=update

func (r *BoutiqueShopReconciler) components() []component {
	return []component{
		{"AdDeployment", "", r.newAdDeployment},
		{"AdService", "", r.newAdService},
		{"CartDeployment", "", r.newCartDeployment},
		{"CartService", "", r.newCartService},
		{"CatalogDeployment", "", r.newCatalogDeployment},
		{"CatalogService", "", r.newCatalogService},
		{"CheckoutDeployment", "", r.newCheckoutDeployment},
		{"CheckoutService", "", r.newCheckoutService},
		{"CurrencyDeployment", "", r.newCurrencyDeployment},
		{"CurrencyService", "", r.newCurrencyService},
		{"EmailDeployment", "", r.newEmailDeployment},
		{"EmailService", "", r.newEmailService},
		{"FrontendDeployment", "", r.newFrontendDeployment},
		{"FrontendIngress", "", r.newFrontendIngress},
		{"FrontendRoute", "", r.newFrontendRoute},
		{"FrontendService", "", r.newFrontendService},
		{"FrontendServiceNodePort", "", r.newFrontendServiceNodePort},
		{"LoadGeneratorDeployment", "", r.newLoadGeneratorDeployment},
		{"PaymentDeployment", "", r.newPaymentDeployment},
		{"PaymentService", "", r.newPaymentService},
		{"RecommendationDeployment", "", r.newRecommendationDeployment},
		{"RecommendationService", "", r.newRecommendationService},
		{"RedisDeployment", "", r.newRedisDeployment},
		{"RedisService", "", r.newRedisService},
		{"ShippingDeployment", "", r.newShippingDeployment},
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
	components := r.components()
	for i, component := range components {
		resource, err := component.fn(ctx, instance)
		if err != nil {
			log.Error(err, "Failed to mutate resource", "Kind", component.name)
			return err
		}

		if !resource.shouldExist {
			continue
		}

		resource.mutateFn()
		// we don't want owner refs since the BoutiqueShop resource won't
		// actually exist
		resource.object.SetOwnerReferences(nil)
		// convert to Unstructured to do further changes
		u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource.object)
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

		if i < len(components)-1 {
			buf.Write([]byte("\n---\n"))
		}
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
		resource, err := component.fn(ctx, instance)
		if err != nil {
			log.Error(err, "Failed to mutate resource", "Kind", component.name)
			return ctrl.Result{}, err
		}

		if resource.shouldExist {
			result, err := controllerutil.CreateOrUpdate(ctx, r.Client, resource.object, resource.mutateFn)
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
			return ctrl.Result{}, nil
		}

		// if the resource is a Route and the API doesn't exist in the
		// apiserver, there's nothing else to do.
		if _, ok := resource.object.(*routev1.Route); ok {
			if !r.RouteAvailable {
				return ctrl.Result{}, nil
			}
		}

		// Ensure the resource does not exist, and call Delete if necessary
		key := client.ObjectKeyFromObject(resource.object)
		if err := r.Client.Get(ctx, key, resource.object); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			log.Error(err, "Get request for resource failed", "Kind", component.name)
			return ctrl.Result{}, err
		}
		err = r.Client.Delete(ctx, resource.object)
		if err != nil {
			log.Error(err, "Delete request for resource failed", "Kind", component.name)
			return ctrl.Result{}, err
		}
		log.Info("Deleted " + component.name)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BoutiqueShopReconciler) SetupWithManager(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).
		For(&demov1alpha1.BoutiqueShop{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{})

	switch r.ExternalAccess {
	case ExternalAccessIngress:
		b = b.Owns(&networkingv1.Ingress{})
	case ExternalAccessRoute:
		b = b.Owns(&routev1.Route{})
	}

	return b.Complete(r)
}
