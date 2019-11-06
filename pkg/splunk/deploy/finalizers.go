// Copyright (c) 2018-2019 Splunk Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deploy

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/splunk/splunk-operator/pkg/apis/enterprise/v1alpha1"
)

const (
	splunkFinalizerDeletePVC = "enterprise.splunk.com/delete-pvc"
)

// CheckSplunkDeletion checks to see if deletion was requested for the SplunkEnterprise resource.
// If so, it will process and remove any remaining finalizers.
func CheckSplunkDeletion(cr *v1alpha1.SplunkEnterprise, c client.Client) (bool, error) {
	currentTime := metav1.Now()

	// just log warning if deletion time is in the future
	if !cr.ObjectMeta.DeletionTimestamp.Before(&currentTime) {
		log.Printf("DeletionTimestamp is in the future for SplunkEnterprise %s/%s (Now='%s', DeletionTimestamp='%s')\n",
			cr.GetNamespace(), cr.GetIdentifier(), currentTime, cr.ObjectMeta.DeletionTimestamp)
		//return false, nil
	}

	log.Printf("Deletion requested for SplunkEnterprise %s/%s\n", cr.GetNamespace(), cr.GetIdentifier())

	// process each finalizer
	for _, finalizer := range cr.ObjectMeta.Finalizers {
		switch finalizer {
		case splunkFinalizerDeletePVC:
			if err := DeleteSplunkPvc(cr, c); err != nil {
				return false, err
			}
			if err := RemoveSplunkFinalizer(cr, c, finalizer); err != nil {
				return false, err
			}
		default:
			return false, fmt.Errorf("Finalizer in SplunkEnterprise %s/%s not recognized: %s", cr.GetNamespace(), cr.GetIdentifier(), finalizer)
		}
	}

	log.Printf("Deletion complete for SplunkEnterprise %s/%s\n", cr.GetNamespace(), cr.GetIdentifier())

	return true, nil
}

// DeleteSplunkPvc removes all corresponding PersistentVolumeClaims that are associated with a SplunkEnterprise resource.
func DeleteSplunkPvc(cr *v1alpha1.SplunkEnterprise, c client.Client) error {

	// get list of PVCs for this cluster
	labels := map[string]string{
		"app": "splunk",
		"for": cr.GetIdentifier(),
	}
	var listopts client.ListOptions
	listopts.InNamespace(cr.GetNamespace())
	listopts.MatchingLabels(labels)
	pvclist := corev1.PersistentVolumeClaimList{}
	if err := c.List(context.Background(), &listopts, &pvclist); err != nil {
		return err
	}

	// delete each PVC
	for _, pvc := range pvclist.Items {
		log.Printf("Deleting PVC for SplunkEnterprise %s/%s: %s\n", cr.GetNamespace(), cr.GetIdentifier(), pvc.ObjectMeta.Name)
		if err := c.Delete(context.Background(), &pvc); err != nil {
			return err
		}
	}

	return nil
}

// RemoveSplunkFinalizer removes a finalizer from a SplunkEnterprise resource.
func RemoveSplunkFinalizer(cr *v1alpha1.SplunkEnterprise, c client.Client, finalizer string) error {
	log.Printf("Removing finalizer for SplunkEnterprise %s/%s: %s\n", cr.GetNamespace(), cr.GetIdentifier(), finalizer)

	// create new list of finalizers that doesn't include the one being removed
	var newFinalizers []string

	// handles multiple occurrences (performance is not significant)
	for _, f := range cr.ObjectMeta.Finalizers {
		if f != finalizer {
			newFinalizers = append(newFinalizers, f)
		}
	}

	// update object
	cr.ObjectMeta.Finalizers = newFinalizers
	return c.Update(context.Background(), cr)
}
