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

package pkg

import "testing"

func TestFindAppropriateAddonVersion(t *testing.T) {
	type args struct {
		addonVersions []string
		dbVersion     string
		funcName      string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// MongoDB
		{
			name: "mongo-4",
			args: args{
				addonVersions: []string{"4.2.3", "4.4.6", "5.0.3", "5.0.15", "6.0.5", "8.0.3"},
				dbVersion:     "4.4.26",
			},
			want: "4.4.6",
		},
		{
			name: "mongo-5-a",
			args: args{
				addonVersions: []string{"4.2.3", "4.4.6", "5.0.3", "5.0.15", "6.0.5", "8.0.3"},
				dbVersion:     "5.0.2",
				funcName:      "mongodb-backup",
			},
			want: "5.0.3",
		},
		{
			name: "mongo-5-b",
			args: args{
				addonVersions: []string{"4.2.3", "4.4.6", "5.0.3", "5.0.15", "6.0.5", "8.0.3"},
				dbVersion:     "5.0.15",
				funcName:      "mongodb-backup",
			},
			want: "5.0.15",
		},
		{
			name: "mongo-6",
			args: args{
				addonVersions: []string{"4.2.3", "4.4.6", "5.0.3", "5.0.15", "6.0.5", "8.0.3"},
				dbVersion:     "6.0.12",
				funcName:      "mongodb-backup",
			},
			want: "6.0.5",
		},
		{
			name: "mongo-7",
			args: args{
				addonVersions: []string{"4.2.3", "4.4.6", "5.0.3", "5.0.15", "6.0.5", "8.0.3"},
				dbVersion:     "7.0.16",
				funcName:      "mongodb-backup",
			},
			want: "6.0.5",
		},
		{
			name: "mongo-8",
			args: args{
				addonVersions: []string{"4.2.3", "4.4.6", "5.0.3", "5.0.15", "6.0.5", "8.0.3"},
				dbVersion:     "8.0.4",
				funcName:      "mongodb-backup",
			},
			want: "8.0.3",
		},
		// MySQL
		{
			name: "mysql-5",
			args: args{
				addonVersions: []string{"5.7.25", "8.0.3", "8.0.21"},
				dbVersion:     "5.7.42",
				funcName:      "mysql-backup",
			},
			want: "5.7.25",
		},
		{
			name: "mysql-8.0",
			args: args{
				addonVersions: []string{"5.7.25", "8.0.3", "8.0.21"},
				dbVersion:     "8.0.5",
				funcName:      "mysql-backup",
			},
			want: "8.0.3",
		},
		{
			name: "mysql-8.1",
			args: args{
				addonVersions: []string{"5.7.25", "8.0.3", "8.0.21"},
				dbVersion:     "8.1.0",
				funcName:      "mysql-backup",
			},
			want: "8.0.21",
		},
		// Xtrabackup
		{
			name: "xtra-2",
			args: args{
				addonVersions: []string{"2.4.29", "8.0.35", "8.1.0", "8.2.0", "8.4.0"},
				dbVersion:     "5.7.25",
				funcName:      "mysql-physical-backup",
			},
			want: "2.4.29",
		},
		{
			name: "xtra-8.0",
			args: args{
				addonVersions: []string{"2.4.29", "8.0.35", "8.1.0", "8.2.0", "8.4.0"},
				dbVersion:     "8.0.35",
				funcName:      "mysql-physical-backup",
			},
			want: "8.0.35",
		},
		{
			name: "xtra-8.1",
			args: args{
				addonVersions: []string{"2.4.29", "8.0.35", "8.1.0", "8.2.0", "8.4.0"},
				dbVersion:     "8.1.0",
				funcName:      "mysql-physical-backup",
			},
			want: "8.1.0",
		},
		{
			name: "xtra-8.2",
			args: args{
				addonVersions: []string{"2.4.29", "8.0.35", "8.1.0", "8.2.0", "8.4.0"},
				dbVersion:     "8.2.0",
				funcName:      "mysql-physical-backup",
			},
			want: "8.2.0",
		},
		{
			name: "xtra-8.4",
			args: args{
				addonVersions: []string{"2.4.29", "8.0.35", "8.1.0", "8.2.0", "8.4.0"},
				dbVersion:     "8.4.0",
				funcName:      "mysql-physical-backup",
			},
			want: "8.4.0",
		},
		// SingleStore
		{
			name: "singlestore-8.1",
			args: args{
				addonVersions: []string{"alma-8.1.32-e3d3cde6da", "alma-8.5.7-bf633c1a54"},
				dbVersion:     "8.1.32",
				funcName:      "singlestore-backup",
			},
			want: "alma-8.1.32-e3d3cde6da",
		},
		{
			name: "singlestore-8.5",
			args: args{
				addonVersions: []string{"alma-8.1.32-e3d3cde6da", "alma-8.5.7-bf633c1a54"},
				dbVersion:     "8.5.72",
				funcName:      "singlestore-backup",
			},
			want: "alma-8.5.7-bf633c1a54",
		},
		// PostgreSQL
		{
			name: "pg-10",
			args: args{
				addonVersions: []string{"12.17", "14.10", "16.1", "17.2"},
				dbVersion:     "10.23",
				funcName:      "postgres-backup",
			},
			want: "12.17",
		},
		{
			name: "pg-12a",
			args: args{
				addonVersions: []string{"12.17", "14.10", "16.1", "17.2"},
				dbVersion:     "12.17",
				funcName:      "postgres-backup",
			},
			want: "12.17",
		},
		{
			name: "pg-12b",
			args: args{
				addonVersions: []string{"12.17", "14.10", "16.1", "17.2"},
				dbVersion:     "12.18",
				funcName:      "postgres-backup",
			},
			want: "12.17",
		},
		{
			name: "pg-13",
			args: args{
				addonVersions: []string{"12.17", "14.10", "16.1", "17.2"},
				dbVersion:     "13.14",
				funcName:      "postgres-backup",
			},
			want: "14.10",
		},
		{
			name: "pg-14",
			args: args{
				addonVersions: []string{"12.17", "14.10", "16.1", "17.2"},
				dbVersion:     "14.11",
				funcName:      "postgres-backup",
			},
			want: "14.10",
		},
		{
			name: "pg-15",
			args: args{
				addonVersions: []string{"12.17", "14.10", "16.1", "17.2"},
				dbVersion:     "15.6",
				funcName:      "postgres-backup",
			},
			want: "16.1",
		},
		{
			name: "pg-16",
			args: args{
				addonVersions: []string{"12.17", "14.10", "16.1", "17.2"},
				dbVersion:     "16.2",
				funcName:      "postgres-backup",
			},
			want: "16.1",
		},
		{
			name: "pg-17",
			args: args{
				addonVersions: []string{"12.17", "14.10", "16.1", "17.2"},
				dbVersion:     "17.2",
				funcName:      "postgres-backup",
			},
			want: "17.2",
		},
		// No Addons
		{
			name: "no-addons",
			args: args{
				addonVersions: []string{},
				dbVersion:     "6.0.12",
				funcName:      "no-backup",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindAppropriateAddonVersion(tt.args.addonVersions, tt.args.dbVersion, tt.args.funcName)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("FindAppropriateAddonVersion() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if got != tt.want {
				t.Errorf("FindAppropriateAddonVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}
