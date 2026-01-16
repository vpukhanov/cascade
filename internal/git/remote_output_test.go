package git

import "testing"

func TestLastRemoteURL(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "empty_output",
			output: "",
			want:   "",
		},
		{
			name:   "single_remote_url",
			output: "remote: Create pull request: https://example.com/org/repo/compare/branch?expand=1",
			want:   "https://example.com/org/repo/compare/branch?expand=1",
		},
		{
			name: "non_remote_ignored",
			output: "Create PR https://example.com/ignored\n" +
				"hint: another https://example.com/also-ignored",
			want: "",
		},
		{
			name: "multiple_remote_urls",
			output: "remote: View changes https://example.com/first\n" +
				"remote: Create pull request https://example.com/second\n" +
				"hint: https://example.com/ignored",
			want: "https://example.com/second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lastRemoteURL(tt.output); got != tt.want {
				t.Fatalf("LastRemoteURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
