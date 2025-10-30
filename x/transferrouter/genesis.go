package transferrouter

import (
	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/sagaxyz/saga-sdk/x/transferrouter/keeper"
	"github.com/sagaxyz/saga-sdk/x/transferrouter/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) []abci.ValidatorUpdate {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		panic(errorsmod.Wrap(err, "could not set parameters at genesis"))
	}

	privKey, err := crypto.HexToECDSA(data.Params.KnownSignerPrivateKey)
	if err != nil {
		panic(errorsmod.Wrap(err, "could not parse known signer private key"))
	}

	knownSignerAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	acc := k.AccountKeeper.NewAccountWithAddress(ctx, sdk.AccAddress(knownSignerAddr.Bytes()))
	k.AccountKeeper.SetAccount(ctx, acc)

	// TMP: add gateway contract address to active static precompiles
	err = k.EVMKeeper.EnableStaticPrecompiles(
		ctx,
		common.HexToAddress(data.Params.GatewayContractAddress),
	)
	if err != nil {
		panic(errorsmod.Wrap(err, "could not enable static precompile"))
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports the current module state to a genesis file.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(errorsmod.Wrap(err, "could not get parameters at genesis"))
	}

	return &types.GenesisState{
		Params: params,
	}
}
