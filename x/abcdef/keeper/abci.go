package keeper

import "context"

func (k *Keeper) BeginBlock(ctx context.Context) error {
	return nil
}

func (k *Keeper) EndBlock(ctx context.Context) error {
	err := k.Send(ctx)
	if err != nil {
		return err
	}

	return nil
}
