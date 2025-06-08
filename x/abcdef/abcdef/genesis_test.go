package abcdef_test

/*import (
	"testing"

	keepertest "github.com/sagaxyz/saga-sdk/testutil/keeper"
	"github.com/sagaxyz/saga-sdk/testutil/nullify"
	"github.com/sagaxyz/saga-sdk/x/abcdef"
	"github.com/sagaxyz/saga-sdk/x/abcdef/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:	types.DefaultParams(),
		PortId: types.PortID,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.AbcdefKeeper(t)
	abcdef.InitGenesis(ctx, k, genesisState)
	got := abcdef.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)

	// this line is used by starport scaffolding # genesis/test/assert
}*/
