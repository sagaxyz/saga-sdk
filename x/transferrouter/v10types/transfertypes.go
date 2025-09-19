package v10types

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/accounts/abi"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/unknownproto"

	sdkmath "cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibcerrors "github.com/cosmos/ibc-go/v8/modules/core/errors"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
)

// IBC transfer sentinel errors
var (
	ErrInvalidPacketTimeout    = errorsmod.Register(ModuleName, 42, "invalid packet timeout2")
	ErrInvalidDenomForTransfer = errorsmod.Register(ModuleName, 43, "invalid denomination for cross-chain transfer2")
	ErrInvalidVersion          = errorsmod.Register(ModuleName, 44, "invalid ICS20 version2")
	ErrInvalidAmount           = errorsmod.Register(ModuleName, 45, "invalid token amount2")
	ErrDenomNotFound           = errorsmod.Register(ModuleName, 46, "denomination not found2")
	ErrSendDisabled            = errorsmod.Register(ModuleName, 47, "fungible token transfers from this chain are disabled2")
	ErrReceiveDisabled         = errorsmod.Register(ModuleName, 48, "fungible token transfers to this chain are disabled2")
	ErrMaxTransferChannels     = errorsmod.Register(ModuleName, 49, "max transfer channels2")
	ErrInvalidAuthorization    = errorsmod.Register(ModuleName, 50, "invalid transfer authorization2")
	ErrInvalidMemo             = errorsmod.Register(ModuleName, 51, "invalid memo2")
	ErrForwardedPacketTimedOut = errorsmod.Register(ModuleName, 52, "forwarded packet timed out2")
	ErrForwardedPacketFailed   = errorsmod.Register(ModuleName, 53, "forwarded packet failed2")
	ErrAbiEncoding             = errorsmod.Register(ModuleName, 54, "encoding abi failed2")
	ErrAbiDecoding             = errorsmod.Register(ModuleName, 55, "decoding abi failed2")
	ErrReceiveFailed           = errorsmod.Register(ModuleName, 56, "receive packet failed2")
)

const (
	// V1 defines first version of the IBC transfer module
	V1 = "ics20-1"
)

// InternalTransferRepresentation defines a struct used internally by the transfer application to represent a fungible token transfer
type InternalTransferRepresentation struct {
	// the tokens to be transferred
	Token Token
	// the sender address
	Sender string
	// the recipient address on the destination chain
	Receiver string
	// optional memo
	Memo string
}

var (
	_ ibcexported.PacketData         = (*transfertypes.FungibleTokenPacketData)(nil)
	_ ibcexported.PacketDataProvider = (*transfertypes.FungibleTokenPacketData)(nil)
)

const (
	EncodingJSON     = "application/json"
	EncodingProtobuf = "application/x-protobuf"
	EncodingABI      = "application/x-solidity-abi"
)

// ValidateBasic is used for validating the token transfer.
// NOTE: The addresses formats are not validated as the sender and recipient can have different
// formats defined by their corresponding chains that are not known to IBC.
func (ftpd InternalTransferRepresentation) ValidateBasic() error {
	if strings.TrimSpace(ftpd.Sender) == "" {
		return errorsmod.Wrap(ibcerrors.ErrInvalidAddress, "sender address cannot be blank")
	}

	if strings.TrimSpace(ftpd.Receiver) == "" {
		return errorsmod.Wrap(ibcerrors.ErrInvalidAddress, "receiver address cannot be blank")
	}

	if err := ftpd.Token.Validate(); err != nil {
		return err
	}

	if len(ftpd.Memo) > transfertypes.MaximumMemoLength {
		return errorsmod.Wrapf(transfertypes.ErrInvalidMemo, "memo must not exceed %d bytes", transfertypes.MaximumMemoLength)
	}

	return nil
}

// GetCustomPacketData interprets the memo field of the packet data as a JSON object
// and returns the value associated with the given key.
// If the key is missing or the memo is not properly formatted, then nil is returned.
func (ftpd InternalTransferRepresentation) GetCustomPacketData(key string) any {
	if len(ftpd.Memo) == 0 {
		return nil
	}

	jsonObject := make(map[string]any)
	err := json.Unmarshal([]byte(ftpd.Memo), &jsonObject)
	if err != nil {
		return nil
	}

	memoData, found := jsonObject[key]
	if !found {
		return nil
	}

	return memoData
}

// GetPacketSender returns the sender address embedded in the packet data.
//
// NOTE:
//   - The sender address is set by the module which requested the packet to be sent,
//     and this module may not have validated the sender address by a signature check.
//   - The sender address must only be used by modules on the sending chain.
//   - sourcePortID is not used in this implementation.
func (ftpd InternalTransferRepresentation) GetPacketSender(sourcePortID string) string {
	return ftpd.Sender
}

// MarshalPacketData attempts to marshal the provided FungibleTokenPacketData into bytes with the provided encoding.
func MarshalPacketData(data transfertypes.FungibleTokenPacketData, ics20Version string, encoding string) ([]byte, error) {
	if ics20Version != V1 {
		panic("unsupported ics20 version")
	}

	switch encoding {
	case EncodingJSON:
		return json.Marshal(data)
	case EncodingProtobuf:
		return proto.Marshal(&data)
	case EncodingABI:
		return EncodeABIFungibleTokenPacketData(&data)
	default:
		return nil, errorsmod.Wrapf(ibcerrors.ErrInvalidType, "invalid encoding provided, must be either empty or one of [%q, %q], got %s", EncodingJSON, EncodingProtobuf, encoding)
	}
}

// UnmarshalPacketData attempts to unmarshal the provided packet data bytes into a InternalTransferRepresentation.
func UnmarshalPacketData(bz []byte, ics20Version string, encoding string) (InternalTransferRepresentation, error) {
	const failedUnmarshalingErrorMsg = "cannot unmarshal %s transfer packet data: %s"

	// Depending on the ics20 version, we use a different default encoding (json for V1, proto for V2)
	// and we have a different type to unmarshal the data into.
	var data proto.Message
	switch ics20Version {
	case V1:
		if encoding == "" {
			encoding = EncodingJSON
		}
		data = &transfertypes.FungibleTokenPacketData{}
	default:
		return InternalTransferRepresentation{}, errorsmod.Wrap(ErrInvalidVersion, ics20Version)
	}

	errorMsgVersion := "ICS20-V1"

	// Here we perform the unmarshaling based on the specified encoding.
	// The functions act on the generic "data" variable which is of type proto.Message (an interface).
	switch encoding {
	case EncodingJSON:
		if err := json.Unmarshal(bz, &data); err != nil {
			return InternalTransferRepresentation{}, errorsmod.Wrapf(ibcerrors.ErrInvalidType, failedUnmarshalingErrorMsg, errorMsgVersion, err.Error())
		}
	case EncodingProtobuf:
		if err := unknownproto.RejectUnknownFieldsStrict(bz, data, unknownproto.DefaultAnyResolver{}); err != nil {
			return InternalTransferRepresentation{}, errorsmod.Wrapf(ibcerrors.ErrInvalidType, failedUnmarshalingErrorMsg, errorMsgVersion, err.Error())
		}

		if err := proto.Unmarshal(bz, data); err != nil {
			return InternalTransferRepresentation{}, errorsmod.Wrapf(ibcerrors.ErrInvalidType, failedUnmarshalingErrorMsg, errorMsgVersion, err.Error())
		}
	case EncodingABI:
		if ics20Version != V1 {
			return InternalTransferRepresentation{}, errorsmod.Wrapf(ibcerrors.ErrInvalidType, "encoding %s is only supported for ICS20-V1", EncodingABI)
		}
		var err error
		data, err = DecodeABIFungibleTokenPacketData(bz)
		if err != nil {
			return InternalTransferRepresentation{}, errorsmod.Wrapf(ibcerrors.ErrInvalidType, failedUnmarshalingErrorMsg, errorMsgVersion, err.Error())
		}
	default:
		return InternalTransferRepresentation{}, errorsmod.Wrapf(ibcerrors.ErrInvalidType, "invalid encoding provided, must be either empty or one of [%q, %q, %q], got %s", EncodingJSON, EncodingProtobuf, EncodingABI, encoding)
	}

	// When the unmarshaling is done, we want to retrieve the underlying data type based on the value of ics20Version
	// Since it has to be v1, we convert the data to FungibleTokenPacketData and then call the conversion function to construct
	// the v2 type.
	datav1, ok := data.(*transfertypes.FungibleTokenPacketData)
	if !ok {
		// We should never get here, as we manually constructed the type at the beginning of the file
		return InternalTransferRepresentation{}, errorsmod.Wrapf(ibcerrors.ErrInvalidType, "cannot convert proto message into FungibleTokenPacketData")
	}
	// The call to ValidateBasic for V1 is done inside PacketDataV1toV2.
	return PacketDataV1ToV2(*datav1)
}

// PacketDataV1ToV2 converts a v1 packet data to a v2 packet data. The packet data is validated
// before conversion.
func PacketDataV1ToV2(packetData transfertypes.FungibleTokenPacketData) (InternalTransferRepresentation, error) {
	if err := packetData.ValidateBasic(); err != nil {
		return InternalTransferRepresentation{}, errorsmod.Wrapf(err, "invalid packet data")
	}

	denom := ExtractDenomFromPath(packetData.Denom)
	return InternalTransferRepresentation{
		Token: Token{
			Denom:  denom,
			Amount: packetData.Amount,
		},
		Sender:   packetData.Sender,
		Receiver: packetData.Receiver,
		Memo:     packetData.Memo,
	}, nil
}

// getICS20ABI returns an abi.Arguments slice describing the Solidity types of the struct.
func getICS20ABI() abi.Arguments {
	// Create the ABI types for each field.
	// The Solidity types used are:
	// - string for Denom, Sender, Receiver and Memo.
	// - uint256 for Amount.
	tupleType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{
			Name: "denom",
			Type: "string",
		},
		{
			Name: "sender",
			Type: "string",
		},
		{
			Name: "receiver",
			Type: "string",
		},
		{
			Name: "amount",
			Type: "uint256",
		},
		{
			Name: "memo",
			Type: "string",
		},
	})
	if err != nil {
		panic(err)
	}

	// Create an ABI argument representing our struct as a single tuple argument.
	arguments := abi.Arguments{
		{
			Type: tupleType,
		},
	}

	return arguments
}

// DecodeABIFungibleTokenPacketData decodes a solidity ABI encoded ics20lib.ICS20LibFungibleTokenPacketData
// and converts it into an ibc-go FungibleTokenPacketData.
func DecodeABIFungibleTokenPacketData(data []byte) (*transfertypes.FungibleTokenPacketData, error) {
	arguments := getICS20ABI()

	packetDataI, err := arguments.Unpack(data)
	if err != nil {
		return nil, errorsmod.Wrapf(ErrAbiDecoding, "failed to unpack data: %s", err)
	}

	packetData, ok := packetDataI[0].(struct {
		Denom    string   `json:"denom"`
		Sender   string   `json:"sender"`
		Receiver string   `json:"receiver"`
		Amount   *big.Int `json:"amount"`
		Memo     string   `json:"memo"`
	})
	if !ok {
		return nil, errorsmod.Wrapf(ErrAbiDecoding, "failed to parse packet data")
	}

	return &transfertypes.FungibleTokenPacketData{
		Denom:    packetData.Denom,
		Sender:   packetData.Sender,
		Receiver: packetData.Receiver,
		Amount:   packetData.Amount.String(),
		Memo:     packetData.Memo,
	}, nil
}

func EncodeABIFungibleTokenPacketData(data *transfertypes.FungibleTokenPacketData) ([]byte, error) {
	amount, ok := new(big.Int).SetString(data.Amount, 10)
	if !ok {
		return nil, errorsmod.Wrapf(ErrAbiEncoding, "failed to parse amount: %s", data.Amount)
	}

	packetData := struct {
		Denom    string   `json:"denom"`
		Sender   string   `json:"sender"`
		Receiver string   `json:"receiver"`
		Amount   *big.Int `json:"amount"`
		Memo     string   `json:"memo"`
	}{
		data.Denom,
		data.Sender,
		data.Receiver,
		amount,
		data.Memo,
	}

	arguments := getICS20ABI()
	// Pack the values in the order defined in the ABI.
	encodedData, err := arguments.Pack(packetData)
	if err != nil {
		return nil, errorsmod.Wrapf(ErrAbiEncoding, "failed to pack data: %s", err)
	}

	return encodedData, nil
}

// Hop defines a port ID, channel ID pair specifying a unique "hop" in a trace
type Hop struct {
	PortId    string `protobuf:"bytes,1,opt,name=port_id,json=portId,proto3" json:"port_id,omitempty"`
	ChannelId string `protobuf:"bytes,2,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
}

// Denom holds the base denom of a Token and a trace of the chains it was sent through.
type Denom struct {
	// the base token denomination
	Base string `protobuf:"bytes,1,opt,name=base,proto3" json:"base,omitempty"`
	// the trace of the token
	Trace []Hop `protobuf:"bytes,3,rep,name=trace,proto3" json:"trace"`
}

// ExtractDenomFromPath returns the denom from the full path.
func ExtractDenomFromPath(fullPath string) Denom {
	denomSplit := strings.Split(fullPath, "/")

	if denomSplit[0] == fullPath {
		return Denom{
			Base: fullPath,
		}
	}

	var (
		trace          []Hop
		baseDenomSlice []string
	)

	length := len(denomSplit)
	for i := 0; i < length; i += 2 {
		// The IBC specification does not guarantee the expected format of the
		// destination port or destination channel identifier. A short term solution
		// to determine base denomination is to expect the channel identifier to be the
		// one ibc-go specifies. A longer term solution is to separate the path and base
		// denomination in the ICS20 packet. If an intermediate hop prefixes the full denom
		// with a channel identifier format different from our own, the base denomination
		// will be incorrectly parsed, but the token will continue to be treated correctly
		// as an IBC denomination. The hash used to store the token internally on our chain
		// will be the same value as the base denomination being correctly parsed.
		if i < length-1 && length > 2 && (channeltypes.IsValidChannelID(denomSplit[i+1]) || clienttypes.IsValidClientID(denomSplit[i+1])) {
			trace = append(trace, Hop{PortId: denomSplit[i], ChannelId: denomSplit[i+1]})
		} else {
			baseDenomSlice = denomSplit[i:]
			break
		}
	}

	base := strings.Join(baseDenomSlice, "/")

	return Denom{
		Base:  base,
		Trace: trace,
	}
}

// Token defines a struct which represents a token to be transferred.
type Token struct {
	// the token denomination
	Denom Denom `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom"`
	// the token amount to be transferred
	Amount string `protobuf:"bytes,2,opt,name=amount,proto3" json:"amount,omitempty"`
}

// maxUint256 is the maximum value for a 256 bit unsigned integer.
var maxUint256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

// Validate validates a token denomination and amount.
func (t Token) Validate() error {
	if err := t.Denom.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid token denom")
	}

	amount, ok := sdkmath.NewIntFromString(t.Amount)
	if !ok {
		return errorsmod.Wrapf(ErrInvalidAmount, "unable to parse transfer amount (%s) into math.Int", t.Amount)
	}

	if !amount.IsPositive() {
		return errorsmod.Wrapf(ErrInvalidAmount, "amount must be strictly positive: got %d", amount)
	}

	return nil
}

// Validate performs a basic validation of the Denom fields.
func (d Denom) Validate() error {
	// NOTE: base denom validation cannot be performed as each chain may define
	// its own base denom validation
	if strings.TrimSpace(d.Base) == "" {
		return errorsmod.Wrap(ErrInvalidDenomForTransfer, "base denomination cannot be blank")
	}

	for _, hop := range d.Trace {
		if err := hop.Validate(); err != nil {
			return errorsmod.Wrap(err, "invalid trace")
		}
	}

	return nil
}

// Validate performs a basic validation of the Hop fields.
func (h Hop) Validate() error {
	if err := host.PortIdentifierValidator(h.PortId); err != nil {
		return errorsmod.Wrapf(err, "invalid hop source port ID %s", h.PortId)
	}
	if err := host.ChannelIdentifierValidator(h.ChannelId); err != nil {
		return errorsmod.Wrapf(err, "invalid hop source channel ID %s", h.ChannelId)
	}

	return nil
}

// Path returns the full denomination according to the ICS20 specification:
// trace + "/" + baseDenom
// If there exists no trace then the base denomination is returned.
func (d Denom) Path() string {
	if d.IsNative() {
		return d.Base
	}

	var sb strings.Builder
	for _, t := range d.Trace {
		sb.WriteString(t.String()) // nolint:revive // no error returned by WriteString
		sb.WriteByte('/')          //nolint:revive // no error returned by WriteByte
	}
	sb.WriteString(d.Base) //nolint:revive
	return sb.String()
}

// IsNative returns true if the denomination is native, thus containing no trace history.
func (d Denom) IsNative() bool {
	return len(d.Trace) == 0
}

// IBCDenom a coin denomination for an ICS20 fungible token in the format
// 'ibc/{hash(trace + baseDenom)}'. If the trace is empty, it will return the base denomination.
func (d Denom) IBCDenom() string {
	if d.IsNative() {
		return d.Base
	}

	return fmt.Sprintf("%s/%s", transfertypes.DenomPrefix, d.Hash())
}

// String returns the Hop in the format:
// <portID>/<channelID>
func (h Hop) String() string {
	return fmt.Sprintf("%s/%s", h.PortId, h.ChannelId)
}

// Hash returns the hex bytes of the SHA256 hash of the Denom fields using the following formula:
//
// hash = sha256(trace + "/" + baseDenom)
func (d Denom) Hash() cmtbytes.HexBytes {
	hash := sha256.Sum256([]byte(d.Path()))
	return hash[:]
}
