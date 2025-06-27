package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(sdk.Context, []byte, interface{})
	Set(sdk.Context, []byte, interface{})
}

//type ChannelKeeper interface {
//	GetChannel(sdk.Context, string, string) (ibcchanneltypes.Channel, bool)
//}
