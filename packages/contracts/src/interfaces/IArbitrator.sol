// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

/// @notice Minimal Kleros-like arbitrator interface for MVP.
///         Designed so the escrow contract can be wired to real Kleros-style arbitrators
///         on testnets, and to a mock arbitrator locally.
interface IArbitrator {
    enum DisputeStatus {
        Waiting,
        Appealable,
        Solved
    }

    function createDispute(uint256 _choices, bytes calldata _extraData) external payable returns (uint256 disputeID);

    function arbitrationCost(bytes calldata _extraData) external view returns (uint256);

    function disputeStatus(uint256 _disputeID) external view returns (DisputeStatus);

    function currentRuling(uint256 _disputeID) external view returns (uint256 ruling);
}

