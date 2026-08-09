package cli

import "testing"

func TestParseConfigAdd(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"missing server name", []string{"add"}, true},
		{"server name", []string{"add", "gitlab"}, false},
		{"too many arguments", []string{"add", "gitlab", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseConfig(
				tt.args,
				GlobalOption{ConfigPath: "config.toml"},
			)

			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("Got error: %v\nExpected error: %t\n", err, tt.wantErr)
			}
		})
	}
}
