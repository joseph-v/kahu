/*
Copyright 2022 The SODA Authors.

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

package backup

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	metaservice "github.com/soda-cdm/kahu/providerframework/metaservice/lib/go"
	"github.com/soda-cdm/kahu/utils"
)

func (ctrl *controller) GetPodAndBackup(name, namespace string,
	backupClient metaservice.MetaService_BackupClient) error {
	pod, err := ctrl.kubeClient.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return ctrl.backupSend(pod, pod.Name, backupClient)

}

func (ctrl *controller) podBackup(namespace string,
	backup *PrepareBackup, backupClient metaservice.MetaService_BackupClient) error {
	ctrl.logger.Infoln("Starting collecting pods")

	var podLabelList []map[string]string
	var labelSelectors map[string]string
	if backup.Spec.Label != nil {
		labelSelectors = backup.Spec.Label.MatchLabels
	}

	selectors := labels.Set(labelSelectors).String()
	podList, err := ctrl.kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selectors,
	})
	if err != nil {
		return err
	}
	var podAllList []string
	for _, pod := range podList.Items {
		podAllList = append(podAllList, pod.Name)
	}

	podAllList = utils.FindMatchedStrings(utils.Pod, podAllList, backup.Spec.IncludeResources,
		backup.Spec.ExcludeResources)

	// only for backup
	for _, pod := range podList.Items {
		if utils.Contains(podAllList, pod.Name) {
			// Run pre hooks for the pod
			// backup the deployment yaml
			err = ctrl.GetPodAndBackup(pod.Name, pod.Namespace, backupClient)
			if err != nil {
				return err
			}

			// backup the volumespec releted object like, configmaps, secret, pvc and sc
			err = ctrl.GetVolumesSpec(pod.Spec, pod.Namespace, backupClient)
			if err != nil {
				return err
			}

			// get service account relared objects
			err = ctrl.GetServiceAccountSpec(pod.Spec, pod.Namespace, backupClient)
			if err != nil {
				return err
			}

			// append the lables of pods to list
			podLabelList = append(podLabelList, pod.Labels)
		}
	}

	// get services used by pod
	err = ctrl.GetServiceForPod(namespace, podLabelList, backupClient, ctrl.kubeClient)
	if err != nil {
		return err
	}
	return nil
}

func (ctrl *controller) GetServiceForPod(namespace string, podLabelList []map[string]string,
	backupClient metaservice.MetaService_BackupClient, k8sClinet kubernetes.Interface) error {

	allServices, err := k8sClinet.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	var allServicesList []string
	for _, sc := range allServices.Items {
		allServicesList = append(allServicesList, sc.Name)
	}

	for _, service := range allServices.Items {
		serviceData, err := ctrl.GetService(namespace, service.Name)
		if err != nil {
			return err
		}

		for skey, svalue := range serviceData.Spec.Selector {
			for _, labels := range podLabelList {
				for lkey, lvalue := range labels {
					if skey == lkey && svalue == lvalue {
						err = ctrl.backupSend(serviceData, serviceData.Name, backupClient)
						if err != nil {
							return err
						}

					}
				}
			}
		}
	}

	return nil
}
