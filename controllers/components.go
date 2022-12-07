package controllers

import (
	demov1alpha1 "github.com/saas-patterns/boutique-shop-operator/api/v1alpha1"
	"golang.org/x/exp/slices"
)

var freeTierComponentToNsDict = map[string]string{
	adName():             "boutique-free",
	cartName():           "boutique-free",
	catalogName():        "boutique-free",
	checkoutName():       "boutique-free",
	currencyName():       "boutique-free",
	emailName():          "boutique-free",
	paymentName():        "boutique-free",
	recommendationName(): "boutique-free",
	redisName():          "boutique-free",
	shippingName():       "boutique-free",
	frontendName():       "boutique-free",
	routeName():          "boutique-free",
}

var silverTierComponentToNsDict = map[string]string{
	adName():             "boutique-silver",
	cartName():           "boutique-silver",
	catalogName():        "boutique-silver",
	checkoutName():       "boutique-silver",
	currencyName():       "boutique-silver",
	emailName():          "boutique-silver",
	paymentName():        "boutique-silver",
	recommendationName(): "boutique-silver",
	redisName():          "boutique-silver",
	shippingName():       "boutique-silver",
	frontendName():       "boutique-silver",
	routeName():          "boutique-silver",
}

var goldTierComponentToNsDict = map[string]string{
	adName():             "tenant",
	cartName():           "tenant",
	catalogName():        "boutique-ops",
	checkoutName():       "tenant",
	currencyName():       "enterprise-utilities",
	emailName():          "enterprise-utilities",
	paymentName():        "enterprise-utilities",
	recommendationName(): "boutique-ops",
	redisName():          "tenant",
	shippingName():       "boutique-ops",
	frontendName():       "tenant",
	routeName():          "tenant",
}

var platinumTierComponentToNsDict = map[string]string{
	adName():             "tenant",
	cartName():           "tenant",
	catalogName():        "tenant",
	checkoutName():       "tenant",
	currencyName():       "tenant",
	emailName():          "tenant",
	paymentName():        "tenant",
	recommendationName(): "tenant",
	redisName():          "tenant",
	shippingName():       "tenant",
	frontendName():       "tenant",
	routeName():          "tenant",
}

var tierToComponentNsDict = map[string]map[string]string{
	"free":     freeTierComponentToNsDict,
	"silver":   silverTierComponentToNsDict,
	"gold":     goldTierComponentToNsDict,
	"platinum": platinumTierComponentToNsDict,
}

var reservedNs = []string{
	"boutique-free", "boutique-silver", "boutique-ops", "enterprise-utilities",
}

func (r *BoutiqueShopReconciler) getNs(instance *demov1alpha1.BoutiqueShop, component string) string {
	ns := tierToComponentNsDict[instance.Spec.Tier][component]
	if ns == "tenant" && !slices.Contains(reservedNs, instance.Namespace) {
		ns = instance.Namespace
	}
	return ns
}

func (r *BoutiqueShopReconciler) shouldExist(instance *demov1alpha1.BoutiqueShop, component string) bool {
	ns := r.getNs(instance, component)
	// if the ns is reserved ns and the instance name is not the reserved ns return false
	// free-tier component should be deployed by only free-tier insance
	if slices.Contains(reservedNs, ns) && !slices.Contains(reservedNs, instance.Name) {
		return false
	}
	// if its not reserved ns then component should be deployed if the namespace matches the instance ns
	return instance.Namespace == ns
}

// Route should not exist if the instance name is
func (r *BoutiqueShopReconciler) shouldRouteExist(instance *demov1alpha1.BoutiqueShop, component string) bool {
	// don't create routes if the instance name is in reservedNS
	if slices.Contains(reservedNs, instance.Name) {
		return false
	}
	// if its not reserved ns then component should be deployed if the namespace matches the instance ns
	return instance.Namespace == r.getNs(instance, component)
}
