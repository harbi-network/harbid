package pow

import (
	"github.com/edsrzf/mmap-go"
	"golang.org/x/crypto/sha3"

	//"crypto/sha3"
	"encoding/binary"
	"os"
	"sync"
)

// ============================================================================

const (
	fnvPrime               = 0x01000193
	fullDatasetItemParents = 512
	numDatasetAccesses     = 32
	lightCacheRounds       = 3
	lightCacheNumItems     = 1179641
	fullDatasetNumItems    = 37748717
)

var (
	// we use the same seed as fish hash whitepaper for testnet debug reasons
	seed = [32]byte{
		0xeb, 0x01, 0x63, 0xae, 0xf2, 0xab, 0x1c, 0x5a,
		0x66, 0x31, 0x0c, 0x1c, 0x14, 0xd6, 0x0f, 0x42,
		0x55, 0xa9, 0xb3, 0x9b, 0x0e, 0xdf, 0x26, 0x53,
		0x98, 0x44, 0xf1, 0x17, 0xad, 0x67, 0x21, 0x19,
	}

	//sharedContext     *fishhashContext
	sharedContextLock sync.Mutex
)

type hash256 [32]byte
type hash512 [64]byte
type hash1024 [128]byte

type fishhashContext struct {
	ready               bool
	LightCacheNumItems  int
	LightCache          []*hash512
	FullDatasetNumItems uint32
	FullDataset         []hash1024
}

type itemState struct {
	cache         []*hash512
	numCacheItems int64
	seed          uint32
	mix           hash512
}

func (state *itemState) update(round uint32) {
	numWords := len(state.mix) / 4
	t := fnv1(state.seed^round, binary.LittleEndian.Uint32(state.mix[4*(round%uint32(numWords)):]))
	parentIndex := t % uint32(state.numCacheItems)
	state.mix = fnv1Hash512(state.mix, *state.cache[parentIndex])
}

func (state *itemState) final() hash512 {
	hash := sha3.NewLegacyKeccak512()
	hash.Write(state.mix[:])
	copy(state.mix[:], hash.Sum(nil))
	return state.mix
}

func bitwiseXOR(x, y hash512) hash512 {

	var result hash512
	for i := 0; i < 8; i++ {
		binary.LittleEndian.PutUint64(result[8*i:], binary.LittleEndian.Uint64(x[8*i:])^binary.LittleEndian.Uint64(y[8*i:]))
	}
	return result
}

func buildLightCache(cache []*hash512, numItems int, seed hash256) {

	item := hash512{}
	hash := sha3.NewLegacyKeccak512()
	hash.Write(seed[:])
	copy(item[:], hash.Sum(nil))
	cache[0] = &item

	for i := 1; i < numItems; i++ {
		hash.Reset()
		hash.Write(cache[i-1][:])
		newitem := hash512{}
		copy(newitem[:], hash.Sum(nil))
		cache[i] = &newitem
	}

	for q := 0; q < lightCacheRounds; q++ {
		for i := 0; i < numItems; i++ {
			indexLimit := uint32(numItems)
			t := binary.LittleEndian.Uint32(cache[i][0:])
			v := t % indexLimit
			w := uint32(numItems+(i-1)) % indexLimit
			x := bitwiseXOR(*cache[v], *cache[w])

			if i == 0 && q == 0 {
				var result hash512
				for k := 0; k < 8; k++ {
					binary.LittleEndian.PutUint64(result[8*k:], binary.LittleEndian.Uint64(cache[v][8*k:])^binary.LittleEndian.Uint64(cache[w][8*k:]))
				}

			}

			hash.Reset()
			hash.Write(x[:])
			copy(cache[i][:], hash.Sum(nil))
		}
	}
}

func buildDatasetSegment(ctx *fishhashContext, start, end uint32) {
	for i := start; i < end; i++ {
		ctx.FullDataset[i] = calculateDatasetItem1024(ctx, i)
	}
}

func prebuildDataset(ctx *fishhashContext, numThreads uint32) {
	log.Infof("Building prebuilt Dataset - we must be on miner")

	if ctx.FullDataset == nil {
		return
	}

	if ctx.ready == true {
		log.Infof("Dataset already generated")
		return
	}

	// TODO: dag file name (hardcoded for debug)
	// must parameterized this
	filename := "hashes.dat"

	log.Infof("Verifying if DAG local storage file already exists ...")
	hashes, err := loadmappedHashesFromFile(filename)
	if err == nil {
		log.Infof("DAG loaded succesfully from local storage ")
		ctx.FullDataset = hashes

		log.Debugf("debug DAG hash[10] : %x", ctx.FullDataset[10])
		log.Debugf("debug DAG hash[42] : %x", ctx.FullDataset[42])
		log.Debugf("debug DAG hash[12345] : %x", ctx.FullDataset[12345])
		ctx.ready = true
		return
	}

	log.Infof("DAG local storage file not found")
	log.Infof("GENERATING DATASET, This operation may take a while, please wait ...")

	if numThreads > 1 {
		log.Infof("Using multithread generation nb threads : %d", numThreads)
		batchSize := ctx.FullDatasetNumItems / numThreads
		var wg sync.WaitGroup

		for i := uint32(0); i < numThreads; i++ {
			start := i * batchSize
			end := start + batchSize
			if i == numThreads-1 {
				end = ctx.FullDatasetNumItems
			}

			wg.Add(1)
			go func(ctx *fishhashContext, start, end uint32) {
				defer wg.Done()
				buildDatasetSegment(ctx, start, end)
			}(ctx, start, end)
		}

		wg.Wait()
	} else {
		buildDatasetSegment(ctx, 0, ctx.FullDatasetNumItems)
	}

	log.Debugf("debug DAG hash[10] : %x", ctx.FullDataset[10])
	log.Debugf("debug DAG hash[42] : %x", ctx.FullDataset[42])
	log.Debugf("debug DAG hash[12345] : %x", ctx.FullDataset[12345])

	log.Infof("Saving DAG to local storage file ...")
	err = mapHashesToFile(ctx.FullDataset, filename)

	if err != nil {
		panic(err)
	}

	log.Infof("DATASET geneated succesfully")
	ctx.ready = true
}

func mapHashesToFile(hashes []hash1024, filename string) error {
	// Create or open fila
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// hash1024 table size (128 per object)
	size := len(hashes) * 128

	// file size setup
	err = file.Truncate(int64(size))
	if err != nil {
		return err
	}

	// Mapping the file in memory
	mmap, err := mmap.Map(file, mmap.RDWR, 0)
	if err != nil {
		return err
	}
	defer mmap.Unmap()

	// Copy data from memory to file
	for i, hash := range hashes {
		copy(mmap[i*128:(i+1)*128], hash[:])
	}

	// Sync data
	err = mmap.Flush()
	if err != nil {
		return err
	}

	return nil
}

func loadmappedHashesFromFile(filename string) ([]hash1024, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Mapping file in memory
	mmap, err := mmap.Map(file, mmap.RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer mmap.Unmap()

	// Get the nb of hash1028 (128bytes)
	numHashes := len(mmap) / 128
	hashes := make([]hash1024, numHashes)

	// Read data and convert in hash1024
	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], mmap[i*128:(i+1)*128])
	}

	return hashes, nil
}
