package alarm

import (
	"context"
	"fmt"
	"testing"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestSkack(t *testing.T) {
	type args struct {
		ctx context.Context
		msg string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "t-1",
			args: args{
				ctx: context.Background(),
				msg: "This is a test message!",
			},
		},
	}

	ValidateEnv()
	for _, tt := range tests {
		fmt.Println("============================== RUN")
		t.Run(tt.name, func(t *testing.T) {
			Slack(tt.args.ctx, tt.args.msg)
		})
	}
}
