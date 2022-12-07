package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	demov1alpha1 "github.com/saas-patterns/boutique-shop-operator/api/v1alpha1"
)

const (
	adServicePort             = 9555
	cartServicePort           = 7070
	catalogServicePort        = 3550
	checkoutServicePort       = 5050
	currencyServicePort       = 7000
	emailServicePort          = 5000
	frontendServicePort       = 80
	paymentServicePort        = 50051
	recommendationServicePort = 8080
	redisServicePort          = 6379
	shippingServicePort       = 50051
)

func (r *BoutiqueShopReconciler) newAdService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	return r.newService(ctx, instance, adName(),
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       adServicePort,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(9555),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newCartService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	return r.newMethod(ctx, instance)
}

func (r *BoutiqueShopReconciler) newMethod(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	return r.newService(ctx, instance, cartName(),
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       cartServicePort,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(7070),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newCatalogService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	return r.newService(ctx, instance, catalogName(),
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       catalogServicePort,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(3550),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newCheckoutService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	return r.newService(ctx, instance, checkoutName(),
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       checkoutServicePort,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(5050),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newCurrencyService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	return r.newService(ctx, instance, currencyName(),
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       currencyServicePort,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(7000),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newEmailService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	return r.newService(ctx, instance, emailName(),
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       emailServicePort,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(8080),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newFrontendService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	return r.newService(ctx, instance, frontendName(),
		[]corev1.ServicePort{
			{
				Name:       "http",
				Port:       frontendServicePort,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(8080),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newPaymentService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	return r.newService(ctx, instance, paymentName(),
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       paymentServicePort,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(50051),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newRecommendationService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	return r.newService(ctx, instance, recommendationName(),
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       recommendationServicePort,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(8080),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newRedisService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	return r.newService(ctx, instance, redisName(),
		[]corev1.ServicePort{
			{
				Name:       "tls-redis",
				Port:       redisServicePort,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(6379),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newShippingService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	return r.newService(ctx, instance, shippingName(),
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       shippingServicePort,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(50051),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newFrontendServiceNodePort(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-frontend-external",
			Namespace: instance.Namespace,
		},
	}

	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
			return err
		}

		if len(service.Spec.Ports) != 1 {
			service.Spec.Ports = []corev1.ServicePort{corev1.ServicePort{}}
		}
		// set individual fields so we don't clobber the NodePort field that the
		// owning controller sets.
		p := &service.Spec.Ports[0]
		p.Name = "http"
		p.Port = 80
		p.Protocol = corev1.ProtocolTCP
		p.TargetPort = intstr.FromInt(8080)

		service.Spec.Selector = map[string]string{
			"app": frontendName(),
		}
		service.Spec.Type = corev1.ServiceTypeNodePort

		return nil
	}

	return &appResource{
		object:      service,
		mutateFn:    mutateFn,
		shouldExist: r.ExternalAccess == ExternalAccessServiceNodePort,
	}, nil
}

func (r *BoutiqueShopReconciler) newService(ctx context.Context, instance *demov1alpha1.BoutiqueShop, name string, ports []corev1.ServicePort) (*appResource, error) {
	labels := map[string]string{
		"app": name,
	}

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
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

	return &appResource{
		object:      service,
		mutateFn:    mutateFn,
		shouldExist: r.shouldExist(instance, name),
	}, nil
}
