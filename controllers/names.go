package controllers

import demov1alpha1 "github.com/saas-patterns/boutique-shop-operator/api/v1alpha1"

func emailName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-email"
}

func checkoutName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-checkout"
}

func recommendationName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-recommendation"
}
