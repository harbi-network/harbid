package main

import (
	"context"
	"fmt"

	"github.com/harbi-network/harbid/cmd/harbiwallet/daemon/client"
	"github.com/harbi-network/harbid/cmd/harbiwallet/daemon/pb"
	"github.com/harbi-network/harbid/cmd/harbiwallet/utils"
)

func balance(conf *balanceConfig) error {
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return err
	}
	defer tearDown()

	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()
	response, err := daemonClient.GetBalance(ctx, &pb.GetBalanceRequest{})
	if err != nil {
		return err
	}

	pendingSuffix := ""
	if response.Pending > 0 {
		pendingSuffix = " (pending)"
	}
	if conf.Verbose {
		pendingSuffix = ""
		println("-----------------------------------------------------------------------------------------------------------")
		println("             ____    ____    ____    _                      ____                    _____   _________      ")
		println("    |    |  |    |  |    |  |    |  | |       |    _    |  |    |  |       |       |       |___   ___|     ")
		println("    |____|  |____|  |____|  |____|  | |       |   | |   |  |____|  |       |       |_____      | |         ")
		println("    |    |  |    |  |   |   |    |  | |       |  |   |  |  |    |  |       |       |           | |         ")
		println("    |    |  |    |  |   |   |____|  |_|       |_|     |_|  |    |  |_____  |_____  |_____      |_|         ")
		println("-----------------------------------------------------------------------------------------------------------")
		println("Address                                                                       Available             Pending")
		println("-----------------------------------------------------------------------------------------------------------")
		for _, addressBalance := range response.AddressBalances {
			fmt.Printf("%s %s %s\n", addressBalance.Address, utils.FormatKas(addressBalance.Available), utils.FormatKas(addressBalance.Pending))
		}
		println("-----------------------------------------------------------------------------------------------------------")
		print("                                                 ")
	}
	fmt.Printf("Total balance, HAR %s %s%s\n", utils.FormatKas(response.Available), utils.FormatKas(response.Pending), pendingSuffix)

	return nil
}
