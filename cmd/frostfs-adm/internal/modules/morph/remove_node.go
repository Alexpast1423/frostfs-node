package morph

import (
	"errors"
	"fmt"

	netmapcontract "github.com/TrueCloudLab/frostfs-contract/netmap"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/io"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract/callflag"
	"github.com/nspcc-dev/neo-go/pkg/vm/emit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func removeNodesCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("at least one node key must be provided")
	}

	nodeKeys := make(keys.PublicKeys, len(args))
	for i := range args {
		var err error
		nodeKeys[i], err = keys.NewPublicKeyFromString(args[i])
		if err != nil {
			return fmt.Errorf("can't parse node public key: %w", err)
		}
	}

	wCtx, err := newInitializeContext(cmd, viper.GetViper())
	if err != nil {
		return fmt.Errorf("can't initialize context: %w", err)
	}
	defer wCtx.close()

	cs, err := wCtx.Client.GetContractStateByID(1)
	if err != nil {
		return fmt.Errorf("can't get NNS contract info: %w", err)
	}

	nmHash, err := nnsResolveHash(wCtx.ReadOnlyInvoker, cs.Hash, netmapContract+".frostfs")
	if err != nil {
		return fmt.Errorf("can't get netmap contract hash: %w", err)
	}

	bw := io.NewBufBinWriter()
	for i := range nodeKeys {
		emit.AppCall(bw.BinWriter, nmHash, "updateStateIR", callflag.All,
			int64(netmapcontract.NodeStateOffline), nodeKeys[i].Bytes())
	}

	if err := emitNewEpochCall(bw, wCtx, nmHash); err != nil {
		return err
	}

	if err := wCtx.sendConsensusTx(bw.Bytes()); err != nil {
		return err
	}

	return wCtx.awaitTx()
}
