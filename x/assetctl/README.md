# `x/assetctl` - Asset Registry Module

## Abstract

The `assetctl` module is responsible for managing a global directory of bridgeable assets within the Hub and for tracking which chainlets have opted-in to utilize this registry. It allows for the registration and unregistration of assets into the global directory. An antehandler will check transfers against this registry, but only for chainlets that have explicitly enabled this feature.

## Concepts

### Asset Registration (Global Directory)

Assets are registered into the Hub's *global* asset directory, typically via an Interchain Account (ICA) message originating from a chainlet or by a governance process on the Hub itself. The registration process involves:
1.  An entity (e.g., chainlet via ICA, Hub governance) initiating a registration message for one or more assets.
2.  For each asset, the message includes its original bank denom name and metadata (display name, description, denom units).
3.  The Hub's `assetctl` module determines the IBC denom for the asset *on the Hub*, ensuring its uniqueness in the global directory, and then appends the asset's details to this directory.

IBC denom names on the Hub are used as the primary identifier for assets within the global directory.

### Asset Unregistration (Global Directory)

Assets can be unregistered from the Hub's global directory, again typically via ICA from a chainlet or by Hub governance. This removes them from the global directory. Transfers of these assets via the Hub would then be blocked by the antehandler for any chainlet that has the registry feature enabled.

### Asset Transfer Control (Antehandler)

The `assetctl` module, through a dedicated antehandler, can intercept ICS20 (token transfer) packets being processed by the Hub.
1.  The antehandler first checks if the destination chainlet (or source, depending on transfer direction and where the check is most relevant) has the asset registry feature enabled (via `EnabledList`).
2.  If the feature is enabled for that chainlet, the antehandler then inspects the asset being transferred.
3.  It verifies if the asset's IBC denom (on the Hub) is present in the Hub's global asset directory (`AssetMetadata`).
4.  If the feature is enabled AND the asset is in the directory, the ICS20 transfer continues.
5.  If the feature is enabled AND the asset is NOT in the directory, the transfer is rejected.
6.  If the feature is NOT enabled for the chainlet, the transfer bypasses this specific module's validation.

### System Components

*   **Asset Hub (this module):** Resides on the Hub chain. It services requests for asset registration/unregistration into the global directory and for chainlets to toggle their registry participation. Its antehandler controls ICS20 transfers against the global directory for opted-in chainlets.
*   **Asset Host (separate module on chainlets/Saga Mainnet):** Resides on each chainlet and Saga Mainnet. It authenticates owner requests and sends ICA messages to the Hub for asset registration/unregistration or for toggling the chainlet's registry status.

## State

The `assetctl` module maintains the following in its state:

*   **Enabled Chainlets (`EnabledList`):** A set of `ChainletID`s. Presence of a `ChainletID` indicates the chainlet has opted-in to the asset registry's transfer controls.
    *   `EnabledListPrefix | ChainletID -> EmptyValue` (KeySet)
*   **Global Asset Directory (`AssetMetadata`):** A collection mapping Hub IBC denoms to their metadata (original denom, display name, description, denom units). This serves as the central registry of all known and allowed assets.
    *   `AssetMetadataPrefix | HubIBCDenom -> ProtocolBuffer(RegisteredAsset)`

## Messages

The `assetctl` module handles the following messages (primarily via ICA, but direct messages could also be enabled):

*   **`MsgToggleChainletRegistry`** (Handled via ICA or direct)
    *   `creator`: The address of the initiator (e.g., ICA host module on chainlet, or a chainlet admin).
    *   `chainlet_id`: The ID of the chainlet.
    *   `enable`: Boolean flag to enable (true) or disable (false) the registry for this chainlet.
    *   *Action:* Adds or removes the `chainlet_id` from the `EnabledList`.

*   **`MsgManageRegisteredAssets`** (Handled via ICA or direct)
    *   `creator`: The address of the initiator (e.g., ICA host module on chainlet, or a Hub address with authority).
    *   `channel_id`: The IBC channel ID for the chainlet.
    *   `assets_to_register`: A list of assets to register, where each asset includes:
        *   `denom`: The original bank denom name of the asset on its source chain.
        *   `display_name`: Human-readable display name.
        *   `description`: Asset description.
        *   `denom_units`: List of denomination units (e.g., base, display).
    *   `assets_to_unregister`: A list of Hub IBC denoms to unregister.
    *   *Action:* Registers and/or unregisters assets in the Hub's global `AssetMetadata` directory.

## Queries

The `assetctl` module provides the following gRPC queries:

*   **`QueryAssetDirectory`**
    *   Request: `QueryAssetDirectoryRequest` (supports pagination)
    *   Response: `QueryAssetDirectoryResponse` (list of `RegisteredAsset` from `AssetMetadata`, pagination info)
    *   *Action:* Returns a paginated list of all assets registered in the Hub's global directory.
*   **`QueryChainletRegistryStatus`** (New Query)
    *   Request: `QueryChainletRegistryStatusRequest { chainlet_id: string }`
    *   Response: `QueryChainletRegistryStatusResponse { is_enabled: bool }`
    *   *Action:* Returns whether the asset registry feature is enabled for the given `chainlet_id`.

## Events

The `assetctl` module can emit the following events:

*   `EventTypeToggleChainletRegistry`: Emitted when a chainlet enables or disables the registry.
    *   Attribute Keys: `chainlet_id`, `enabled_status` (true/false)
*   `EventTypeRegisterAsset`: Emitted when a new asset is registered to the global directory.
    *   Attribute Keys: `ibc_denom`, `original_denom`
*   `EventTypeUnregisterAsset`: Emitted when an asset is unregistered from the global directory.
    *   Attribute Keys: `ibc_denom`
*   `EventTypeAssetTransferBlocked`: Emitted when an ICS20 transfer is blocked by the antehandler for an opted-in chainlet due to the asset not being in the global directory.
    *   Attribute Keys: `chainlet_id`, `ibc_denom`

## Client

A user can interact with the `assetctl` module using gRPC or the command-line interface (CLI).

### CLI

Scaffolding for CLI commands will be added for queries and transactions:

```bash
# Query the global asset directory
appd query assetctl asset-directory --page <page_num> --limit <limit>

# Query a chainlet's registry status
appd query assetctl chainlet-status <chainlet_id>

# Toggle a chainlet's registry status (example, if direct messages are allowed)
appd tx assetctl toggle-chainlet-registry <chainlet_id> <true|false> --from <key_or_address>

# Register assets (example)
appd tx assetctl register-assets <path/to/assets.json> --from <key_or_address>

# Unregister assets (example)
appd tx assetctl unregister-assets <ibc_denom1,ibc_denom2> --from <key_or_address>
```

### gRPC

Queries can be made via gRPC using the `saga.assetctl.v1.Query` service.
Message submission would use the `saga.assetctl.v1.Msg` service.

## Future Improvements

*   Detailed specification for the antehandler logic.
*   Governance mechanisms for managing the global asset registry directly on the Hub.
*   Clear definition of authority for who can toggle chainlet registry status (e.g., chainlet's ICA controller, specific admin accounts).



Setup

We start with a pre-existing ICA created by the chainlet on the Hub, the Hub somehow recognizes this ICA as the "controller" of the chainlet.
This will be created by a chainlet's module, which on the chainlet's end is very straightforward, but on the Hub's end might be a bit tricky, as there has to be a validation mechanism to ensure that the ICA is indeed the controller of the chainlet.

PFM has to be enabled in the Hub.











