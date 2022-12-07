package controllers

import (
	"context"
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	demov1alpha1 "github.com/saas-patterns/boutique-shop-operator/api/v1alpha1"
)

func (r *BoutiqueShopReconciler) newAdDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
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
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			InitialDelaySeconds: 20,
			PeriodSeconds:       15,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:9555"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			InitialDelaySeconds: 20,
			PeriodSeconds:       15,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:9555"},
				},
			},
		},
		// SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": adName(),
	}
	deployment := newDeployment(adName(), instance.Namespace, labels)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := &deployment.Spec.Template.Spec.Containers[0]
			c.Name = container.Name
			c.Image = container.Image
			c.Ports = container.Ports
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}
		//deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = pointer.Int64(5)

		return nil
	}

	return &appResource{
		deployment,
		mutateFn,
		r.shouldExist(instance, adName()),
	}, nil
}

func (r *BoutiqueShopReconciler) newCartDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
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
				Value: fmt.Sprintf("%s.%s.svc:%d", redisName(), r.getNs(instance, redisName()), redisServicePort),
			},
		},
		LivenessProbe: &corev1.Probe{
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:7070", "-rpc-timeout=5s"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:7070", "-rpc-timeout=5s"},
				},
			},
		},
		SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": cartName(),
	}

	deployment := newDeployment(cartName(), instance.Namespace, labels)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := &deployment.Spec.Template.Spec.Containers[0]
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

	return &appResource{
		deployment,
		mutateFn,
		r.shouldExist(instance, cartName()),
	}, nil
}

func (r *BoutiqueShopReconciler) newCatalogDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
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
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:3550"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:3550"},
				},
			},
		},
		SecurityContext: &corev1.SecurityContext{
			// AllowPrivilegeEscalation: pointer.Bool(false),
			// Capabilities: &corev1.Capabilities{
			// 	Drop: []corev1.Capability{"ALL"},
			// },
			// ReadOnlyRootFilesystem: pointer.Bool(true),
			// //RunAsNonRoot:           pointer.Bool(true),
			// SeccompProfile: &corev1.SeccompProfile{
			// 	Type: corev1.SeccompProfileTypeRuntimeDefault,
			// },
		},
	}

	labels := map[string]string{
		"app": catalogName(),
	}

	deployment := newDeployment(catalogName(), instance.Namespace, labels)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := &deployment.Spec.Template.Spec.Containers[0]
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
	return &appResource{
		deployment,
		mutateFn,
		r.shouldExist(instance, catalogName()),
	}, nil
}

func (r *BoutiqueShopReconciler) newCheckoutDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
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
				Value: fmt.Sprintf("%s.%s.svc:%d", cartName(), r.getNs(instance, cartName()), cartServicePort),
			},
			{
				Name:  "CURRENCY_SERVICE_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", currencyName(), r.getNs(instance, currencyName()), currencyServicePort),
			},
			{
				Name:  "EMAIL_SERVICE_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", emailName(), r.getNs(instance, emailName()), emailServicePort),
			},
			{
				Name:  "PAYMENT_SERVICE_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", paymentName(), r.getNs(instance, paymentName()), paymentServicePort),
			},
			{
				Name:  "PRODUCT_CATALOG_SERVICE_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", catalogName(), r.getNs(instance, catalogName()), catalogServicePort),
			},
			{
				Name:  "SHIPPING_SERVICE_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", shippingName(), r.getNs(instance, shippingName()), shippingServicePort),
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
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:5050"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:5050"},
				},
			},
		},
		SecurityContext: &corev1.SecurityContext{
			// AllowPrivilegeEscalation: pointer.Bool(false),
			// Capabilities: &corev1.Capabilities{
			// 	Drop: []corev1.Capability{"ALL"},
			// },
			// ReadOnlyRootFilesystem: pointer.Bool(true),
			// //RunAsNonRoot:           pointer.Bool(true),
			// SeccompProfile: &corev1.SeccompProfile{
			// 	Type: corev1.SeccompProfileTypeRuntimeDefault,
			// },
		},
	}

	labels := map[string]string{
		"app": checkoutName(),
	}

	deployment := newDeployment(checkoutName(), instance.Namespace, labels)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := &deployment.Spec.Template.Spec.Containers[0]
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

	return &appResource{
		deployment,
		mutateFn,
		r.shouldExist(instance, checkoutName()),
	}, nil
}

func (r *BoutiqueShopReconciler) newCurrencyDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
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
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:7000"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:7000"},
				},
			},
		},
		SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": currencyName(),
	}

	deployment := newDeployment(currencyName(), instance.Namespace, labels)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := &deployment.Spec.Template.Spec.Containers[0]
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

	return &appResource{
		deployment,
		mutateFn,
		r.shouldExist(instance, currencyName()),
	}, nil
}

func (r *BoutiqueShopReconciler) newEmailDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
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
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:8080"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:8080"},
				},
			},
		},
		SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": emailName(),
	}

	deployment := newDeployment(emailName(), instance.Namespace, labels)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := &deployment.Spec.Template.Spec.Containers[0]
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

	return &appResource{
		deployment,
		mutateFn,
		r.shouldExist(instance, emailName()),
	}, nil
}

func (r *BoutiqueShopReconciler) newFrontendDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
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
				Value: fmt.Sprintf("%s.%s.svc:%d", adName(), r.getNs(instance, adName()), adServicePort),
			},
			{
				Name:  "CART_SERVICE_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", cartName(), r.getNs(instance, cartName()), cartServicePort),
			},
			{
				Name:  "CHECKOUT_SERVICE_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", checkoutName(), r.getNs(instance, checkoutName()), checkoutServicePort),
			},
			{
				Name:  "CURRENCY_SERVICE_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", currencyName(), r.getNs(instance, currencyName()), currencyServicePort),
			},
			{
				Name:  "PRODUCT_CATALOG_SERVICE_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", catalogName(), r.getNs(instance, catalogName()), catalogServicePort),
			},
			{
				Name:  "RECOMMENDATION_SERVICE_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", recommendationName(), r.getNs(instance, recommendationName()), recommendationServicePort),
			},
			{
				Name:  "SHIPPING_SERVICE_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", shippingName(), r.getNs(instance, shippingName()), shippingServicePort),
			},
		},
		LivenessProbe: &corev1.Probe{
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/_healthz",
					Port:   intstr.FromInt(8080),
					Scheme: corev1.URISchemeHTTP,
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
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path:   "/_healthz",
					Port:   intstr.FromInt(8080),
					Scheme: corev1.URISchemeHTTP,
					HTTPHeaders: []corev1.HTTPHeader{
						{
							Name:  "Cookie",
							Value: "shop_session-id=x-readiness-probe",
						},
					},
				},
			},
		},
		SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": frontendName(),
	}

	deployment := newDeployment(frontendName(), instance.Namespace, labels)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := &deployment.Spec.Template.Spec.Containers[0]
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

	return &appResource{
		deployment,
		mutateFn,
		r.shouldExist(instance, frontendName()),
	}, nil
}

func (r *BoutiqueShopReconciler) newLoadGeneratorDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	labels := map[string]string{
		"app": loadGeneratorName(),
	}

	// don't run the load generator service if this is not set
	if instance.Spec.LoadGeneratorUsers == nil {
		return &appResource{
			object:      newDeployment(loadGeneratorName(), instance.Namespace, labels),
			mutateFn:    func() error { return nil },
			shouldExist: false,
		}, nil
	}

	container := corev1.Container{
		Name:  "main",
		Image: "gcr.io/google-samples/microservices-demo/loadgenerator:v0.3.9",
		Env: []corev1.EnvVar{
			{
				Name:  "FRONTEND_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", frontendName(), instance.Namespace, frontendServicePort),
			},
			{
				Name:  "USERS",
				Value: strconv.Itoa(*instance.Spec.LoadGeneratorUsers),
			},
		},
		SecurityContext: newContainerSecurityContext(),
	}

	initContainer := corev1.Container{
		Name:  "frontend-check",
		Image: "docker.io/busybox:latest",
		Env: []corev1.EnvVar{
			{
				Name:  "FRONTEND_ADDR",
				Value: fmt.Sprintf("%s.%s.svc:%d", frontendName(), r.getNs(instance, frontendName()), frontendServicePort),
			},
		},
		Command: []string{
			"/bin/sh",
			"-exc",
			"echo \"Init container pinging frontend: ${FRONTEND_ADDR}...\"\n" +
				"STATUSCODE=$(wget --server-response http://${FRONTEND_ADDR} 2>&1 | awk '/^  HTTP/{print $2}')\n" +
				"if test $STATUSCODE -ne 200; then\n" +
				"	echo \"Error: Could not reach frontend - Status code: ${STATUSCODE}\"\n" +
				"	exit 1\n" +
				"fi\n",
		},
		SecurityContext: newContainerSecurityContext(),
	}

	deployment := newDeployment(loadGeneratorName(), instance.Namespace, labels)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := &deployment.Spec.Template.Spec.Containers[0]
			c.Name = container.Name
			c.Image = container.Image
			c.Ports = container.Ports
			c.Env = container.Env
			c.LivenessProbe = container.LivenessProbe
			c.ReadinessProbe = container.ReadinessProbe
			c.SecurityContext = container.SecurityContext
		}

		if len(deployment.Spec.Template.Spec.InitContainers) != 1 {
			deployment.Spec.Template.Spec.InitContainers = []corev1.Container{initContainer}
		} else {
			c := &deployment.Spec.Template.Spec.InitContainers[0]
			c.Command = initContainer.Command
			c.Env = initContainer.Env
			c.Image = initContainer.Image
			c.Name = initContainer.Name
			c.SecurityContext = initContainer.SecurityContext
		}

		if deployment.ObjectMeta.Annotations == nil {
			deployment.ObjectMeta.Annotations = map[string]string{}
		}
		deployment.ObjectMeta.Annotations["sidecar.istio.io/rewriteAppHTTPProbers"] = "true"
		deployment.Spec.Template.Spec.SecurityContext = newPodSecurityContext()
		deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = pointer.Int64(5)

		return nil
	}

	return &appResource{
		deployment,
		mutateFn,
		r.shouldExist(instance, frontendName()),
	}, nil
}

func (r *BoutiqueShopReconciler) newPaymentDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
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
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:50051"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:50051"},
				},
			},
		},
		SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": paymentName(),
	}

	deployment := newDeployment(paymentName(), instance.Namespace, labels)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := &deployment.Spec.Template.Spec.Containers[0]
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

	return &appResource{
		deployment,
		mutateFn,
		r.shouldExist(instance, paymentName()),
	}, nil
}

func (r *BoutiqueShopReconciler) newRecommendationDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
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
				Value: fmt.Sprintf("%s.%s.svc:%d", catalogName(), r.getNs(instance, catalogName()), catalogServicePort),
			},
		},
		LivenessProbe: &corev1.Probe{
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:8080"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:8080"},
				},
			},
		},
		SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": recommendationName(),
	}

	deployment := newDeployment(recommendationName(), instance.Namespace, labels)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := &deployment.Spec.Template.Spec.Containers[0]
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

	return &appResource{
		deployment,
		mutateFn,
		r.shouldExist(instance, recommendationName()),
	}, nil
}

func (r *BoutiqueShopReconciler) newRedisDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
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
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(6379),
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
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
		"app": redisName(),
	}

	deployment := newDeployment(redisName(), instance.Namespace, labels)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := &deployment.Spec.Template.Spec.Containers[0]
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

	return &appResource{
		deployment,
		mutateFn,
		r.shouldExist(instance, redisName()),
	}, nil
}

func (r *BoutiqueShopReconciler) newShippingDeployment(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
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
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:50051"},
				},
			},
		},
		ReadinessProbe: &corev1.Probe{
			FailureThreshold:    3,
			InitialDelaySeconds: 15,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			TimeoutSeconds:      1,
			ProbeHandler: corev1.ProbeHandler{
				Exec: &corev1.ExecAction{
					Command: []string{"/bin/grpc_health_probe", "-addr=:50051"},
				},
			},
		},
		SecurityContext: newContainerSecurityContext(),
	}

	labels := map[string]string{
		"app": shippingName(),
	}

	deployment := newDeployment(shippingName(), instance.Namespace, labels)

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, deployment, r.Scheme); err != nil {
			return err
		}

		// don't clobber fields that were defaulted
		if len(deployment.Spec.Template.Spec.Containers) != 1 {
			deployment.Spec.Template.Spec.Containers = []corev1.Container{container}
		} else {
			c := &deployment.Spec.Template.Spec.Containers[0]
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

	return &appResource{
		deployment,
		mutateFn,
		r.shouldExist(instance, shippingName()),
	}, nil
}

func newContainerSecurityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		// AllowPrivilegeEscalation: pointer.Bool(false),
		// Capabilities: &corev1.Capabilities{
		// 	Drop: []corev1.Capability{"ALL"},
		// },
		// ReadOnlyRootFilesystem: pointer.Bool(true),
		// SeccompProfile: &corev1.SeccompProfile{
		// 	Type: corev1.SeccompProfileTypeRuntimeDefault,
		// },
	}
}

func newPodSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		// FSGroup:      pointer.Int64(1000),
		// RunAsGroup:   pointer.Int64(1000),
		// RunAsUser:    pointer.Int64(1000),
		// RunAsNonRoot: pointer.Bool(true),
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
