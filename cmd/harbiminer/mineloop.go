package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/harbi-network/harbid/version"
	"github.com/harbi-network/harbid/app/appmessage"
	"github.com/harbi-network/harbid/cmd/harbiminer/templatemanager"
	"github.com/harbi-network/harbid/domain/consensus/model/externalapi"
	"github.com/harbi-network/harbid/domain/consensus/utils/consensushashing"
	"github.com/harbi-network/harbid/domain/consensus/utils/pow"
	"github.com/harbi-network/harbid/infrastructure/network/netadapter/router"
	"github.com/harbi-network/harbid/util"
)

var hashesTried uint64
var dagReady = false

const logHashRateInterval = 10 * time.Second

func mineLoop(client *minerClient, numberOfBlocks uint64, targetBlocksPerSecond float64, mineWhenNotSynced bool,
	miningAddr util.Address) error {
	errChan := make(chan error)
	doneChan := make(chan struct{})

	foundBlockChan := make(chan *externalapi.DomainBlock, router.DefaultMaxMessages/2)

	spawn("templatesLoop", func() {
		templatesLoop(client, miningAddr, errChan)
	})

	numCores := runtime.NumCPU()
	for i := 0; i < numCores; i++ {
		spawn(fmt.Sprintf("miningWorker-%d", i), func() {
			mineWorker(foundBlockChan, mineWhenNotSynced)
		})
	}

	spawn("blocksLoop", func() {
		const windowSize = 10
		hasBlockRateTarget := targetBlocksPerSecond != 0
		var windowTicker, blockTicker *time.Ticker
		if hasBlockRateTarget {
			windowRate := time.Duration(float64(time.Second) / (targetBlocksPerSecond / windowSize))
			blockRate := time.Duration(float64(time.Second) / (targetBlocksPerSecond * windowSize))
			log.Infof("Minimum average time per %d blocks: %s, smaller minimum time per block: %s", windowSize, windowRate, blockRate)
			windowTicker = time.NewTicker(windowRate)
			blockTicker = time.NewTicker(blockRate)
			defer windowTicker.Stop()
			defer blockTicker.Stop()
		}
		windowStart := time.Now()
		for blockIndex := 1; ; blockIndex++ {
			block := <-foundBlockChan
			if hasBlockRateTarget {
				<-blockTicker.C
				if (blockIndex % windowSize) == 0 {
					tickerStart := time.Now()
					<-windowTicker.C
					log.Infof("Finished mining %d blocks in: %s. slept for: %s", windowSize, time.Since(windowStart), time.Since(tickerStart))
					windowStart = time.Now()
				}
			}
			err := handleFoundBlock(client, block)
			if err != nil {
				errChan <- err
				return
			}
		}
	})

	logHashRate()

	select {
	case err := <-errChan:
		return err
	case <-doneChan:
		return nil
	}
}

func logHashRate() {
	spawn("logHashRate", func() {
		lastCheck := time.Now()
		for range time.Tick(logHashRateInterval) {
			if !dagReady {
				log.Infof("Generating DAG, please wait ...")
				continue
			}

			currentHashesTried := atomic.LoadUint64(&hashesTried)
			currentTime := time.Now()
			kiloHashesTried := float64(currentHashesTried) / 1000.0
			hashRate := kiloHashesTried / currentTime.Sub(lastCheck).Seconds()
			log.Infof("Current hash rate is %.2f Khash/s", hashRate)
			lastCheck = currentTime
			atomic.StoreUint64(&hashesTried, 0)
		}
	})
}

func handleFoundBlock(client *minerClient, block *externalapi.DomainBlock) error {
	blockHash := consensushashing.BlockHash(block)
	log.Infof("Submitting block %s to %s", blockHash, client.Address())

	rejectReason, err := client.SubmitBlock(block)
	if err != nil {
		if errors.Is(err, router.ErrTimeout) {
			log.Warnf("Got timeout while submitting block %s to %s: %s", blockHash, client.Address(), err)
			return client.Reconnect()
		}
		if errors.Is(err, router.ErrRouteClosed) {
			log.Debugf("Got route is closed while requesting block template from %s. "+
				"The client is most likely reconnecting", client.Address())
			return nil
		}
		if rejectReason == appmessage.RejectReasonIsInIBD {
			const waitTime = 1 * time.Second
			log.Warnf("Block %s was rejected because the node is in IBD. Waiting for %s", blockHash, waitTime)
			time.Sleep(waitTime)
			return nil
		}
		return fmt.Errorf("error submitting block %s to %s: %w", blockHash, client.Address(), err)
	}
	return nil
}

func mineWorker(foundBlockChan chan<- *externalapi.DomainBlock, mineWhenNotSynced bool) {
	for {
		block, err := mineNextBlock(mineWhenNotSynced)
		if err != nil {
			log.Errorf("Error mining block: %s", err)
			continue
		}
		foundBlockChan <- block
	}
}

func mineNextBlock(mineWhenNotSynced bool) (*externalapi.DomainBlock, error) {
	var nonce uint64
	_ = binary.Read(rand.Reader, binary.LittleEndian, &nonce)

	for {
		if !dagReady {
			continue
		}
		nonce++
		block, state := getBlockForMining(mineWhenNotSynced)
		state.Nonce = nonce
		atomic.AddUint64(&hashesTried, 1)
		if state.CheckProofOfWork() {
			mutHeader := block.Header.ToMutable()
			mutHeader.SetNonce(nonce)
			block.Header = mutHeader.ToImmutable()
			log.Infof("Found block %s with parents %s", consensushashing.BlockHash(block), block.Header.DirectParents())
			return block, nil
		}
	}
}

func getBlockForMining(mineWhenNotSynced bool) (*externalapi.DomainBlock, *pow.State) {
	tryCount := 0

	const sleepTime = 500 * time.Millisecond
	const sleepTimeWhenNotSynced = 5 * time.Second

	for {
		tryCount++
		shouldLog := (tryCount-1)%10 == 0
		template, state, isSynced := templatemanager.Get()
		if template == nil {
			if shouldLog {
				log.Info("Waiting for the initial template")
			}
			time.Sleep(sleepTime)
			continue
		}
		if !isSynced && !mineWhenNotSynced {
			if shouldLog {
				log.Warnf("harbid is not synced. Skipping current block template")
			}
			time.Sleep(sleepTimeWhenNotSynced)
			continue
		}

		return template, state
	}
}

func templatesLoop(client *minerClient, miningAddr util.Address, errChan chan error) {
	getBlockTemplate := func() {
		template, err := client.GetBlockTemplate(miningAddr.String(), "harbiminer-"+version.Version())
		if errors.Is(err, router.ErrTimeout) {
			log.Warnf("Got timeout while requesting block template from %s: %s", client.Address(), err)
			reconnectErr := client.Reconnect()
			if reconnectErr != nil {
				errChan <- reconnectErr
			}
			return
		}
		if errors.Is(err, router.ErrRouteClosed) {
			log.Debugf("Got route is closed while requesting block template from %s. "+
				"The client is most likely reconnecting", client.Address())
			return
		}
		if err != nil {
			errChan <- fmt.Errorf("error getting block template from %s: %w", client.Address(), err)
			return
		}
		err = templatemanager.Set(template, backendLog)
		dagReady = true
		if err != nil {
			errChan <- fmt.Errorf("error setting block template from %s: %w", client.Address(), err)
			return
		}
	}

	getBlockTemplate()
	const tickerTime = 500 * time.Millisecond
	ticker := time.NewTicker(tickerTime)
	for {
		select {
		case <-client.newBlockTemplateNotificationChan:
			getBlockTemplate()
			ticker.Reset(tickerTime)
		case <-ticker.C:
			getBlockTemplate()
		}
	}
}
