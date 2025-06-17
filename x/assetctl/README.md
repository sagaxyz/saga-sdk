# `x/assetctl` - Asset Control Module

## Abstract

The `assetctl` module implements the Saga Asset Control architecture. This system is designed to manage a global directory of assets on a central **Controller** chain (the Hub) and enforce asset transfer rules for multiple **Host** chains (the chainlets).

This architecture has two primary components:
1.  **Controller (on the Hub):** A module that maintains a global, informational directory of all known assets and, more importantly, manages an access control list (ACL) for asset transfers on a per-chainlet basis.
2.  **Host (on Chainlets):** A module on each chainlet that communicates with the Controller. It allows chainlet owners to register their own native assets to the global directory and to specify which assets they want to support (i.e., allow for transfer).

## Concepts

The core concept is a separation of concerns. The Hub acts as a central authority for asset control, but the decision of *which* assets to support is delegated to each individual chainlet. Chainlets use their Host module to send Interchain Account (ICA) messages to the Hub's Controller module, which then executes the requested changes to the asset lists. An `AnteDecorator` on the Hub intercepts all IBC transfers (`MsgTransfer`) and validates them against the Controller's "supported assets" list, effectively enforcing the chainlets' preferences.

---

## **Controller (Hub)**

The Controller component resides on the Hub chainlet and acts as the central point of coordination for asset management.

### State

The Controller module maintains the following in its state:

*   **Global Asset Directory (`AssetMetadata`):** A collection mapping Hub-side IBC denoms to their `RegisteredAsset` protobuf, which contains the full `bank.v1beta1.Metadata`. This is for informational and discovery purposes.
    *   `AssetMetadataPrefix | HubIBCDenom -> RegisteredAsset`
*   **Supported Assets (`SupportedAssets`):** An access control list. The presence of a `(ChannelID, HubIBCDenom)` pair indicates that the asset is approved for transfer to the chainlet associated with that channel. This is the list used by the antehandler for validation.
    *   `SupportedAssetsPrefix | ChannelID | HubIBCDenom -> EmptyValue` (KeySet)
*   **Parameters (`Params`):** Module-specific parameters for the Controller.
    *   `ParamsPrefix -> Params`

### Messages

The Controller handles the following messages, which are expected to be sent from a Host module's ICA on a chainlet.

*   **`MsgManageRegisteredAssets`**
    *   *Purpose:* To add or remove assets from the global, informational `AssetMetadata` directory.
    *   `authority`: The chainlet's ICA address on the Hub.
    *   `channel_id`: The IBC channel connecting the chainlet to the Hub.
    *   `assets_to_register`: A list of `bank.v1beta1.Metadata` to register.
    *   `assets_to_unregister`: A list of Hub IBC denoms to unregister.

*   **`MsgManageSupportedAssets`**
    *   *Purpose:* To approve or deny the transfer of specific assets to the chainlet. This modifies the `SupportedAssets` ACL.
    *   `authority`: The chainlet's ICA address on the Hub.
    *   `channel_id`: The IBC channel ID for the chainlet.
    *   `add_ibc_denoms`: A list of Hub IBC denoms to allow for transfer.
    *   `remove_ibc_denoms`: A list of Hub IBC denoms to block from transfer.

*   **`MsgUpdateParams`**
    *   *Purpose:* To update the Controller module's parameters.
    *   `authority`: The address of the module's designated authority (e.g., governance).
    *   `params`: The new parameters to set.

### Queries

*   **`QueryAssetDirectory`**: Returns a paginated list of all assets in the global `AssetMetadata` directory.
*   **`QueryParams`**: Returns the current parameters of the Controller module.

### Antehandler

The Controller includes an `AnteDecorator` that intercepts every `MsgTransfer`. It checks if the `(SourceChannel, Token.Denom)` pair exists in the `SupportedAssets` list. If the pair is not found, the transaction is rejected.

### Client

```bash
# Query the global asset directory
appd query assetctl controller asset-directory

# Query the controller params
appd query assetctl controller params

# The following transactions are meant to be sent via a chainlet's ICA
# and would not typically be run manually from a user account.
appd tx assetctl controller manage-registered-assets ...
appd tx assetctl controller manage-supported-assets ...
```

---

## **Host (Chainlet)**

The Host component resides on each chainlet and acts as the interface for chainlet owners to manage their assets on the Hub.

### State

The Host module maintains the following in its state:

*   **ICA on Hub (`ICAData`):** Stores the address and channel information of the Interchain Account it has created on the Hub.
    *   `ICAOnHubPrefix -> ICAOnHub`
*   **In-Flight Requests (`InFlightRequests`):** A map of pending ICA request sequences to their message type. This tracks ongoing communications with the Controller.
    *   `InFlightRequestsPrefix | Sequence -> String(MsgTypeURL)`
*   **Parameters (`Params`):** Module-specific parameters for the Host.
    *   `ParamsPrefix -> Params`

### Messages

A chainlet owner (with `x/acl` permissions) can send the following messages to the Host module on their own chainlet. The Host module will then dispatch the corresponding ICA message to the Controller on the Hub.

*   **`MsgCreateICAOnHub`**
    *   *Purpose:* To establish the initial connection by creating an Interchain Account on the Hub, controlled by the Host module.
    *   `authority`: The chainlet admin's address.

*   **`MsgManageSupportedAssets`**
    *   *Purpose:* To tell the Controller on the Hub which assets this chainlet wants to support.
    *   `authority`: The chainlet admin's address.
    *   `add_ibc_denoms`: A list of IBC denoms (on the Hub) to support.
    *   `remove_ibc_denoms`: A list of IBC denoms to stop supporting.

*   **`MsgManageRegisteredAssets`**
    *   *Purpose:* To register the chainlet's own native assets in the Hub's global directory.
    *   `authority`: The chainlet admin's address.
    *   `assets_to_register`: A list of local denoms to register.
    *   `assets_to_unregister`: A list of local denoms to unregister.

*   **`MsgUpdateParams`**
    *   *Purpose:* To update the Host module's parameters.
    *   `authority`: The chainlet admin's address.
    *   `params`: The new parameters to set.

### Queries

*   **`QueryICAOnHub`**: Returns information about the Host's ICA on the Hub, such as its address and channel ID.
*   **`QueryParams`**: Returns the current parameters of the Host module.

### Client

```bash
# Create the ICA on the Hub to initialize the system
appd tx assetctl host create-ica-on-hub --from <admin_key>

# Add a supported asset for this chainlet
appd tx assetctl host manage-supported-assets --add <hub_ibc_denom> --from <admin_key>

# Register a native asset in the Hub's global directory
appd tx assetctl host manage-registered-assets --register <native_denom> --from <admin_key>

# Query the chainlet's ICA address on the Hub
appd query assetctl host ica-on-hub

# Query the host params
appd query assetctl host params
```

## Events

The `assetctl` module can emit the following events (TODO: add these):

*   `EventTypeRegisterAsset`: Emitted when a new asset is registered to the global directory.
    *   Attribute Keys: `ibc_denom`, `original_denom`
*   `EventTypeUnregisterAsset`: Emitted when an asset is unregistered from the global directory.
    *   Attribute Keys: `ibc_denom`
*   `EventTypeAddSupportedAsset`: Emitted when an asset is added to a chainlet's supported list.
    *   Attribute Keys: `chainlet_id`, `ibc_denom`
*   `EventTypeRemoveSupportedAsset`: Emitted when an asset is removed from a chainlet's supported list.
    *   Attribute Keys: `chainlet_id`, `ibc_denom`
*   `EventTypeAssetTransferBlocked`: Emitted when an ICS20 transfer is blocked by the antehandler.
    *   Attribute Keys: `chainlet_id`, `ibc_denom`

## Client

A user can interact with the `assetctl` module using gRPC or the command-line interface (CLI).

### CLI

TODO: to be tested

*Note: CLI commands are subject to change based on final implementation.*

```bash
# Query the global asset directory
appd query assetctl controller asset-directory --page <page_num> --limit <limit>

# Query the controller params
appd query assetctl controller params

# Register assets (example, sent from chainlet's ICA)
appd tx assetctl controller manage-registered-assets <path/to/assets.json> --from <chainlet-ica-address>

# Add supported assets (example, sent from chainlet's ICA)
appd tx assetctl controller manage-supported-assets --add <ibc_denom1,ibc_denom2> --channel-id <channel-id> --from <chainlet-ica-address>
```

### gRPC

Queries can be made via gRPC using the `saga.assetctl.controller.v1.Query` service.
Message submission would use the `saga.assetctl.controller.v1.Msg` service.







