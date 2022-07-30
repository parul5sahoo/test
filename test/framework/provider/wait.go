/*
Copyright 2021 The Crossplane Authors.

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

package provider

import (
	"context"
	"log"
	"time"
	"fmt"

	v1 "github.com/crossplane/crossplane/apis/pkg/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)


// Wait for Provider to be successfully installed.
func WaitForAllProvidersInstalled(ctx context.Context, c client.Client, interval time.Duration, timeout time.Duration) error {
	if err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		l := &v1.ProviderList{}
		for item := range l.Items {
			log.Println("Providers Installed: ")
			log.Printf("%v", item)
		}
		if err := c.List(ctx, l); err != nil {
			return false, err
		}
		if len(l.Items) != 1 {
			log.Println("WARNING: The number of providers installed is not equal to 1")
			return false, nil
		}
		for _, p := range l.Items {
			if p.GetCondition(v1.TypeInstalled).Status != corev1.ConditionTrue {
				log.Println("ERROR: The Provider wasn't installed successfully.")
				return false, nil
			}
			if p.GetCondition(v1.TypeHealthy).Status != corev1.ConditionTrue {
				log.Println("ERROR: The Provider is not in a healthy state.")
				return false, nil
			}
		}
		return true, nil
	}); err != nil {
		return err
	}
	return nil
}

// Wait for Provider to be successfully updated.
func WaitForRevisionTransition(ctx context.Context, c client.Client, p2 string, p1 string, interval time.Duration, timeout time.Duration) error {
	if err := wait.PollImmediate(interval, timeout, func() (bool, error) {
		l := &v1.ProviderRevisionList{}
		for item := range l.Items {
			log.Println("Provider Revisions: ")
			log.Printf("%v", item)
		}
		if err := c.List(ctx, l); err != nil {
			return false, err
		}
		// There should be a revision present for the initial revision and the upgrade.
		if len(l.Items) != 2 {
			log.Println("WARNING: The number of provider revisions is not equal to 2")
			return false, nil
		}
		for _, p := range l.Items {
			// New ProviderRevision should be Active.
			if p.Spec.Package == p2 && p.GetDesiredState() != v1.PackageRevisionActive {
				log.Println("ERROR: The Provider revision is not active.")
				return false, nil
			}
			// Old ProviderRevision should be Inactive.
			if p.Spec.Package == p1 && p.GetDesiredState() != v1.PackageRevisionInactive {
				log.Println("ERROR: The Provider revision has not been discarded.")
				return false, nil
			}
			// Both ProviderRevisions should be healthy.
			if p.GetCondition(v1.TypeHealthy).Status != corev1.ConditionTrue {
				log.Println("ERROR: The Provider revision is not in a healthy state.")
				return false, nil
			}
		}
		return true, nil
	}); err != nil {
		return err
	}
	return nil
}

// Wait for Provider to be successfully deleted.
func WaitForAllProvidersDeleted(ctx context.Context, c client.Client, interval time.Duration, timeout time.Duration) error {
	return wait.PollImmediate(interval, timeout, func() (bool, error) {
		l := &v1.ProviderList{}
		for item := range l.Items {
			log.Println("Providers to be deleted: ")
			log.Printf("%v", item)
		}
		if err := c.List(ctx, l); err != nil {
			log.Printf("The Providers were not deleted successfully.")
			return false, err
		}
		return len(l.Items) == 0, nil
	})
}
