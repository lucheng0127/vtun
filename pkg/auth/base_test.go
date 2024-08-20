package auth

import (
	"os"
	"sync"
	"testing"

	"bou.ke/monkey"
)

func TestBaseAuthMgr_ValidateUser(t *testing.T) {
	type args struct {
		user   string
		passwd string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				user:   "user2",
				passwd: "LaughOutLoudly",
			},
			wantErr: false,
		},
		{
			name: "user not exist",
			args: args{
				user:   "user3",
				passwd: "LaughOutLoudly",
			},
			wantErr: true,
		},
		{
			name: "wrong passwd",
			args: args{
				user:   "user1",
				passwd: "wrong passwd",
			},
			wantErr: true,
		},
	}

	mgr := &BaseAuthMgr{DB: "fake db file", AuthedUser: map[string]string{}, MLock: sync.Mutex{}}
	monkey.Patch(os.ReadFile, func(string) ([]byte, error) {
		return []byte(`user1,TGF1Z2hPdXRMb3VkbHk=
user2,TGF1Z2hPdXRMb3VkbHk=`), nil
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mgr.ValidateUser(tt.args.user, tt.args.passwd); (err != nil) != tt.wantErr {
				t.Errorf("BaseAuthMgr.ValidateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
