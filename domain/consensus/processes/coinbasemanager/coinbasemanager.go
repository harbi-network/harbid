package coinbasemanager

import (
	"math"

	"github.com/harbi-network/harbid/domain/consensus/model"
	"github.com/harbi-network/harbid/domain/consensus/model/externalapi"
	"github.com/harbi-network/harbid/domain/consensus/utils/constants"
	"github.com/harbi-network/harbid/domain/consensus/utils/hashset"
	"github.com/harbi-network/harbid/domain/consensus/utils/subnetworks"
	"github.com/harbi-network/harbid/domain/consensus/utils/transactionhelper"
	"github.com/harbi-network/harbid/domain/consensus/utils/txscript"
	"github.com/harbi-network/harbid/infrastructure/db/database"
	"github.com/harbi-network/harbid/util"
	"github.com/pkg/errors"
)

type coinbaseManager struct {
	subsidyGenesisReward                    uint64
	preDeflationaryPhaseBaseSubsidy         uint64
	coinbasePayloadScriptPublicKeyMaxLength uint8
	genesisHash                             *externalapi.DomainHash
	deflationaryPhaseDaaScore               uint64
	deflationaryPhaseBaseSubsidy            uint64

	databaseContext     model.DBReader
	dagTraversalManager model.DAGTraversalManager
	ghostdagDataStore   model.GHOSTDAGDataStore
	acceptanceDataStore model.AcceptanceDataStore
	daaBlocksStore      model.DAABlocksStore
	blockStore          model.BlockStore
	pruningStore        model.PruningStore
	blockHeaderStore    model.BlockHeaderStore
}

func (c *coinbaseManager) ExpectedCoinbaseTransaction(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	coinbaseData *externalapi.DomainCoinbaseData) (expectedTransaction *externalapi.DomainTransaction, hasRedReward bool, err error) {

	ghostdagData, err := c.ghostdagDataStore.Get(c.databaseContext, stagingArea, blockHash, true)
	if !database.IsNotFoundError(err) && err != nil {
		return nil, false, err
	}

	// If there's ghostdag data with trusted data we prefer it because we need the original merge set non-pruned merge set.
	if database.IsNotFoundError(err) {
		ghostdagData, err = c.ghostdagDataStore.Get(c.databaseContext, stagingArea, blockHash, false)
		if err != nil {
			return nil, false, err
		}
	}

	acceptanceData, err := c.acceptanceDataStore.Get(c.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, false, err
	}

	daaAddedBlocksSet, err := c.daaAddedBlocksSet(stagingArea, blockHash)
	if err != nil {
		return nil, false, err
	}

	txOuts := make([]*externalapi.DomainTransactionOutput, 0, len(ghostdagData.MergeSetBlues()))
	acceptanceDataMap := acceptanceDataFromArrayToMap(acceptanceData)
	if constants.BlockVersion == 1 {
		for _, blue := range ghostdagData.MergeSetBlues() {
			txOut, hasReward, err := c.coinbaseOutputForBlueBlockV1(stagingArea, blue, acceptanceDataMap[*blue], daaAddedBlocksSet)
			if err != nil {
				return nil, false, err
			}
			if hasReward {
				txOuts = append(txOuts, txOut)
			}
		}
		txOut, hasRedReward, err := c.coinbaseOutputForRewardFromRedBlocksV1(
			stagingArea, ghostdagData, acceptanceData, daaAddedBlocksSet, coinbaseData)
		if err != nil {
			return nil, false, err
		}

		if hasRedReward {
			txOuts = append(txOuts, txOut)
		}
	} else if constants.BlockVersion == 2 {
		for _, blue := range ghostdagData.MergeSetBlues() {
			txOut, devTx, hasReward, err := c.coinbaseOutputForBlueBlockV2(stagingArea, blue, acceptanceDataMap[*blue], daaAddedBlocksSet)
			if err != nil {
				return nil, false, err
			}
			if hasReward {
				txOuts = append(txOuts, txOut)
				txOuts = append(txOuts, devTx)
			}
		}

	txOut, devTx, hasRedReward, err := c.coinbaseOutputForRewardFromRedBlocksV2(
			stagingArea, ghostdagData, acceptanceData, daaAddedBlocksSet, coinbaseData)
		if err != nil {
			return nil, false, err
		}

		if hasRedReward {
			txOuts = append(txOuts, txOut)
			txOuts = append(txOuts, devTx)
		}
	}

	subsidy, err := c.CalcBlockSubsidy(stagingArea, blockHash)
	if err != nil {
		return nil, false, err
	}

	payload, err := c.serializeCoinbasePayload(ghostdagData.BlueScore(), coinbaseData, subsidy)
	if err != nil {
		return nil, false, err
	}

	domainTransaction := &externalapi.DomainTransaction{
		Version:      constants.MaxTransactionVersion,
		Inputs:       []*externalapi.DomainTransactionInput{},
		Outputs:      txOuts,
		LockTime:     0,
		SubnetworkID: subnetworks.SubnetworkIDCoinbase,
		Gas:          0,
		Payload:      payload,
	}
	return domainTransaction, hasRedReward, nil
}

func (c *coinbaseManager) daaAddedBlocksSet(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (
	hashset.HashSet, error) {

	daaAddedBlocks, err := c.daaBlocksStore.DAAAddedBlocks(c.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return hashset.NewFromSlice(daaAddedBlocks...), nil
}

// coinbaseOutputForBlueBlock calculates the output that should go into the coinbase transaction of blueBlock
// If blueBlock gets no fee - returns nil for txOut
func (c *coinbaseManager) coinbaseOutputForBlueBlockV2(stagingArea *model.StagingArea,
	blueBlock *externalapi.DomainHash, blockAcceptanceData *externalapi.BlockAcceptanceData,
	mergingBlockDAAAddedBlocksSet hashset.HashSet) (*externalapi.DomainTransactionOutput, *externalapi.DomainTransactionOutput, bool, error) {
	blockReward, err := c.calcMergedBlockReward(stagingArea, blueBlock, blockAcceptanceData, mergingBlockDAAAddedBlocksSet)
	if err != nil {
		return nil, nil, false, err
	}
	
	devFeeDecodedAddress, err := util.DecodeAddress(constants.DevFeeAddress, util.Bech32PrefixHarbi)
	if err != nil {
		return nil, nil, false, err
		
	}
	devFeeScriptPublicKey, err := txscript.PayToAddrScript(devFeeDecodedAddress)
	if err != nil {
		return nil, nil, false, err
	}
	devFeeQuantity := uint64(float64(constants.DevFee) / 100 * float64(blockReward))
	blockReward -= devFeeQuantity
	if blockReward <= 0 {
		return nil, nil, false, nil
	}
	
	// the ScriptPublicKey for the coinbase is parsed from the coinbase payload
	_, coinbaseData, _, err := c.ExtractCoinbaseDataBlueScoreAndSubsidy(blockAcceptanceData.TransactionAcceptanceData[0].Transaction)
	if err != nil {
		return nil, nil, false, err
	}
	txOut := &externalapi.DomainTransactionOutput{
		Value:           blockReward,
		ScriptPublicKey: coinbaseData.ScriptPublicKey,
	}
	devTx := &externalapi.DomainTransactionOutput{
		Value:           devFeeQuantity,
		ScriptPublicKey: devFeeScriptPublicKey,
	}
	return txOut, devTx, true, nil
}
func (c *coinbaseManager) coinbaseOutputForBlueBlockV1(stagingArea *model.StagingArea,
	blueBlock *externalapi.DomainHash, blockAcceptanceData *externalapi.BlockAcceptanceData,
	mergingBlockDAAAddedBlocksSet hashset.HashSet) (*externalapi.DomainTransactionOutput, bool, error) {

	blockReward, err := c.calcMergedBlockReward(stagingArea, blueBlock, blockAcceptanceData, mergingBlockDAAAddedBlocksSet)
	if err != nil {
		return nil, false, err
	}

	if blockReward <= 0 {
		return nil, false, nil
	}

	// the ScriptPublicKey for the coinbase is parsed from the coinbase payload
	_, coinbaseData, _, err := c.ExtractCoinbaseDataBlueScoreAndSubsidy(blockAcceptanceData.TransactionAcceptanceData[0].Transaction)
	if err != nil {
		return nil, false, err
	}

	txOut := &externalapi.DomainTransactionOutput{
		Value:           blockReward,
		ScriptPublicKey: coinbaseData.ScriptPublicKey,
	}

	return txOut, true, nil
}

func (c *coinbaseManager) coinbaseOutputForRewardFromRedBlocksV2(stagingArea *model.StagingArea,
	ghostdagData *externalapi.BlockGHOSTDAGData, acceptanceData externalapi.AcceptanceData, daaAddedBlocksSet hashset.HashSet,
	coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransactionOutput, *externalapi.DomainTransactionOutput, bool, error) {
	acceptanceDataMap := acceptanceDataFromArrayToMap(acceptanceData)
	totalReward := uint64(0)
	for _, red := range ghostdagData.MergeSetReds() {
		reward, err := c.calcMergedBlockReward(stagingArea, red, acceptanceDataMap[*red], daaAddedBlocksSet)
		if err != nil {
			return nil, nil, false, err
		}
		totalReward += reward
	}
	devFeeDecodedAddress, err := util.DecodeAddress(constants.DevFeeAddress, util.Bech32PrefixHarbi)
	if err != nil {
		return nil, nil, false, err
	}
	devFeeScriptPublicKey, err := txscript.PayToAddrScript(devFeeDecodedAddress)
	if err != nil {
		return nil, nil, false, err
	}
	devFeeQuantity := uint64(float64(constants.DevFee) / 100 * float64(totalReward))
	totalReward -= devFeeQuantity
	if totalReward <= 0 {
		return nil, nil, false, nil
	}
	txOut := &externalapi.DomainTransactionOutput{
		Value:           totalReward,
		ScriptPublicKey: coinbaseData.ScriptPublicKey,
	}
	devTx := &externalapi.DomainTransactionOutput{
		Value:           devFeeQuantity,
		ScriptPublicKey: devFeeScriptPublicKey,
	}
	return txOut, devTx, true, nil
}
func (c *coinbaseManager) coinbaseOutputForRewardFromRedBlocksV1(stagingArea *model.StagingArea,
	ghostdagData *externalapi.BlockGHOSTDAGData, acceptanceData externalapi.AcceptanceData, daaAddedBlocksSet hashset.HashSet,
	coinbaseData *externalapi.DomainCoinbaseData) (*externalapi.DomainTransactionOutput, bool, error) {

	acceptanceDataMap := acceptanceDataFromArrayToMap(acceptanceData)
	totalReward := uint64(0)
	for _, red := range ghostdagData.MergeSetReds() {
		reward, err := c.calcMergedBlockReward(stagingArea, red, acceptanceDataMap[*red], daaAddedBlocksSet)
		if err != nil {
			return nil, false, err
		}

		totalReward += reward
	}

	if totalReward <= 0 {
		return nil, false, nil
	}

	txOut := &externalapi.DomainTransactionOutput{
		Value:           totalReward,
		ScriptPublicKey: coinbaseData.ScriptPublicKey,
	}
	return txOut, true, nil
}

func acceptanceDataFromArrayToMap(acceptanceData externalapi.AcceptanceData) map[externalapi.DomainHash]*externalapi.BlockAcceptanceData {
	acceptanceDataMap := make(map[externalapi.DomainHash]*externalapi.BlockAcceptanceData, len(acceptanceData))
	for _, blockAcceptanceData := range acceptanceData {
		acceptanceDataMap[*blockAcceptanceData.BlockHash] = blockAcceptanceData
	}
	return acceptanceDataMap
}

// CalcBlockSubsidy returns the subsidy amount a block at the provided blue score
// should have. This is mainly used for determining how much the coinbase for
// newly generated blocks awards as well as validating the coinbase for blocks
// has the expected value.
func (c *coinbaseManager) CalcBlockSubsidy(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (uint64, error) {
	if blockHash.Equal(c.genesisHash) {
		return c.subsidyGenesisReward, nil
	}
	blockDaaScore, err := c.daaBlocksStore.DAAScore(c.databaseContext, stagingArea, blockHash)
	if err != nil {
		return 0, err
	}
	if blockDaaScore < c.deflationaryPhaseDaaScore {
		return c.preDeflationaryPhaseBaseSubsidy, nil
	}

	blockSubsidy := c.calcDeflationaryPeriodBlockSubsidy(blockDaaScore)
	return blockSubsidy, nil
}

func (c *coinbaseManager) calcDeflationaryPeriodBlockSubsidy(blockDaaScore uint64) uint64 {
	// We define a year as 365.25 days and a month as 365.25 / 12 = 30.4375
	// secondsPerMonth = 30.4375 * 24 * 60 * 60
	const secondsPerMonth = 2629800
	// Note that this calculation implicitly assumes that block per second = 1 (by assuming daa score diff is in second units).
	monthsSinceDeflationaryPhaseStarted := (blockDaaScore - c.deflationaryPhaseDaaScore) / secondsPerMonth
	// Return the pre-calculated value from subsidy-per-month table
	return c.getDeflationaryPeriodBlockSubsidyFromTable(monthsSinceDeflationaryPhaseStarted)
}

/*
This table was pre-calculated by calling `calcDeflationaryPeriodBlockSubsidyFloatCalc` for all months until reaching 0 subsidy.
To regenerate this table, run `TestBuildSubsidyTable` in coinbasemanager_test.go (note the `deflationaryPhaseBaseSubsidy` therein)
*/
var subsidyByDeflationaryMonthTable = []uint64{
	3400000000, 2800000000, 2643284112, 2495339606, 2355675548, 2223828482, 2099360891, 1981859746, 1870935135, 1766218970, 1667363765, 1574041482, 1485942443, 1400000000, 1321642056, 1247669803, 1177837774, 1111914241, 1049680445, 990929872, 935467567, 883109485, 833681882, 787020740, 742971221, 700000000,
	660821028, 623834901, 588918887, 555957120, 524840222, 495464936, 467733783, 441554742, 416840941, 393510370, 371485610, 350000000, 330410514, 311917450, 294459443, 277978560, 264220111, 247732468, 233866891, 220777371, 208420470, 19675518, 185742805, 175000000, 165205257,
	155958725, 147229721, 138989280, 131210055, 123866234, 116933445, 110388685, 104210235, 98377592, 92871402, 87000000, 82130613, 77533766, 73194204, 69097527, 65230142, 61579213, 58132627, 54878946, 51807374, 48907717, 46170354, 43586000, 41065306, 38766883,
	36597102, 34548763, 32615071, 30789606, 29066313, 27439473, 25903687, 24453858, 23085177, 21793100, 20485451, 19338881, 18256485, 17234670, 16270046, 15359413, 14499747, 13688197, 12922069, 12198821, 11516053, 10871500, 10195524, 9624881, 9086177,
	8577624, 8097534, 7644316, 7216464, 6812558, 6431260, 6071302, 5731492, 5410700, 5097762, 4812440, 4543088, 4288812, 4048767, 3822158, 3608232, 3406279, 3215630, 3035651, 2865746, 2705350, 2548881, 2406220, 2271544, 2144406, 2024383, 1911079, 
	1804116, 1703139, 1607815, 1517825, 1432873, 1352675, 1158550, 1093706, 1032491, 974703, 920149, 868648, 820030, 774133, 730805, 689901, 651288, 613619, 579275, 546853, 516245, 487351, 460074, 434324, 410015, 387066,
	365402, 344950, 325644, 307417, 290211, 273968, 258634, 244158, 230493, 217592, 205413, 193916, 183063, 172817, 163144, 154013, 145393, 137255, 129573, 122321, 115475, 109011, 102910, 97150, 91713, 
	86579, 81734, 77159, 72840, 68763, 64915, 61281, 57851, 54613, 51557, 48671, 45947, 43375, 40948, 38656, 36492, 34450, 32521, 30701, 28983, 27361, 25829, 24384, 23019, 21730, 20514, 
	19366, 18282, 17259, 16293, 15381, 14520, 13707, 12940, 12216, 11532, 10886, 10277, 9702, 9159, 8646, 8162, 7705, 7273, 6865, 6480, 6117, 5774, 5450, 5144, 4856, 4584, 
	4327, 4084, 3855, 3639, 3435, 3242, 3060, 2888, 2726, 2573, 2428, 2292, 2163, 2041, 1926, 1818, 1716, 1619, 1528, 1442, 1361, 1284, 1212, 1144, 1079, 1018, 
	961, 907, 856, 808, 762, 719, 678, 640, 604, 570, 538, 507, 478, 451, 425, 401, 378, 356, 336, 317, 299, 282, 266, 251, 236, 222, 209, 197, 
	185, 174, 164, 154, 145, 136, 128, 120, 113, 106, 100, 94, 88, 83, 78, 73, 68, 64, 60, 56, 52, 49, 46, 43, 40, 37, 34, 
	32, 30, 28, 26, 24, 22, 20, 18, 16, 15, 14, 13, 12, 11, 10, 9, 8, 8, 7, 7, 6, 6, 6, 5, 5, 5, 
	4, 4, 4, 4, 3, 3, 3, 3, 3, 2, 2, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0,
}

func (c *coinbaseManager) getDeflationaryPeriodBlockSubsidyFromTable(month uint64) uint64 {
	if month >= uint64(len(subsidyByDeflationaryMonthTable)) {
		month = uint64(len(subsidyByDeflationaryMonthTable) - 1)
	}
	return subsidyByDeflationaryMonthTable[month]
}

func (c *coinbaseManager) calcDeflationaryPeriodBlockSubsidyFloatCalc(month uint64) uint64 {
	baseSubsidy := c.deflationaryPhaseBaseSubsidy
	subsidy := float64(baseSubsidy) / math.Pow(2, float64(month)/12)
	return uint64(subsidy)
}

func (c *coinbaseManager) calcMergedBlockReward(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	blockAcceptanceData *externalapi.BlockAcceptanceData, mergingBlockDAAAddedBlocksSet hashset.HashSet) (uint64, error) {

	if !blockHash.Equal(blockAcceptanceData.BlockHash) {
		return 0, errors.Errorf("blockAcceptanceData.BlockHash is expected to be %s but got %s",
			blockHash, blockAcceptanceData.BlockHash)
	}

	if !mergingBlockDAAAddedBlocksSet.Contains(blockHash) {
		return 0, nil
	}

	totalFees := uint64(0)
	for _, txAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
		if txAcceptanceData.IsAccepted {
			totalFees += txAcceptanceData.Fee
		}
	}

	block, err := c.blockStore.Block(c.databaseContext, stagingArea, blockHash)
	if err != nil {
		return 0, err
	}

	_, _, subsidy, err := c.ExtractCoinbaseDataBlueScoreAndSubsidy(block.Transactions[transactionhelper.CoinbaseTransactionIndex])
	if err != nil {
		return 0, err
	}

	return subsidy + totalFees, nil
}

// New instantiates a new CoinbaseManager
func New(
	databaseContext model.DBReader,

	subsidyGenesisReward uint64,
	preDeflationaryPhaseBaseSubsidy uint64,
	coinbasePayloadScriptPublicKeyMaxLength uint8,
	genesisHash *externalapi.DomainHash,
	deflationaryPhaseDaaScore uint64,
	deflationaryPhaseBaseSubsidy uint64,

	dagTraversalManager model.DAGTraversalManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	acceptanceDataStore model.AcceptanceDataStore,
	daaBlocksStore model.DAABlocksStore,
	blockStore model.BlockStore,
	pruningStore model.PruningStore,
	blockHeaderStore model.BlockHeaderStore) model.CoinbaseManager {

	return &coinbaseManager{
		databaseContext: databaseContext,

		subsidyGenesisReward:                    subsidyGenesisReward,
		preDeflationaryPhaseBaseSubsidy:         preDeflationaryPhaseBaseSubsidy,
		coinbasePayloadScriptPublicKeyMaxLength: coinbasePayloadScriptPublicKeyMaxLength,
		genesisHash:                             genesisHash,
		deflationaryPhaseDaaScore:               deflationaryPhaseDaaScore,
		deflationaryPhaseBaseSubsidy:            deflationaryPhaseBaseSubsidy,

		dagTraversalManager: dagTraversalManager,
		ghostdagDataStore:   ghostdagDataStore,
		acceptanceDataStore: acceptanceDataStore,
		daaBlocksStore:      daaBlocksStore,
		blockStore:          blockStore,
		pruningStore:        pruningStore,
		blockHeaderStore:    blockHeaderStore,
	}
}
