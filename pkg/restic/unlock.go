/*
Copyright AppsCode Inc. and Contributors

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

package restic

import (
	"bytes"
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (w *ResticWrapper) UnlockRepository(repository string) error {
	_, err := w.unlock(repository)
	return err
}

// getLockIDs lists every lock ID currently held in the repository.
func (w *ResticWrapper) getLockIDs(repository string) ([]string, error) {
	w.sh.ShowCMD = true
	out, err := w.listLocks(repository)
	if err != nil {
		return nil, err
	}
	return extractLockIDs(bytes.NewReader(out))
}

// getLockStats returns the decoded JSON for a single lock.
func (w *ResticWrapper) getLockStats(repository, lockID string) (*LockStats, error) {
	w.sh.ShowCMD = true
	out, err := w.lockStats(repository, lockID)
	if err != nil {
		return nil, err
	}
	return extractLockStats(out)
}

// scanLocks scans all locks and returns the pod name of any exclusive lock found and the count of non-exclusive locks.
func (w *ResticWrapper) scanLocks(repository string) (string, int, error) {
	klog.Infoln("Scanning locks in the repository...")
	ids, err := w.getLockIDs(repository)
	if err != nil {
		return "", 0, fmt.Errorf("failed to list locks: %w", err)
	}
	if len(ids) == 0 {
		klog.Infof("No locks found in repository: %s", repository)
		return "", 0, nil
	}

	var exclusiveLockPodName string
	nonExclusiveCount := 0
	for _, id := range ids {
		st, err := w.getLockStats(repository, id)
		if err != nil {
			return "", 0, fmt.Errorf("failed to get lock stats for %s: %w", id, err)
		}
		if st.Exclusive {
			// There's no chance to get multiple exclusive locks, so we can use the first one we find.
			if exclusiveLockPodName == "" {
				exclusiveLockPodName = st.Hostname
				klog.Infof("Found exclusive lock: %s (hostname: %s)", id, st.Hostname)
			}
		} else {
			nonExclusiveCount++
			klog.Infof("Found non-exclusive lock: %s (hostname: %s)", id, st.Hostname)
		}
	}
	return exclusiveLockPodName, nonExclusiveCount, nil
}

// removeNonExclusiveLocksIfAny removes all non-exclusive (stale) locks from the repository.
func (w *ResticWrapper) removeNonExclusiveLocksIfAny(repository string, nonExclusiveCount int) error {
	if nonExclusiveCount > 0 {
		klog.Infof("Found %d non-exclusive lock(s). Removing locks from repository: %s", nonExclusiveCount, repository)
		_, err := w.unlock(repository)
		if err != nil {
			return fmt.Errorf("failed to remove locks: %w", err)
		}
		klog.Infof("Successfully removed stale non-exclusive locks from repository: %s", repository)
	} else {
		klog.Infof("No non-exclusive locks found in repository: %s", repository)
	}

	return nil
}

// removeExclusiveLocksIfAny removes exclusive locks from the repository.
func (w *ResticWrapper) removeExclusiveLocksIfAny(rClient client.Client, repository string, namespace string, lockPodName string) error {
	if lockPodName != "" {
		klog.Infof("Found exclusive lock. Waiting for Pod %s to complete...", lockPodName)
		err := wait.PollUntilContextTimeout(
			context.Background(),
			5*time.Second,
			kutil.ReadinessTimeout,
			true,
			func(ctx context.Context) (bool, error) {
				klog.Infof("Checking Pod status: %s...", lockPodName)

				var pod corev1.Pod
				err := rClient.Get(context.Background(), client.ObjectKey{Name: lockPodName, Namespace: namespace}, &pod)
				switch {
				case errors.IsNotFound(err): // Pod gone → unlock
					klog.Infof("Pod %s not found. Assuming it has terminated. Attempting to unlock repository: %s", lockPodName, repository)
					_, unlockErr := w.unlock(repository)
					return true, unlockErr

				case err != nil: // API error → stop
					return false, fmt.Errorf("error fetching Pod %s: %w", lockPodName, err)

				case pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed: // Pod finished → unlock
					klog.Infof("Pod %s finished with phase %s. Attempting to unlock repository: %s", lockPodName, pod.Status.Phase, repository)
					_, unlockErr := w.unlock(repository)
					return true, unlockErr

				default: // Not finished yet → keep waiting
					klog.Infof("Pod %s is still running with phase %s. Waiting...", lockPodName, pod.Status.Phase)
					return false, nil
				}
			})
		if err != nil {
			return fmt.Errorf("failed to wait for Pod %s to terminate: %w", lockPodName, err)
		}
	} else {
		klog.Infof("No exclusive locks found in repository: %s", repository)
	}

	return nil
}

// EnsureNoExclusiveLock blocks until any exclusive lock is released.
// If a lock is held by a Running Pod, it waits; otherwise it unlocks.
func (w *ResticWrapper) EnsureNoExclusiveLock(rClient client.Client, namespace string) error {
	klog.Infoln("Ensuring no lock is held on any repositories...")

	for _, b := range w.Config.Backends {
		klog.Infof("Scanning locks for repository: %s", b.Repository)

		// Scan all locks once to get both exclusive lock info and non-exclusive count
		exclusiveLockPodName, nonExclusiveCount, err := w.scanLocks(b.Repository)
		if err != nil {
			return fmt.Errorf("failed to scan locks for repository %s: %w", b.Repository, err)
		}

		klog.Infof("Removing exclusive locks from repository: %s", b.Repository)
		if err := w.removeExclusiveLocksIfAny(rClient, b.Repository, namespace, exclusiveLockPodName); err != nil {
			return fmt.Errorf("failed to remove exclusive locks from repository %s: %w", b.Repository, err)
		}

		// If there's an exclusive lock, `removeExclusiveLocksIfAny` has already removed all locks.
		if exclusiveLockPodName == "" {
			klog.Infof("Removing non-exclusive locks from repository: %s", b.Repository)
			if err := w.removeNonExclusiveLocksIfAny(b.Repository, nonExclusiveCount); err != nil {
				return fmt.Errorf("failed to remove non-exclusive locks from repository %s: %w", b.Repository, err)
			}
		}
	}

	klog.Infoln("All repositories are free of exclusive locks.")
	return nil
}

/*
Exclusive Locks
- Only one exclusive lock can run at a time.
- It blocks all non-exclusive locks.

Non-Exclusivity Locks
- Multiple non-exclusive locks can run at the same time.
- They do not block other non-exclusive locks.
- They do block exclusive locks (writers).
*/
