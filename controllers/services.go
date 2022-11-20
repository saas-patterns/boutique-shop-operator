package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	demov1alpha1 "github.com/saas-patterns/boutique-shop-operator/api/v1alpha1"
)

func (r *BoutiqueShopReconciler) newAdService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	return r.newService(ctx, instance, adName,
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       9555,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(9555),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newCartService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	return r.newService(ctx, instance, cartName,
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       7070,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(7070),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newCatalogService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	return r.newService(ctx, instance, catalogName,
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       3550,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(3550),
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

func (r *BoutiqueShopReconciler) newCurrencyService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	return r.newService(ctx, instance, currencyName,
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       7000,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(7000),
			},
		},
	)
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

func (r *BoutiqueShopReconciler) newFrontendService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	return r.newService(ctx, instance, frontendName,
		[]corev1.ServicePort{
			{
				Name:       "http",
				Port:       80,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(8080),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newPaymentService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	return r.newService(ctx, instance, paymentName,
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       50051,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(50051),
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

func (r *BoutiqueShopReconciler) newRedisService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	return r.newService(ctx, instance, redisName,
		[]corev1.ServicePort{
			{
				Name:       "tls-redis",
				Port:       6379,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(6379),
			},
		},
	)
}

func (r *BoutiqueShopReconciler) newShippingService(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (client.Object, controllerutil.MutateFn, error) {
	return r.newService(ctx, instance, shippingName,
		[]corev1.ServicePort{
			{
				Name:       "grpc",
				Port:       50051,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(50051),
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
