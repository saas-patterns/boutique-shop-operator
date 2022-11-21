package controllers

import demov1alpha1 "github.com/saas-patterns/boutique-shop-operator/api/v1alpha1"

func adName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-ad"
}

func cartName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-cart"
}

func catalogName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-catalog"
}

func checkoutName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-checkout"
}

func currencyName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-currency"
}

func emailName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-email"
}

func frontendName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-frontend"
}

func loadGeneratorName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-loadgenerator"
}

func paymentName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-payment"
}

func recommendationName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-recommendation"
}

func redisName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-redis"
}

func routeName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-route"
}

func shippingName(instance *demov1alpha1.BoutiqueShop) string {
	return instance.Name + "-shipping"
}
