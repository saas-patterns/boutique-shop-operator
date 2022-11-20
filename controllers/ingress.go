package controllers

import (
	"context"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	demov1alpha1 "github.com/saas-patterns/boutique-shop-operator/api/v1alpha1"
)

func (r *BoutiqueShopReconciler) newFrontendIngress(ctx context.Context, instance *demov1alpha1.BoutiqueShop) (*appResource, error) {
	ingress := networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{APIVersion: "networking/v1", Kind: "Ingress"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-ingress",
			Namespace: instance.Namespace,
		},
	}

	mutateFn := func() error {
		ingress.Spec.Rules = []networkingv1.IngressRule{
			{
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: (*networkingv1.PathType)(pointer.StringPtr(string(networkingv1.PathTypePrefix))),
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: frontendName(instance),
										Port: networkingv1.ServiceBackendPort{
											Number: frontendServicePort,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		return nil
	}

	return &appResource{
		&ingress,
		mutateFn,
		r.ExternalAccess == ExternalAccessIngress,
	}, nil
}
