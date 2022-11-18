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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
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
		{"EmailDeployment", "", r.newEmailDeployment},
		{"EmailService", "", r.newEmailService},
		{"CheckoutService", "", r.newCheckoutService},
		{"RecommendationService", "", r.newRecommendationService},
	}
}

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
		obj.SetOwnerReferences(nil)
		u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}
		delete(u, "status")
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

func (r *BoutiqueShopReconciler) newEmailDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	container := corev1.Container{
		Name:  "server",
		Image: "gcr.io/google-samples/microservices-demo/emailservice:v0.3.9",
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: int32(8080),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "PORT",
				Value: "8080",
			},
			{
				Name:  "DISABLE_TRACING",
				Value: "1",
			},
			{
				Name:  "DISABLE_PROFILER",
				Value: "1",
			},
		},
		LivenessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:8080"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:8080"},
				},
			},
		},
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: pointer.Bool(false),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
			ReadOnlyRootFilesystem: pointer.Bool(true),
			//RunAsNonRoot:           pointer.Bool(true),
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
		},
	}

	labels := map[string]string{
		"app": emailName(instance),
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      emailName(instance),
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Name:   instance.ObjectMeta.Name,
				},
			},
		},
	}

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := deployment.Spec.Template.Spec.Containers[0]
			c.Name = container.Name
			c.Image = container.Image
			c.Ports = container.Ports
			c.ImagePullPolicy = container.ImagePullPolicy
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}

		return nil
	}

	return deployment, mutateFn, nil
}

func (r *BoutiqueShopReconciler) newEmailService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	return r.newService(ctx, instance, emailName,
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       5000,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(8080),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newCheckoutService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	return r.newService(ctx, instance, checkoutName,
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       5050,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(5050),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newRecommendationService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	return r.newService(ctx, instance, recommendationName,
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       8080,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(8080),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newService(ctx context.Context, instance *demov1alpha1.BoutiqueShop, nameFunc func(*demov1alpha1.BoutiqueShop) string, ports []corev1.ServicePort) (client.Object, controllerutil.MutateFn, error) {
	labels := map[string]string{
		"app": nameFunc(instance),
	}

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameFunc(instance),
			Namespace: instance.Namespace,
		},
	}

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
			return err
		}

		service.Spec.Ports = ports
		service.Spec.Selector = labels

		return nil
	}

	return service, mutateFn, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BoutiqueShopReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1alpha1.BoutiqueShop{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
