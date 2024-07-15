package merkle

import (
	"math"

	"github.com/harbi-network/harbid/domain/consensus/model/externalapi"
	"github.com/harbi-network/harbid/domain/consensus/utils/consensushashing"
	"github.com/harbi-network/harbid/domain/consensus/utils/hashes"
)

// nextPowerOfTwo returns the next highest power of two from a given number if
// it is not already a power of two. This is a helper function used during the
// calculation of a merkle tree.
func nextPowerOfTwo(n int) int {
	// Return the number if it's already a power of 2.
	if n&(n-1) == 0 {
		return n
	}

	// Figure out and return the next power of two.
	exponent := uint(math.Log2(float64(n))) + 1
	return 1 << exponent // 2^exponent
}

// hashMerkleBranches takes two hashes, treated as the left and right tree
// nodes, and returns the hash of their concatenation. This is a helper
// function used to aid in the generation of a merkle tree.
func hashMerkleBranches(left, right *externalapi.DomainHash) *externalapi.DomainHash {
	// Concatenate the left and right nodes.
	w := hashes.NewMerkleBranchHashWriter()

	w.InfallibleWrite(left.ByteSlice())
	w.InfallibleWrite(right.ByteSlice())

	return w.Finalize()
}

// CalculateHashMerkleRoot calculates the merkle root of a tree consisted of the given transaction hashes.
// See `merkleRoot` for more info.
func CalculateHashMerkleRoot(transactions []*externalapi.DomainTransaction) *externalapi.DomainHash {
	txHashes := make([]*externalapi.DomainHash, len(transactions))
	for i, tx := range transactions {
		txHashes[i] = consensushashing.TransactionHash(tx)
	}
	return merkleRoot(txHashes)
}

// CalculateIDMerkleRoot calculates the merkle root of a tree consisted of the given transaction IDs.
// See `merkleRoot` for more info.
func CalculateIDMerkleRoot(transactions []*externalapi.DomainTransaction) *externalapi.DomainHash {
	if len(transactions) == 0 {
		return &externalapi.DomainHash{}
	}

	txIDs := make([]*externalapi.DomainHash, len(transactions))
	for i, tx := range transactions {
		txIDs[i] = (*externalapi.DomainHash)(consensushashing.TransactionID(tx))
	}
	return merkleRoot(txIDs)
}

// merkleRoot creates a merkle tree from a slice of hashes, and returns its root.
func merkleRoot(hashes []*externalapi.DomainHash) *externalapi.DomainHash {
	// Calculate how many entries are required to hold the binary merkle
	// tree as a linear array and create an array of that size.
	nextPoT := nextPowerOfTwo(len(hashes))
	arraySize := nextPoT*2 - 1
	merkles := make([]*externalapi.DomainHash, arraySize)

	// Create the base transaction hashes and populate the array with them.
	for i, hash := range hashes {
		merkles[i] = hash
	}

	// Start the array offset after the last transaction and adjusted to the
	// next power of two.
	offset := nextPoT
	for i := 0; i < arraySize-1; i += 2 {
		switch {
		// When there is no left child node, the parent is nil too.
		case merkles[i] == nil:
			merkles[offset] = nil

		// When there is no right child, the parent is generated by
		// hashing the concatenation of the left child with zeros.
		case merkles[i+1] == nil:
			newHash := hashMerkleBranches(merkles[i], &externalapi.DomainHash{})
			merkles[offset] = newHash

		// The normal case sets the parent node to the hash
		// of the concatentation of the left and right children.
		default:
			newHash := hashMerkleBranches(merkles[i], merkles[i+1])
			merkles[offset] = newHash
		}
		offset++
	}

	return merkles[len(merkles)-1]
}
