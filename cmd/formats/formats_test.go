package formats

import (
	"reflect"
	"testing"
)

func TestParseFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []FormatType
		wantErr bool
	}{
		{
			name:    "single format",
			input:   "mobi",
			want:    []FormatType{FormatMobi},
			wantErr: false,
		},
		{
			name:    "multiple formats",
			input:   "mobi,epub,kepub",
			want:    []FormatType{FormatMobi, FormatEpub, FormatKepub},
			wantErr: false,
		},
		{
			name:    "with spaces",
			input:   "mobi, epub, kepub",
			want:    []FormatType{FormatMobi, FormatEpub, FormatKepub},
			wantErr: false,
		},
		{
			name:    "mixed case",
			input:   "MOBI,ePub,KePUB",
			want:    []FormatType{FormatMobi, FormatEpub, FormatKepub},
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "mobi,invalid",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormats(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFormats() = %v, want %v", got, tt.want)
			}
		})
	}
}
