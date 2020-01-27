//nolint
package app

import (
	"io"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

var (
	genesisFile        string
	paramsFile         string
	exportParamsPath   string
	exportParamsHeight int
	exportStatePath    string
	exportStatsPath    string
	seed               int64
	initialBlockHeight int
	numBlocks          int
	blockSize          int
	enabled            bool
	verbose            bool
	lean               bool
	commit             bool
	period             int
	onOperation        bool // TODO Remove in favor of binary search for invariant violation
	allInvariants      bool
	genesisTime        int64
)

// DONTCOVER

// NewMTAppUNSAFE is used for debugging purposes only.
//
// NOTE: to not use this function with non-test code
func NewMTAppUNSAFE(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*baseapp.BaseApp),
) (mtapp *MTApp, keyMain, keyStaking *sdk.KVStoreKey, stakingKeeper staking.Keeper) {

	mtapp = NewMTApp(logger, db, traceStore, loadLatest, invCheckPeriod, baseAppOptions...)
	return mtapp, mtapp.keys[baseapp.MainStoreKey], mtapp.keys[staking.StoreKey], mtapp.stakingKeeper
}
