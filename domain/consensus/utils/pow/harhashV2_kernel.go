package pow

import (
	"github.com/harbi-network/harbid/domain/consensus/model/externalapi"

	//"crypto/sha3"
	"encoding/binary"
)

func harhashV2Kernel(ctx *fishhashContext, seed hash512) hash256 {
	indexLimit := uint32(ctx.FullDatasetNumItems)
	mix := mergeHashes(seed, seed)

	//log.Debugf("lookup matrix : ")
	for i := uint32(0); i < numDatasetAccesses; i++ {

		mixGroup := [8]uint32{}
		for c := uint32(0); c < 8; c++ {
			mixGroup[c] = binary.LittleEndian.Uint32(mix[(4*4*c+0):]) ^ binary.LittleEndian.Uint32(mix[(4*4*c+4):]) ^ binary.LittleEndian.Uint32(mix[(4*4*c+8):]) ^ binary.LittleEndian.Uint32(mix[(4*4*c+12):])
		}

		p0 := (mixGroup[0] ^ mixGroup[3] ^ mixGroup[6]) % indexLimit
		p1 := (mixGroup[1] ^ mixGroup[4] ^ mixGroup[7]) % indexLimit
		p2 := (mixGroup[2] ^ mixGroup[5] ^ i) % indexLimit

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

func harhashV2(ctx *fishhashContext, hashin *externalapi.DomainHash) *externalapi.DomainHash {

	seed := hash512{}
	copy(seed[:], hashin.ByteSlice())

	output := harhashV2Kernel(ctx, seed)
	outputArray := [32]byte{}
	copy(outputArray[:], output[:])
	return externalapi.NewDomainHashFromByteArray(&outputArray)
}
