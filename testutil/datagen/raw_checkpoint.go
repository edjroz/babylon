package datagen

import (
	"math/rand"

	"github.com/boljen/go-bitmap"

	"github.com/babylonchain/babylon/btctxformatter"
	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
)

// GenRandomBitmap generates a random bitmap for the validator set
// It returns a random bitmap and the number of validators in the subset
func GenRandomBitmap() (bitmap.Bitmap, int) {
	bmBytes := GenRandomByteArray(txformat.BitMapLength)
	bm := bitmap.Bitmap(bmBytes)
	numSubset := 0
	for i := 0; i < bm.Len(); i++ {
		if bitmap.Get(bm, i) {
			numSubset++
		}
	}
	return bm, numSubset
}

func GetRandomRawBtcCheckpoint() *btctxformatter.RawBtcCheckpoint {
	rawCkpt := GenRandomRawCheckpoint()
	return &btctxformatter.RawBtcCheckpoint{
		Epoch:            rawCkpt.EpochNum,
		LastCommitHash:   *rawCkpt.LastCommitHash,
		BitMap:           rawCkpt.Bitmap,
		SubmitterAddress: GenRandomByteArray(btctxformatter.AddressLength),
		BlsSig:           rawCkpt.BlsMultiSig.Bytes(),
	}
}

func GenRandomRawCheckpointWithMeta() *types.RawCheckpointWithMeta {
	ckptWithMeta := &types.RawCheckpointWithMeta{
		Ckpt:     GenRandomRawCheckpoint(),
		Status:   GenRandomStatus(),
		PowerSum: 0,
	}
	return ckptWithMeta
}

func GenRandomRawCheckpoint() *types.RawCheckpoint {
	randomHashBytes := GenRandomLastCommitHash()
	randomBLSSig := GenRandomBlsMultiSig()
	return &types.RawCheckpoint{
		EpochNum:       GenRandomEpochNum(),
		LastCommitHash: &randomHashBytes,
		Bitmap:         bitmap.New(types.BitmapBits),
		BlsMultiSig:    &randomBLSSig,
	}
}

// GenRandomSequenceRawCheckpointsWithMeta generates random checkpoints from epoch 0 to a random epoch
func GenRandomSequenceRawCheckpointsWithMeta() []*types.RawCheckpointWithMeta {
	var topEpoch, finalEpoch uint64
	epoch1 := GenRandomEpochNum()
	epoch2 := GenRandomEpochNum()
	if epoch1 > epoch2 {
		topEpoch = epoch1
		finalEpoch = epoch2
	} else {
		topEpoch = epoch2
		finalEpoch = epoch1
	}
	var checkpoints []*types.RawCheckpointWithMeta
	for e := uint64(0); e <= topEpoch; e++ {
		ckpt := GenRandomRawCheckpointWithMeta()
		ckpt.Ckpt.EpochNum = e
		if e <= finalEpoch {
			ckpt.Status = types.Finalized
		}
		checkpoints = append(checkpoints, ckpt)
	}

	return checkpoints
}

func GenSequenceRawCheckpointsWithMeta(tipEpoch uint64) []*types.RawCheckpointWithMeta {
	ckpts := make([]*types.RawCheckpointWithMeta, int(tipEpoch)+1)
	for e := uint64(0); e <= tipEpoch; e++ {
		ckpt := GenRandomRawCheckpointWithMeta()
		ckpt.Ckpt.EpochNum = e
		ckpts[int(e)] = ckpt
	}
	return ckpts
}

func GenerateBLSSigs(keys []bls12381.PrivateKey, msg []byte) []bls12381.Signature {
	var sigs []bls12381.Signature
	for _, privkey := range keys {
		sig := bls12381.Sign(privkey, msg)
		sigs = append(sigs, sig)
	}

	return sigs
}

func GenerateLegitimateRawCheckpoint(privKeys []bls12381.PrivateKey) *types.RawCheckpoint {
	// number of validators, at least 4
	n := len(privKeys)
	// ensure sufficient signers
	signerNum := n/3 + 1
	epochNum := GenRandomEpochNum()
	lch := GenRandomLastCommitHash()
	msgBytes := types.GetSignBytes(epochNum, lch)
	sigs := GenerateBLSSigs(privKeys[:signerNum], msgBytes)
	multiSig, _ := bls12381.AggrSigList(sigs)
	bm := bitmap.New(types.BitmapBits)
	for i := 0; i < signerNum; i++ {
		bm.Set(i, true)
	}
	btcCheckpoint := &types.RawCheckpoint{
		EpochNum:       epochNum,
		LastCommitHash: &lch,
		Bitmap:         bm,
		BlsMultiSig:    &multiSig,
	}

	return btcCheckpoint
}

func GenRandomLastCommitHash() types.LastCommitHash {
	return GenRandomByteArray(types.HashSize)
}

func GenRandomBlsMultiSig() bls12381.Signature {
	return GenRandomByteArray(bls12381.SignatureSize)
}

// GenRandomStatus generates random status except for Finalized
func GenRandomStatus() types.CheckpointStatus {
	return types.CheckpointStatus(rand.Int31n(int32(len(types.CheckpointStatus_name) - 1)))
}
