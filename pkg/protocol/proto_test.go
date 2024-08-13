package protocol

import (
	"reflect"
	"testing"

	"github.com/lucheng0127/vtun/pkg/cipher"
)

func TestDecode(t *testing.T) {
	// Init cipher
	nc := new(cipher.NonCipher)
	ac := &cipher.AESCipher{Key: "0123456789ABCDEF"}
	emptyPayload := make([]byte, 0)
	payload := []byte("I like friday night")
	flag := HDR_FLG_DAT

	type args struct {
		payload  []byte
		encipher cipher.Cipher
	}
	tests := []struct {
		name    string
		args    args
		want    uint16
		want1   []byte
		wantErr bool
	}{
		{
			name: "no cipher empty payload",
			args: args{
				payload:  emptyPayload,
				encipher: nc,
			},
			want:    flag,
			want1:   emptyPayload,
			wantErr: false,
		},
		{
			name: "no cipher payload",
			args: args{
				payload:  payload,
				encipher: nc,
			},
			want:    flag,
			want1:   payload,
			wantErr: false,
		},
		{
			name: "aes cipher empty payload",
			args: args{
				payload:  emptyPayload,
				encipher: ac,
			},
			want:    flag,
			want1:   emptyPayload,
			wantErr: false,
		},
		{
			name: "aes cipher payload",
			args: args{
				payload:  payload,
				encipher: ac,
			},
			want:    flag,
			want1:   payload,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// New pkt
			pkt, err := NewVTPkt(flag, tt.args.payload, tt.args.encipher)
			if err != nil {
				t.Errorf("New vtun packet %v", err)
				return
			}

			// Encode pkt
			stream, err := pkt.Encode()
			if err != nil {
				t.Errorf("Encode vtun packet %v", err)
				return
			}

			// Decode
			got, got1, err := Decode(stream, tt.args.encipher)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Decode() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Decode() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
