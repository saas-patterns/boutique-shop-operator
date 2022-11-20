package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	demov1alpha1 "github.com/saas-patterns/boutique-shop-operator/api/v1alpha1"
)

func (r *BoutiqueShopReconciler) newAdDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	container := corev1.Container{
		Name:  "server",
		Image: "gcr.io/google-samples/microservices-demo/adservice:v0.3.9",
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: int32(9555),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "PORT",
				Value: "9555",
			},
			{
				Name:  "DISABLE_STATS",
				Value: "1",
			},
			{
				Name:  "DISABLE_TRACING",
				Value: "1",
			},
		},
		LivenessProbe: &corev1.Probe{
			InitialDelaySeconds: 20,
			PeriodSeconds:       15,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:9555"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			InitialDelaySeconds: 20,
			PeriodSeconds:       15,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:9555"},
				},
			},
		},
		SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": adName(instance),
	}

	deployment := newDeployment(adName(instance), instance.Namespace, labels)

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
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}
		deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = pointer.Int64(5)

		return nil
	}

	return deployment, mutateFn, nil
}

func (r *BoutiqueShopReconciler) newCartDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	container := corev1.Container{
		Name:  "server",
		Image: "gcr.io/google-samples/microservices-demo/cartservice:v0.3.9",
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: int32(7070),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "REDIS_ADDR",
				Value: fmt.Sprintf("%s:%d", redisName(instance), redisServicePort),
			},
		},
		LivenessProbe: &corev1.Probe{
			InitialDelaySeconds: 15,
			PeriodSeconds:       10,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:7070", "-rpc-timeout=5s"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			InitialDelaySeconds: 15,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:7070", "-rpc-timeout=5s"},
				},
			},
		},
		SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": cartName(instance),
	}

	deployment := newDeployment(cartName(instance), instance.Namespace, labels)

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
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}
		deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = pointer.Int64(5)

		return nil
	}

	return deployment, mutateFn, nil
}

func (r *BoutiqueShopReconciler) newCatalogDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	container := corev1.Container{
		Name:  "server",
		Image: "gcr.io/google-samples/microservices-demo/productcatalogservice:v0.3.9",
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: int32(3550),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "PORT",
				Value: "3550",
			},
			{
				Name:  "DISABLE_PROFILER",
				Value: "1",
			},
			{
				Name:  "DISABLE_STATS",
				Value: "1",
			},
			{
				Name:  "DISABLE_TRACING",
				Value: "1",
			},
		},
		LivenessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:3550"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:3550"},
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
		"app": catalogName(instance),
	}

	deployment := newDeployment(catalogName(instance), instance.Namespace, labels)

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
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}
		deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = pointer.Int64(5)

		return nil
	}

	return deployment, mutateFn, nil
}

func (r *BoutiqueShopReconciler) newCheckoutDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	container := corev1.Container{
		Name:  "server",
		Image: "gcr.io/google-samples/microservices-demo/checkoutservice:v0.3.9",
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: int32(5050),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "PORT",
				Value: "5050",
			},
			{
				Name:  "CART_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", cartName(instance), cartServicePort),
			},
			{
				Name:  "CURRENCY_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", currencyName(instance), currencyServicePort),
			},
			{
				Name:  "EMAIL_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", emailName(instance), emailServicePort),
			},
			{
				Name:  "PAYMENT_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", paymentName(instance), paymentServicePort),
			},
			{
				Name:  "PRODUCT_CATALOG_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", catalogName(instance), catalogServicePort),
			},
			{
				Name:  "SHIPPING_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", shippingName(instance), shippingServicePort),
			},
			{
				Name:  "DISABLE_PROFILER",
				Value: "1",
			},
			{
				Name:  "DISABLE_STATS",
				Value: "1",
			},
			{
				Name:  "DISABLE_TRACING",
				Value: "1",
			},
		},
		LivenessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:5050"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:5050"},
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
		"app": checkoutName(instance),
	}

	deployment := newDeployment(checkoutName(instance), instance.Namespace, labels)

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
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}
		deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()

		return nil
	}

	return deployment, mutateFn, nil
}

func (r *BoutiqueShopReconciler) newCurrencyDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	container := corev1.Container{
		Name:  "server",
		Image: "gcr.io/google-samples/microservices-demo/currencyservice:v0.3.9",
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: int32(7000),
				Name:          "grpc",
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "PORT",
				Value: "7000",
			},
			{
				Name:  "DISABLE_DEBUGGER",
				Value: "1",
			},
			{
				Name:  "DISABLE_PROFILER",
				Value: "1",
			},
			{
				Name:  "DISABLE_TRACING",
				Value: "1",
			},
		},
		LivenessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:7000"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:7000"},
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
		"app": currencyName(instance),
	}

	deployment := newDeployment(currencyName(instance), instance.Namespace, labels)

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
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}
		deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = pointer.Int64(5)

		return nil
	}

	return deployment, mutateFn, nil
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

	deployment := newDeployment(emailName(instance), instance.Namespace, labels)

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
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}
		deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = pointer.Int64(5)

		return nil
	}

	return deployment, mutateFn, nil
}

func (r *BoutiqueShopReconciler) newFrontendDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	container := corev1.Container{
		Name:  "server",
		Image: "gcr.io/google-samples/microservices-demo/frontend:v0.3.9",
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
			{
				Name:  "AD_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", adName(instance), adServicePort),
			},
			{
				Name:  "CART_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", cartName(instance), cartServicePort),
			},
			{
				Name:  "CHECKOUT_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", checkoutName(instance), checkoutServicePort),
			},
			{
				Name:  "CURRENCY_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", currencyName(instance), currencyServicePort),
			},
			{
				Name:  "PRODUCT_CATALOG_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", catalogName(instance), catalogServicePort),
			},
			{
				Name:  "RECOMMENDATION_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", recommendationName(instance), recommendationServicePort),
			},
			{
				Name:  "SHIPPING_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", shippingName(instance), shippingServicePort),
			},
		},
		LivenessProbe: &corev1.Probe{
			PeriodSeconds: 10,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/_healthz",
					Port: intstr.FromInt(8080),
					HTTPHeaders: []corev1.HTTPHeader{
						{
							Name:  "Cookie",
							Value: "shop_session-id=x-liveness-probe",
						},
					},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds: 10,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/_healthz",
					Port: intstr.FromInt(8080),
					HTTPHeaders: []corev1.HTTPHeader{
						{
							Name:  "Cookie",
							Value: "shop_session-id=x-readiness-probe",
						},
					},
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
		"app": frontendName(instance),
	}

	deployment := newDeployment(frontendName(instance), instance.Namespace, labels)

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
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}
		if deployment.ObjectMeta.Annotations == nil {
			deployment.ObjectMeta.Annotations = map[string]string{}
		}
		deployment.ObjectMeta.Annotations["sidecar.istio.io/rewriteAppHTTPProbers"] = "true"
		deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()

		return nil
	}

	return deployment, mutateFn, nil
}

func (r *BoutiqueShopReconciler) newLoadGeneratorDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	// TODO: implement this
	return nil, nil, nil
}

func (r *BoutiqueShopReconciler) newPaymentDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	container := corev1.Container{
		Name:  "server",
		Image: "gcr.io/google-samples/microservices-demo/paymentservice:v0.3.9",
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: int32(50051),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "PORT",
				Value: "50051",
			},
			{
				Name:  "DISABLE_DEBUGGER",
				Value: "1",
			},
			{
				Name:  "DISABLE_PROFILER",
				Value: "1",
			},
			{
				Name:  "DISABLE_TRACING",
				Value: "1",
			},
		},
		LivenessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:50051"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:50051"},
				},
			},
		},
		SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": paymentName(instance),
	}

	deployment := newDeployment(paymentName(instance), instance.Namespace, labels)

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
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}
		deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = pointer.Int64(5)

		return nil
	}

	return deployment, mutateFn, nil
}

func (r *BoutiqueShopReconciler) newRecommendationDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	container := corev1.Container{
		Name:  "server",
		Image: "gcr.io/google-samples/microservices-demo/recommendationservice:v0.3.9",
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
				Name:  "DISABLE_DEBUGGER",
				Value: "1",
			},
			{
				Name:  "DISABLE_PROFILER",
				Value: "1",
			},
			{
				Name:  "DISABLE_TRACING",
				Value: "1",
			},
			{
				Name:  "PRODUCT_CATALOG_SERVICE_ADDR",
				Value: fmt.Sprintf("%s:%d", catalogName(instance), catalogServicePort),
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
		SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": recommendationName(instance),
	}

	deployment := newDeployment(recommendationName(instance), instance.Namespace, labels)

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
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}
		deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = pointer.Int64(5)

		return nil
	}

	return deployment, mutateFn, nil
}

func (r *BoutiqueShopReconciler) newRedisDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	container := corev1.Container{
		Name:  "redis",
		Image: "docker.io/redis:alpine",
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: int32(6379),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		LivenessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(6379),
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(6379),
				},
			},
		},
		SecurityContext: newContainerSecurityContext(),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "redis-data",
				MountPath: "/data",
			},
		},
	}

	labels := map[string]string{
		"app": redisName(instance),
	}

	deployment := newDeployment(redisName(instance), instance.Namespace, labels)

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
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}
		deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()
		deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "redis-data",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		}

		return nil
	}

	return deployment, mutateFn, nil
}

func (r *BoutiqueShopReconciler) newShippingDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	container := corev1.Container{
		Name:  "server",
		Image: "gcr.io/google-samples/microservices-demo/shippingservice:v0.3.9",
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: int32(50051),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "PORT",
				Value: "50051",
			},
			{
				Name:  "DISABLE_STATS",
				Value: "1",
			},
			{
				Name:  "DISABLE_PROFILER",
				Value: "1",
			},
			{
				Name:  "DISABLE_TRACING",
				Value: "1",
			},
		},
		LivenessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:50051"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds: 5,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:50051"},
				},
			},
		},
		SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": shippingName(instance),
	}

	deployment := newDeployment(shippingName(instance), instance.Namespace, labels)

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
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}
		deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = pointer.Int64(5)

		return nil
	}

	return deployment, mutateFn, nil
}

func newContainerSecurityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		AllowPrivilegeEscalation: pointer.Bool(false),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
		ReadOnlyRootFilesystem: pointer.Bool(true),
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
}

func newPodSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		FSGroup:      pointer.Int64(1000),
		RunAsGroup:   pointer.Int64(1000),
		RunAsUser:    pointer.Int64(1000),
		RunAsNonRoot: pointer.Bool(true),
	}
}

func newDeployment(name, namespace string, labels map[string]string) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
			},
		},
	}
}
