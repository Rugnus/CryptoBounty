// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IArbitrator} from "./IArbitrator.sol";

interface IArbitrable {
    event Dispute(IArbitrator indexed arbitrator, uint256 indexed disputeID, uint256 indexed metaEvidenceID, uint256 evidenceGroupID);
    event Ruling(IArbitrator indexed arbitrator, uint256 indexed disputeID, uint256 ruling);

    function rule(uint256 _disputeID, uint256 _ruling) external;
}
