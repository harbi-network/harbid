package client

import (
	"context"
	"time"

	"github.com/harbi-network/harbid/cmd/harbiwallet/daemon/server"

	"github.com/pkg/errors"

	"github.com/harbi-network/harbid/cmd/harbiwallet/daemon/pb"
	"google.golang.org/grpc"
)

// Connect connects to the harbiwalletd server, and returns the client instance
func Connect(address string) (pb.HarbiwalletdClient, func(), error) {
	// Connection is local, so 1 second timeout is sufficient
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(server.MaxDaemonSendMsgSize)))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, errors.New("harbiwallet daemon is not running, start it with `harbiwallet start-daemon`")
		}
		return nil, nil, err
	}

	return pb.NewHarbiwalletdClient(conn), func() {
		conn.Close()
	}, nil
}
