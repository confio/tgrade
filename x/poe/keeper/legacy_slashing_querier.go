package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/confio/tgrade/x/poe/types"
)

var _ slashingtypes.QueryServer = &legacySlashingGRPCQuerier{}

type legacySlashingGRPCQuerier struct {
	keeper          ContractSource
	contractQuerier types.SmartQuerier
}

func NewLegacySlashingGRPCQuerier(keeper Keeper, contractQuerier types.SmartQuerier) *legacySlashingGRPCQuerier {
	return &legacySlashingGRPCQuerier{keeper: keeper, contractQuerier: contractQuerier}
}

// SigningInfo legacy support for cosmos-sdk signing info. Note that not all field are available on tgrade
func (g legacySlashingGRPCQuerier) SigningInfo(c context.Context, req *slashingtypes.QuerySigningInfoRequest) (*slashingtypes.QuerySigningInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	valAddr, err := sdk.AccAddressFromBech32(req.ConsAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "validator address")
	}
	return &slashingtypes.QuerySigningInfoResponse{
		ValSigningInfo: slashingtypes.ValidatorSigningInfo{
			Address:             valAddr.String(),
			StartHeight:         0,
			IndexOffset:         0,
			Tombstoned:          false,
			MissedBlocksCounter: 0,
		},
	}, nil
}

// SigningInfos is not supported and will return unimplemented error
func (g legacySlashingGRPCQuerier) SigningInfos(c context.Context, req *slashingtypes.QuerySigningInfosRequest) (*slashingtypes.QuerySigningInfosResponse, error) {
	logNotImplemented(sdk.UnwrapSDKContext(c), "SigningInfos")
	return nil, status.Error(codes.Unimplemented, "not available, yet")
}

// Params is not supported. Method returns default slashing module params.
func (g legacySlashingGRPCQuerier) Params(c context.Context, req *slashingtypes.QueryParamsRequest) (*slashingtypes.QueryParamsResponse, error) {
	return &slashingtypes.QueryParamsResponse{
		Params: slashingtypes.Params{
			SignedBlocksWindow:      0,
			MinSignedPerWindow:      sdk.ZeroDec(),
			DowntimeJailDuration:    0,
			SlashFractionDoubleSign: sdk.ZeroDec(),
			SlashFractionDowntime:   sdk.ZeroDec(),
		}}, nil
}
