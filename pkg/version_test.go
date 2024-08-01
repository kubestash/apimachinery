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
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "mongo-4",
			args: args{
				addonVersions: []string{"4.2.3", "4.4.6", "5.0.3", "5.0.15", "6.0.5"},
				dbVersion:     "4.4.26",
			},
			want: "4.4.6",
		},
		{
			name: "mongo-5-a",
			args: args{
				addonVersions: []string{"4.2.3", "4.4.6", "5.0.3", "5.0.15", "6.0.5"},
				dbVersion:     "5.0.2",
			},
			want: "5.0.3",
		},
		{
			name: "mongo-5-b",
			args: args{
				addonVersions: []string{"4.2.3", "4.4.6", "5.0.3", "5.0.15", "6.0.5"},
				dbVersion:     "5.0.15",
			},
			want: "5.0.15",
		},
		{
			name: "mongo-6",
			args: args{
				addonVersions: []string{"4.2.3", "4.4.6", "5.0.3", "5.0.15", "6.0.5"},
				dbVersion:     "6.0.12",
			},
			want: "6.0.5",
		},
		{
			name: "mysql-5",
			args: args{
				addonVersions: []string{"5.7.25", "8.0.3", "8.0.21"},
				dbVersion:     "5.7.42",
			},
			want: "5.7.25",
		},
		{
			name: "mysql-8.0",
			args: args{
				addonVersions: []string{"5.7.25", "8.0.3", "8.0.21"},
				dbVersion:     "8.0.5",
			},
			want: "8.0.3",
		},
		{
			name: "mysql-8.1",
			args: args{
				addonVersions: []string{"5.7.25", "8.0.3", "8.0.21"},
				dbVersion:     "8.1.0",
			},
			want: "8.0.21",
		},
		{
			name: "singlestore-8.1",
			args: args{
				addonVersions: []string{"alma-8.1.32-e3d3cde6da", "alma-8.5.7-bf633c1a54"},
				dbVersion:     "8.1.32",
			},
			want: "alma-8.1.32-e3d3cde6da",
		},
		{
			name: "singlestore-8.5",
			args: args{
				addonVersions: []string{"alma-8.1.32-e3d3cde6da", "alma-8.5.7-bf633c1a54"},
				dbVersion:     "8.5.72",
			},
			want: "alma-8.5.7-bf633c1a54",
		},
		{
			name: "pg-10",
			args: args{
				addonVersions: []string{"12.17.0", "14.10.0", "16.1.0"},
				dbVersion:     "10.23-bullseye",
			},
			want: "12.17.0",
		},
		{
			name: "no-addons",
			args: args{
				addonVersions: []string{},
				dbVersion:     "6.0.12",
			},
			wantErr: true,
		},
		//{
		//	name: "no-major-matched",
		//	args: args{
		//		addonVersions: []string{"4.2.3", "4.4.6", "6.0.5"},
		//		dbVersion:     "5.0.15",
		//	},
		//	wantErr: true,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindAppropriateAddonVersion(tt.args.addonVersions, tt.args.dbVersion)
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
