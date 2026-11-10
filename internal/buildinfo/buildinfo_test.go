package buildinfo

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	tests := []struct {
		name                  string
		version, date, commit string
		want                  string
	}{
		{
			name:    "all set",
			version: "v1.2.3",
			date:    "2026-11-04",
			commit:  "abc1234",
			want:    "Build version: v1.2.3\nBuild date: 2026-11-04\nBuild commit: abc1234\n",
		},
		{
			name: "none set",
			want: "Build version: N/A\nBuild date: N/A\nBuild commit: N/A\n",
		},
		{
			name:    "only version set",
			version: "v1.2.3",
			want:    "Build version: v1.2.3\nBuild date: N/A\nBuild commit: N/A\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			Print(&buf, tt.version, tt.date, tt.commit)
			assert.Equal(t, tt.want, buf.String())
		})
	}
}
