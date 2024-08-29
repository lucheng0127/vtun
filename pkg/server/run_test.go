package server

import (
	"testing"

	"bou.ke/monkey"
	"github.com/golang/mock/gomock"
	"github.com/lucheng0127/vtun/pkg/config"
	mock_server "github.com/lucheng0127/vtun/pkg/mock/server"
	"github.com/urfave/cli/v2"
	"github.com/vishvananda/netlink"
)

type PatchObj struct {
	PatchFunc  interface{}
	TargetFunc interface{}
}

func TestRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mock_server.NewMockSvc(ctrl)
	cfg := &config.ServerConfig{
		Port:     6123,
		IP:       "192.168.223.254/24",
		LogLevel: "info",
		Key:      "0123456789ABCDEF",
		IPRange:  "192.168.223.100-192.168.223.200",
	}
	cfgWrong := &config.ServerConfig{
		Port:     6123,
		IP:       "192.168.223.254",
		LogLevel: "info",
		Key:      "0123456789ABCDEF",
		IPRange:  "192.168.223.100-192.168.223.200",
	}

	type args struct {
		cCtx *cli.Context
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		patchList []*PatchObj
	}{
		{
			name:      "config file not exist",
			args:      args{cCtx: &cli.Context{}},
			wantErr:   true,
			patchList: []*PatchObj{},
		},
		{
			name:    "wrong ip addr",
			args:    args{cCtx: &cli.Context{}},
			wantErr: true,
			patchList: []*PatchObj{
				{
					PatchFunc: config.LoadServerConfigFile,
					TargetFunc: func(string) (*config.ServerConfig, error) {
						return cfgWrong, nil
					},
				},
			},
		},
		{
			name:    "ok",
			args:    args{cCtx: &cli.Context{}},
			wantErr: false,
			patchList: []*PatchObj{
				{
					PatchFunc: config.LoadServerConfigFile,
					TargetFunc: func(string) (*config.ServerConfig, error) {
						return cfg, nil
					},
				},
				{
					PatchFunc: NewServer,
					TargetFunc: func(string, string, string, int, *netlink.Addr, []string) (Svc, error) {
						return mockSvc, nil
					},
				},
			},
		},
	}
	for _, tt := range tests {
		// Monkey patch
		for _, obj := range tt.patchList {
			monkey.Patch(obj.PatchFunc, obj.TargetFunc)
		}

		if tt.name == "ok" {
			// mockSvc.EXPECT().HandleSignal(gomock.Any()).Times(1)
			mockSvc.EXPECT().Launch().Times(1)
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := Run(tt.args.cCtx); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
