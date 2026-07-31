package update

import "testing"

func TestExtractErrorReason(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		want      string
		wantFound bool
	}{
		{
			name:      "single error line",
			output:    "Installing kanba v1.2.0 (linux/amd64)...\nError: release asset not found: https://...\n",
			want:      "release asset not found: https://...",
			wantFound: true,
		},
		{
			name:      "picks last error line among several",
			output:    "Error: unsupported OS\nsome other output\nError: checksum verification failed. Aborting install.",
			want:      "checksum verification failed. Aborting install.",
			wantFound: true,
		},
		{
			name:      "no error line falls back to not found",
			output:    "curl: command not found",
			want:      "",
			wantFound: false,
		},
		{
			name:      "empty output falls back to not found",
			output:    "",
			want:      "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		got, found := extractErrorReason(tt.output)
		if got != tt.want || found != tt.wantFound {
			t.Errorf("%s: extractErrorReason() = (%q, %v), want (%q, %v)", tt.name, got, found, tt.want, tt.wantFound)
		}
	}
}
