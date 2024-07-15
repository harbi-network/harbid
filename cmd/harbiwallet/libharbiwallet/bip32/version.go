package bip32

import "github.com/pkg/errors"

// BitcoinMainnetPrivate is the version that is used for
// bitcoin mainnet bip32 private extended keys.
// Ecnodes to xprv in base58.
var BitcoinMainnetPrivate = [4]byte{
	0x04,
	0x88,
	0xad,
	0xe4,
}

// BitcoinMainnetPublic is the version that is used for
// bitcoin mainnet bip32 public extended keys.
// Ecnodes to xpub in base58.
var BitcoinMainnetPublic = [4]byte{
	0x04,
	0x88,
	0xb2,
	0x1e,
}

// HarbiMainnetPrivate is the version that is used for
// harbi mainnet bip32 private extended keys.
// Ecnodes to xprv in base58.
var HarbiMainnetPrivate = [4]byte{
	0x03,
	0x8f,
	0x2e,
	0xf4,
}

// HarbiMainnetPublic is the version that is used for
// harbi mainnet bip32 public extended keys.
// Ecnodes to kpub in base58.
var HarbiMainnetPublic = [4]byte{
	0x03,
	0x8f,
	0x33,
	0x2e,
}

// HarbiTestnetPrivate is the version that is used for
// harbi testnet bip32 public extended keys.
// Ecnodes to ktrv in base58.
var HarbiTestnetPrivate = [4]byte{
	0x03,
	0x90,
	0x9e,
	0x07,
}

// HarbiTestnetPublic is the version that is used for
// harbi testnet bip32 public extended keys.
// Ecnodes to ktub in base58.
var HarbiTestnetPublic = [4]byte{
	0x03,
	0x90,
	0xa2,
	0x41,
}

// HarbidevnetPrivate is the version that is used for
// harbi devnet bip32 public extended keys.
// Ecnodes to kdrv in base58.
var HarbidevnetPrivate = [4]byte{
	0x03,
	0x8b,
	0x3d,
	0x80,
}

// HarbidevnetPublic is the version that is used for
// harbi devnet bip32 public extended keys.
// Ecnodes to xdub in base58.
var HarbidevnetPublic = [4]byte{
	0x03,
	0x8b,
	0x41,
	0xba,
}

// HarbiSimnetPrivate is the version that is used for
// harbi simnet bip32 public extended keys.
// Ecnodes to ksrv in base58.
var HarbiSimnetPrivate = [4]byte{
	0x03,
	0x90,
	0x42,
	0x42,
}

// HarbiSimnetPublic is the version that is used for
// harbi simnet bip32 public extended keys.
// Ecnodes to xsub in base58.
var HarbiSimnetPublic = [4]byte{
	0x03,
	0x90,
	0x46,
	0x7d,
}

func toPublicVersion(version [4]byte) ([4]byte, error) {
	switch version {
	case BitcoinMainnetPrivate:
		return BitcoinMainnetPublic, nil
	case HarbiMainnetPrivate:
		return HarbiMainnetPublic, nil
	case HarbiTestnetPrivate:
		return HarbiTestnetPublic, nil
	case HarbidevnetPrivate:
		return HarbidevnetPublic, nil
	case HarbiSimnetPrivate:
		return HarbiSimnetPublic, nil
	}

	return [4]byte{}, errors.Errorf("unknown version %x", version)
}

func isPrivateVersion(version [4]byte) bool {
	switch version {
	case BitcoinMainnetPrivate:
		return true
	case HarbiMainnetPrivate:
		return true
	case HarbiTestnetPrivate:
		return true
	case HarbidevnetPrivate:
		return true
	case HarbiSimnetPrivate:
		return true
	}

	return false
}
