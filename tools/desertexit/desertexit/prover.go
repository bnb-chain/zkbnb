package desertexit

import (
	"encoding/json"
	"fmt"
	"github.com/bnb-chain/zkbnb-crypto/circuit/desert"
	"github.com/bnb-chain/zkbnb/common/prove"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/std"
	"github.com/zeromicro/go-zero/core/logx"
	"runtime"
)

var (
	verifyingKeys groth16.VerifyingKey
	provingKeys   []groth16.ProvingKey
	r1cs          constraint.ConstraintSystem
)

func (c *GenerateProof) proveDesert(cDesert *desert.Desert) (string, error) {
	std.RegisterHints()
	nbConstraints, err := prove.LoadR1CSLen(c.config.KeyPath + ".r1cslen")
	if err != nil {
		logx.Severe("r1cs nb constraints read error")
		panic("r1cs nb constraints read error")
	}

	r1cs = groth16.NewCS(ecc.BN254)
	r1cs.LoadFromSplitBinaryConcurrent(c.config.KeyPath, nbConstraints, nbConstraints, runtime.NumCPU())
	if err != nil {
		logx.Severe("r1cs init error")
		panic("r1cs init error")
	}
	logx.Infof("blockConstraints constraints: %d", r1cs.GetNbConstraints())
	logx.Info("finish compile blockConstraints")
	// read proving and verifying keys
	provingKeys, err = prove.LoadProvingKey(c.config.KeyPath)
	if err != nil {
		logx.Severe("provingKey loading error")
		panic("provingKey loading error")
	}
	verifyingKeys, err = prove.LoadVerifyingKey(c.config.KeyPath)
	if err != nil {
		logx.Severe("verifyingKey loading error")
		panic("verifyingKey loading error")
	}

	return c.doProveDesert(cDesert)

}

func (c *GenerateProof) doProveDesert(cDesert *desert.Desert) (string, error) {
	// Generate proof.
	blockProof, err := prove.GenerateDesertProof(r1cs, provingKeys, verifyingKeys, cDesert, c.config.KeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to generateProof, err: %v", err)
	}

	formattedProof, err := prove.FormatDesertProof(blockProof, cDesert.StateRoot, cDesert.Commitment)
	if err != nil {
		return "", fmt.Errorf("unable to format blockProof: %v", err)
	}

	// Marshal formatted proof.
	proofBytes, err := json.Marshal(formattedProof)
	if err != nil {
		return "", err
	}
	return string(proofBytes), err
}
