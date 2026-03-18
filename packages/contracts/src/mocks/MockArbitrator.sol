// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IArbitrator} from "../interfaces/IArbitrator.sol";
import {IArbitrable} from "../interfaces/IArbitrable.sol";

/// @notice Simple arbitrator for local testing: creates disputes and lets owner resolve by calling `giveRuling`.
contract MockArbitrator is IArbitrator {
    uint256 public cost;
    address public owner;

    struct Dispute {
        address arbitrated;
        DisputeStatus status;
        uint256 ruling;
    }

    Dispute[] public disputes;

    constructor(uint256 _cost) {
        cost = _cost;
        owner = msg.sender;
        disputes.push(Dispute({arbitrated: address(0), status: DisputeStatus.Solved, ruling: 0})); // 0-index unused
    }

    function arbitrationCost(bytes calldata) external view override returns (uint256) {
        return cost;
    }

    function createDispute(uint256, bytes calldata) external payable override returns (uint256 disputeID) {
        require(msg.value >= cost, "fee");
        disputes.push(Dispute({arbitrated: msg.sender, status: DisputeStatus.Waiting, ruling: 0}));
        disputeID = disputes.length - 1;
    }

    function disputeStatus(uint256 _disputeID) external view override returns (DisputeStatus) {
        return disputes[_disputeID].status;
    }

    function currentRuling(uint256 _disputeID) external view override returns (uint256 ruling) {
        return disputes[_disputeID].ruling;
    }

    function giveRuling(uint256 disputeID, uint256 ruling) external {
        require(msg.sender == owner, "not owner");
        Dispute storage d = disputes[disputeID];
        require(d.status == DisputeStatus.Waiting, "not waiting");
        d.status = DisputeStatus.Solved;
        d.ruling = ruling;
        IArbitrable(d.arbitrated).rule(disputeID, ruling);
    }
}

