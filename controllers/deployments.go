package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	demov1alpha1 "github.com/saas-patterns/boutique-shop-operator/api/v1alpha1"
)

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
