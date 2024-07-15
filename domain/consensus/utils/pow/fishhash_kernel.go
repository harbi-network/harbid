package pow

import (
	"github.com/harbi-network/harbid/domain/consensus/model/externalapi"
	"golang.org/x/crypto/sha3"

	//"crypto/sha3"
	"encoding/binary"
)

func mergeHashes(hash1, hash2 hash512) (result hash1024) {
	copy(result[:len(hash1)], hash1[:])
	copy(result[len(hash1):], hash2[:])
	return
}

func fnv1(u, v uint32) uint32 {
	return (u * fnvPrime) ^ v
}

func fnv1Hash512(u, v hash512) hash512 {
	var result hash512
	for j := 0; j < 16; j++ {
		binary.LittleEndian.PutUint32(result[4*j:], fnv1(binary.LittleEndian.Uint32(u[4*j:]), binary.LittleEndian.Uint32(v[4*j:])))
	}
	return result
}

func newItemState(ctx *fishhashContext, index int64) *itemState {
	state := &itemState{
		cache:         ctx.LightCache,
		numCacheItems: int64(ctx.LightCacheNumItems),
		seed:          uint32(index),
	}

	state.mix = *state.cache[index%state.numCacheItems]
	binary.LittleEndian.PutUint32(state.mix[0:], binary.LittleEndian.Uint32(state.mix[0:])^state.seed)

	hash := sha3.NewLegacyKeccak512()
	hash.Write(state.mix[:])
	copy(state.mix[:], hash.Sum(nil))

	return state
}

func calculateDatasetItem1024(ctx *fishhashContext, index uint32) hash1024 {
	item0 := newItemState(ctx, int64(index)*2)
	item1 := newItemState(ctx, int64(index)*2+1)

	for j := uint32(0); j < fullDatasetItemParents; j++ {
		item0.update(j)
		item1.update(j)
	}

	it0 := item0.final()
	it1 := item1.final()

	return mergeHashes(it0, it1)
}

func lookup(ctx *fishhashContext, index uint32) hash1024 {
	if ctx.FullDataset != nil {
		item := &ctx.FullDataset[index]
		if item[0] == 0 {
			*item = calculateDatasetItem1024(ctx, index)
		}
		return *item
	}
	return calculateDatasetItem1024(ctx, index)
}

func fishhashKernel(ctx *fishhashContext, seed hash512) hash256 {
	indexLimit := uint32(ctx.FullDatasetNumItems)
	mix := mergeHashes(seed, seed)

	//log.Debugf("lookup matrix : ")
	for i := uint32(0); i < numDatasetAccesses; i++ {

		p0 := binary.LittleEndian.Uint32(mix[0:]) % indexLimit
		p1 := binary.LittleEndian.Uint32(mix[4*4:]) % indexLimit
		p2 := binary.LittleEndian.Uint32(mix[8*4:]) % indexLimit

		fetch0 := lookup(ctx, p0)
		fetch1 := lookup(ctx, p1)
		fetch2 := lookup(ctx, p2)

		for j := 0; j < 32; j++ {
			binary.LittleEndian.PutUint32(
				fetch1[4*j:],
				fnv1(binary.LittleEndian.Uint32(mix[4*j:4*j+4]), binary.LittleEndian.Uint32(fetch1[4*j:4*j+4])))
			binary.LittleEndian.PutUint32(
				fetch2[4*j:],
				binary.LittleEndian.Uint32(mix[4*j:4*j+4])^binary.LittleEndian.Uint32(fetch2[4*j:4*j+4]))
		}

		//fmt.Printf("The NEW fetch1 is : %x \n", fetch1)
		//fmt.Printf("The NEW fetch2 is : %x \n", fetch2)

		for j := 0; j < 16; j++ {
			binary.LittleEndian.PutUint64(
				mix[8*j:],
				binary.LittleEndian.Uint64(fetch0[8*j:8*j+8])*binary.LittleEndian.Uint64(fetch1[8*j:8*j+8])+binary.LittleEndian.Uint64(fetch2[8*j:8*j+8]))
		}
		//log.Debugf("\n")
	}

	mixHash := hash256{}
	for i := 0; i < (len(mix) / 4); i += 4 {
		j := 4 * i
		h1 := fnv1(binary.LittleEndian.Uint32(mix[j:]), binary.LittleEndian.Uint32(mix[j+4:]))
		h2 := fnv1(h1, binary.LittleEndian.Uint32(mix[j+8:]))
		h3 := fnv1(h2, binary.LittleEndian.Uint32(mix[j+12:]))
		binary.LittleEndian.PutUint32(mixHash[i:], h3)
	}

	return mixHash
}

func fishHash(ctx *fishhashContext, hashin *externalapi.DomainHash) *externalapi.DomainHash {

	seed := hash512{}
	copy(seed[:], hashin.ByteSlice())

	output := fishhashKernel(ctx, seed)
	outputArray := [32]byte{}
	copy(outputArray[:], output[:])
	return externalapi.NewDomainHashFromByteArray(&outputArray)
}
