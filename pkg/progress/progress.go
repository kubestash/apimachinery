/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package progress

import (
	"time"

	"gomodules.xyz/restic"
	"k8s.io/klog/v2"
)

const progressPollInterval = 10 * time.Second

func (pg *Progress) Start() {
	for _, b := range pg.wrapper.Config.Backends {
		pg.wg.Add(1)
		go func(repo string) {
			defer pg.wg.Done()
			pg.pollAndSetStatus(repo)
		}(b.Repository)
	}
}

func (pg *Progress) Stop() {
	pg.cancel()
	pg.wg.Wait()
}

func (pg *Progress) pollAndSetStatus(repo string) {
	ticker := time.NewTicker(progressPollInterval)
	defer ticker.Stop()

	var cursor int
	for {
		select {
		case <-pg.ctx.Done():
			return
		case <-ticker.C:
			status, next, err := pg.latestStatus(repo, cursor)
			if err != nil {
				klog.Infoln("error getting backup progress for repo", repo, err)
				continue
			}
			if status == nil {
				continue
			}
			if pg.backupSession != nil {
				if err := pg.setBackupProgress(repo, status); err != nil {
					klog.Infoln("error setting backup progress for repo", repo, err)
				}
			} else if pg.restoreSession != nil {
				if err := pg.setRestoreProgress(status); err != nil {
					klog.Infoln("error setting restore progress for repo", repo, err)
				}
			}
			cursor = next
		}
	}
}

func (pg *Progress) latestStatus(repo string, since int) (*restic.ResticStatus, int, error) {
	length, statuses, err := pg.wrapper.StatusSince(repo, since)
	if err != nil {
		return nil, since, err
	}
	if length > since && len(statuses) > 0 {
		return &statuses[len(statuses)-1], length, nil
	}
	return nil, since, nil
}

/*
##### Workload Backup:

status:
  components:
    dump:
      driver: Restic
      path: repository/v1/frequent-backup/dump
      phase: Running
      resticStats:
      - progress:
          backupDone: 393.9MB
          percentDone: 7.48%
          secondsElapsed: 60
          total: 5.268GB
          totalFiles: 2


##### Workload Restore:

status:
  components:
    dump:
      phase: Running
      progress:
        percentDone: 0.02%
        restoreDone: 1.3MB
        secondsElapsed: 10
        total: 5.268GB
        totalFiles: 2

##### Databases Backup:

status:
  components:
    dump:
      driver: Restic
      phase: Running
      resticStats:
      - progress:
          backupDone: 1.039GB
          secondsElapsed: 20
          totalFiles: 1


##### Databases Restore:



##### Utils:

watch " kubectl get snapshot -n demo pvc-backup-pvc-backup-frequent-backup-1777529812 \
  -o jsonpath='{.status.components}' | jq"

*/
