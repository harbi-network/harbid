package main

import (
	"context"
	"fmt"
	"github.com/harbi-network/harbid/cmd/harbiwallet/daemon/client"
	"github.com/harbi-network/harbid/cmd/harbiwallet/daemon/pb"
)

func showAddresses(conf *showAddressesConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()

	response, err := daemonClient.ShowAddresses(ctx, &pb.ShowAddressesRequest{})
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
	fmt.Printf("All addresses (%d):\n", len(response.Address))
	println("-----------------------------------------------------------------------------------------------------------")
	for _, address := range response.Address {
		fmt.Println(address)
	}
    println("-----------------------------------------------------------------------------------------------------------")
	fmt.Printf("\nNote: the above are only addresses that were manually created by the 'new-address' command. If you want to see a list of all addresses, including change addresses, " +
		"that have a positive balance, use the command 'balance -v'\n")
	return nil
}
