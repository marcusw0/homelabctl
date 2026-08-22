package check

import "testing"

func FuzzNormalizeHTTPURL(f *testing.F) {
	seeds := []string{
		"example.com",
		"https://example.com:443/path",
		"HTTP://example.com",
		"[::1]:443",
		"bad_host",
		"",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		got, err := NormalizeHTTPURL(input)
		if err != nil {
			return
		}

		again, err := NormalizeHTTPURL(got)
		if err != nil {
			t.Fatalf("normalized: %q, was rejected: %v", got, err)
		}

		if again != got {
			t.Errorf("normalized first: %q, second: %q", got, again)
		}
	})
}

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		host    string
		wantErr bool
	}{
		{"example.com", false},
		{"bad_host.edu", true},
		{"example..com", true},
		{"host.i", false},
		{"exa/mple.com", true},
		{"test.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			err := ValidateHostname(tt.host)
			if err != nil && !tt.wantErr {
				t.Errorf("want: %v got: %v", tt.wantErr, err)
			}
			if err == nil && tt.wantErr {
				t.Errorf("want: %v got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestNormalizeHTTPURL(t *testing.T) {
	tests := []struct {
		host    string
		want    string
		wantErr bool
	}{
		{"example.com:443", "https://example.com:443", false},
		{"bad_host.edu", "", true},
		{"example..com", "", true},
		{"tcp://example.com", "", true},
		{"example.com:65536", "", true},
		{"http://exam.com", "http://exam.com", false},
		{"https://exam.com:80", "https://exam.com:80", false},
		{"httpss://example.com", "", true},
		{"://example.com", "", true},
		{"HTTP://exam.com", "http://exam.com", false},
		{"example.com:0", "", true},
		{"httPs://192.168.1.5", "https://192.168.1.5", false},
		{"httpbin.org", "https://httpbin.org", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got, err := NormalizeHTTPURL(tt.host)
			if got != tt.want {
				t.Errorf("want: %s got: %s", tt.want, got)
			}
			if err != nil && !tt.wantErr {
				t.Errorf("want: %v got: %v", tt.wantErr, err)
			}
			if err == nil && tt.wantErr {
				t.Errorf("want: %v got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateAddrPort(t *testing.T) {
	tests := []struct {
		host    string
		wantErr bool
	}{
		{"192.168.5.1:25", false},
		{"localhost:80", false},
		{"257.60.0.1:443", true},
		{"123.0.0.5:", true},
		{"5:80", true},
		{"162.1.2.3:68888", true},
		{"199..54.2:443", true},
		{"local.80:443", true},
		{"[::1]:443", false},
		{"example.com.:443", false},
		{"bad_host:443", true},
		{"example.com:0", true},
		{"example.com", true},
		{"-10.0.0.50:443", true},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			err := ValidateAddrPort(tt.host)
			if err != nil && !tt.wantErr {
				t.Errorf("want: %v got: %v", tt.wantErr, err)
			}
			if err == nil && tt.wantErr {
				t.Errorf("want: %v got: %v", tt.wantErr, err)
			}
		})
	}
}

// 127.0.0.1:443
// [::1]:443
// example.com:443
// example.com.:443
// bad_host:443
// example.com:0
// example.com:65536
// example.com
// httpbin.org
// HTTP://example.com
// ftp://example.com
// https://
