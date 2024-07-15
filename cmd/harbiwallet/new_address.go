package main

import (
	"context"
	"fmt"
	"github.com/harbi-network/harbid/cmd/harbiwallet/daemon/client"
	"github.com/harbi-network/harbid/cmd/harbiwallet/daemon/pb"
)

func newAddress(conf *newAddressConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	response, err := daemonClient.NewAddress(ctx, &pb.NewAddressRequest{})
	if err != nil {
		return err
	}
	println("-----------------------------------------------------------------------------------------------------------")
	println("             ____    ____    ____    _                      ____                    _____   _________      ")
	println("    |    |  |    |  |    |  |    |  | |       |    _    |  |    |  |       |       |       |___   ___|     ")
	println("    |____|  |____|  |____|  |____|  | |       |   | |   |  |____|  |       |       |_____      | |         ")
	println("    |    |  |    |  |   |   |    |  | |       |  |   |  |  |    |  |       |       |           | |         ")
	println("    |    |  |    |  |   |   |____|  |_|       |_|     |_|  |    |  |_____  |_____  |_____      |_|         ")
	println("-----------------------------------------------------------------------------------------------------------")
	fmt.Printf("New address:\n%s\n", response.Address)
	println("-----------------------------------------------------------------------------------------------------------")
	return nil
}
