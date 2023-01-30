/*
Copyright 2022 Beez Innovation Labs.

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

package models

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beezlabs-org/cloudflare-tunnel-operator/controllers/constants"
)

type DeploymentModel struct {
	Name       string
	Namespace  string
	TunnelID   string
	Container  *corev1.Container
	Deployment *appsv1.DeploymentSpec
}

func (d *DeploymentModel) GetContainer() *corev1.Container {
	if d.Container.Image == "" {
		d.Container.Image = "cloudflare/cloudflared:latest"
	}

	if len(d.Container.Command) == 0 {
		d.Container.Command = []string{"cloudflared"}
	}

	if len(d.Container.Args) == 0 {
		d.Container.Args = []string{"tunnel", "--metrics", "localhost:9090", "--no-autoupdate", "--config", "/config/config.yaml", "run"}
	}

	d.Container.Name = "cloudfalred"

	d.Container.Ports = append(d.Container.Ports, corev1.ContainerPort{
		Name:          "metrics",
		ContainerPort: 9090,
		Protocol:      corev1.ProtocolTCP,
	})

	d.Container.VolumeMounts = append(
		d.Container.VolumeMounts,
		corev1.VolumeMount{
			Name:      "cloudflared-config",
			MountPath: "/config/config.yaml",
			SubPath:   "config.yaml",
		},
		corev1.VolumeMount{
			Name:      "cloudflared-creds",
			MountPath: "/config/" + d.TunnelID + ".json",
			SubPath:   d.TunnelID + ".json",
		},
	)

	return d.Container
}

func Deployment(model DeploymentModel) *DeploymentModel {
	return &model
}

func (d *DeploymentModel) GetDeployment() *appsv1.Deployment {
	deploy := d.Deployment
	deploy.Template.Spec.Containers = append(deploy.Template.Spec.Containers, *d.GetContainer())
	deploy.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/default-container"] = "cloudflared"
	deploy.Template.ObjectMeta.Labels["app.kubernetes.io/name"] = d.Name
	deploy.Selector.MatchLabels["app.kubernetes.io/name"] = d.Name
	deploy.Template.Spec.Volumes = append(deploy.Template.Spec.Volumes, corev1.Volume{
		Name: "cloudflared-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: d.Name + "-" + constants.ResourceSuffix},
			},
		},
	}, corev1.Volume{
		Name: "cloudflared-creds",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: d.Name + "-" + constants.ResourceSuffix,
			},
		},
	})

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.Name + "-" + constants.ResourceSuffix,
			Namespace: d.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       d.Name,
				"app.kubernetes.io/component":  "controller",
				"app.kubernetes.io/created-by": constants.OperatorName,
			},
		},
		Spec: *deploy,
	}
}
