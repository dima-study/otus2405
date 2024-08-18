package hw10programoptimization

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestEmailReader_readLine(t *testing.T) {
	lines := []string{
		`{"Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov"}`,
		`{"Username":"qRichardson","Email":"mLynch@broWsecat.com"`,
		``,
		`{"Username":"tButler","Email":"5Moore@Teklist.net"}`,
	}
	data := strings.Join(lines, "\n") + "\n"

	tests := []struct {
		name         string
		n            int
		want         [][]byte
		wantErr      bool
		err          error
		skipErrEmpty bool
	}{
		{
			name: "#1",
			n:    2,
			want: [][]byte{
				[]byte(lines[0]),
				[]byte(lines[1]),
			},
			wantErr: false,
		},
		{
			name:    "#2",
			n:       3,
			want:    [][]byte{},
			wantErr: true,
			err:     ErrEmpty,
		},
		{
			name: "#3",
			n:    5,
			want: [][]byte{
				[]byte(lines[0]),
				[]byte(lines[1]),
				nil,
				[]byte(lines[3]),
				nil,
			},
			wantErr:      false,
			skipErrEmpty: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewEmailReader(bytes.NewBufferString(data))
			result := [][]byte{}
			var err error
			for range tt.n {
				var got []byte
				got, err = r.readLine()
				result = append(result, got)

				if err != nil {
					if tt.skipErrEmpty && errors.Is(err, ErrEmpty) {
						continue
					}
					break
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("%d x EmailReader.readLine() error = %v, wantErr %v", tt.n, err, tt.wantErr)
				return
			}

			if tt.err != nil && !errors.Is(err, tt.err) {
				t.Errorf("%d x EmailReader.readLine() want error = %v, got %v", tt.n, tt.err, err)
				return
			}

			if err == nil && !reflect.DeepEqual(result, tt.want) {
				t.Errorf("%d x EmailReader.readLine() = %v, want %v", tt.n, result, tt.want)
			}
		})
	}
}

func TestEmailReader_NextEmail(t *testing.T) {
	lines := []string{
		`{"Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov"}`,
		``,
		`{"Username":"tButler","Email":"5Moore@Teklist.net"}`,
	}
	data := strings.Join(lines, "\n") + "\n"

	tests := []struct {
		name         string
		n            int
		want         []string
		wantErr      bool
		err          error
		skipErrEmpty bool
	}{
		{
			name: "#1",
			n:    1,
			want: []string{
				"aliquid_qui_ea@browsedrive.gov",
			},
			wantErr: false,
		},
		{
			name:    "#2",
			n:       2,
			wantErr: true,
			err:     ErrEmpty,
		},
		{
			name:    "#3",
			n:       3,
			wantErr: true,
		},
		{
			name: "#3",
			n:    4,
			want: []string{
				"aliquid_qui_ea@browsedrive.gov",
				"",
				"5moore@teklist.net",
				"",
			},
			wantErr:      true,
			err:          io.EOF,
			skipErrEmpty: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewEmailReader(bytes.NewBufferString(data))
			result := []string{}
			var err error
			for range tt.n {
				var got string
				got, err = r.NextEmail()
				result = append(result, got)

				if err != nil {
					if tt.skipErrEmpty && errors.Is(err, ErrEmpty) {
						continue
					}
					break
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("%d x EmailReader.readLine() error = %v, wantErr %v", tt.n, err, tt.wantErr)
				return
			}

			if tt.err != nil && !errors.Is(err, tt.err) {
				t.Errorf("%d x EmailReader.readLine() want error = %v, got %v", tt.n, tt.err, err)
				return
			}

			if err == nil && !reflect.DeepEqual(result, tt.want) {
				t.Errorf("%d x EmailReader.readLine() = %v, want %v", tt.n, result, tt.want)
			}
		})
	}
}
