package pkg

import "testing"

func TestFindAppropriateAddonVersion(t *testing.T) {
	type args struct {
		addonVersions []string
		dbVersion     string
	}
	tests := []struct {
		name string
		args args
		want string
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
			name: "mongo-5",
			args: args{
				addonVersions: []string{"4.2.3", "4.4.6", "5.0.3", "5.0.15", "6.0.5"},
				dbVersion:     "5.0.2",
			},
			want: "5.0.3",
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
				dbVersion:     "8.0.2",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindAppropriateAddonVersion(tt.args.addonVersions, tt.args.dbVersion)
			if err != nil {
				t.Errorf("FindAppropriateAddonVersion() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("FindAppropriateAddonVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}
