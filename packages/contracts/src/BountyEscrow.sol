// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ReentrancyGuard} from "openzeppelin-contracts/contracts/utils/ReentrancyGuard.sol";
import {Ownable} from "openzeppelin-contracts/contracts/access/Ownable.sol";
import {IERC20} from "openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";

import {IArbitrator} from "./interfaces/IArbitrator.sol";
import {IArbitrable} from "./interfaces/IArbitrable.sol";

/// @notice Escrow-based bounty contract (ETH + whitelisted ERC20) with Kleros-like arbitration.
contract BountyEscrow is ReentrancyGuard, Ownable, IArbitrable {
    enum Status {
        Created,
        Assigned,
        Submitted,
        Approved,
        Disputed,
        ResolvedPaidHunter,
        ResolvedRefundSponsor,
        PaidOut,
        Refunded,
        Cancelled
    }

    struct Bounty {
        address sponsor;
        address token; // address(0) = ETH
        uint256 amount;
        address hunter; // set on assignment
        Status status;
        string metadataURI;
        bytes32 metadataHash;
        uint64 createdAt;
        uint64 assignedAt;
        uint64 submittedAt;
        uint64 resolvedAt;
        uint64 payoutAt;
    }

    struct Application {
        bool exists;
        string messageURI; // off-chain (optional)
        uint64 createdAt;
    }

    event BountyCreated(
        uint256 indexed bountyId,
        address indexed sponsor,
        address indexed token,
        uint256 amount,
        string metadataURI,
        bytes32 metadataHash
    );
    event ApplicationSubmitted(uint256 indexed bountyId, address indexed hunter, string messageURI);
    event HunterAssigned(uint256 indexed bountyId, address indexed hunter);
    event WorkSubmitted(uint256 indexed bountyId, string workURI, bytes32 workHash);
    event Approved(uint256 indexed bountyId);
    event Disputed(uint256 indexed bountyId, uint256 indexed disputeId);
    event PaidOut(uint256 indexed bountyId, address indexed to, uint256 amount);
    event Refunded(uint256 indexed bountyId, address indexed to, uint256 amount);
    event Cancelled(uint256 indexed bountyId);

    uint256 public constant RULING_OPTIONS = 2;
    uint256 public constant RULING_SPONSOR_WINS = 1;
    uint256 public constant RULING_HUNTER_WINS = 2;

    IArbitrator public arbitrator;
    bytes public arbitratorExtraData;
    uint256 public metaEvidenceId;

    uint256 public bountyCount;
    mapping(uint256 => Bounty) public bounties;
    mapping(uint256 => mapping(address => Application)) public applications; // bountyId => hunter => application

    // disputeId => bountyId
    mapping(uint256 => uint256) public disputeToBounty;

    // token whitelist (only for ERC20)
    mapping(address => bool) public tokenWhitelist;

    constructor(IArbitrator _arbitrator, bytes memory _extraData, uint256 _metaEvidenceId) Ownable(msg.sender) {
        arbitrator = _arbitrator;
        arbitratorExtraData = _extraData;
        metaEvidenceId = _metaEvidenceId;
    }

    function setTokenWhitelist(address token, bool allowed) external onlyOwner {
        tokenWhitelist[token] = allowed;
    }

    function setArbitrator(IArbitrator _arbitrator, bytes calldata _extraData, uint256 _metaEvidenceId) external onlyOwner {
        arbitrator = _arbitrator;
        arbitratorExtraData = _extraData;
        metaEvidenceId = _metaEvidenceId;
    }

    function createBountyETH(string calldata metadataURI, bytes32 metadataHash) external payable nonReentrant returns (uint256 bountyId) {
        require(msg.value > 0, "amount=0");
        bountyId = _createBounty(msg.sender, address(0), msg.value, metadataURI, metadataHash);
    }

    function createBountyERC20(address token, uint256 amount, string calldata metadataURI, bytes32 metadataHash)
        external
        nonReentrant
        returns (uint256 bountyId)
    {
        require(token != address(0), "token=0");
        require(tokenWhitelist[token], "token not allowed");
        require(amount > 0, "amount=0");
        _safeTransferFrom(token, msg.sender, address(this), amount);
        bountyId = _createBounty(msg.sender, token, amount, metadataURI, metadataHash);
    }

    function apply(uint256 bountyId, string calldata messageURI) external {
        Bounty storage b = bounties[bountyId];
        require(b.sponsor != address(0), "not found");
        require(b.status == Status.Created, "not open");
        Application storage a = applications[bountyId][msg.sender];
        require(!a.exists, "already applied");
        applications[bountyId][msg.sender] = Application({exists: true, messageURI: messageURI, createdAt: uint64(block.timestamp)});
        emit ApplicationSubmitted(bountyId, msg.sender, messageURI);
    }

    function assignHunter(uint256 bountyId, address hunter) external {
        Bounty storage b = bounties[bountyId];
        require(b.sponsor == msg.sender, "not sponsor");
        require(b.status == Status.Created, "not created");
        require(hunter != address(0), "hunter=0");
        require(applications[bountyId][hunter].exists, "no application");
        b.hunter = hunter;
        b.status = Status.Assigned;
        b.assignedAt = uint64(block.timestamp);
        emit HunterAssigned(bountyId, hunter);
    }

    function submitWork(uint256 bountyId, string calldata workURI, bytes32 workHash) external {
        Bounty storage b = bounties[bountyId];
        require(b.hunter == msg.sender, "not hunter");
        require(b.status == Status.Assigned, "not assigned");
        b.status = Status.Submitted;
        b.submittedAt = uint64(block.timestamp);
        emit WorkSubmitted(bountyId, workURI, workHash);
    }

    function approve(uint256 bountyId) external {
        Bounty storage b = bounties[bountyId];
        require(b.sponsor == msg.sender, "not sponsor");
        require(b.status == Status.Submitted, "not submitted");
        b.status = Status.Approved;
        emit Approved(bountyId);
    }

    function payout(uint256 bountyId) external nonReentrant {
        Bounty storage b = bounties[bountyId];
        require(b.sponsor != address(0), "not found");
        require(b.hunter != address(0), "no hunter");
        require(
            b.status == Status.Approved || b.status == Status.ResolvedPaidHunter,
            "not payable"
        );
        require(b.status != Status.PaidOut, "already paid");

        b.status = Status.PaidOut;
        b.payoutAt = uint64(block.timestamp);

        _payoutToken(b.token, b.hunter, b.amount);
        emit PaidOut(bountyId, b.hunter, b.amount);
    }

    function cancelBySponsor(uint256 bountyId) external nonReentrant {
        Bounty storage b = bounties[bountyId];
        require(b.sponsor == msg.sender, "not sponsor");
        require(b.status == Status.Created, "not cancellable");
        b.status = Status.Cancelled;
        _payoutToken(b.token, b.sponsor, b.amount);
        emit Cancelled(bountyId);
        emit Refunded(bountyId, b.sponsor, b.amount);
    }

    function rejectAndDispute(uint256 bountyId) external payable nonReentrant returns (uint256 disputeId) {
        Bounty storage b = bounties[bountyId];
        require(b.sponsor == msg.sender, "not sponsor");
        require(b.status == Status.Submitted, "not submitted");

        uint256 cost = arbitrator.arbitrationCost(arbitratorExtraData);
        require(msg.value >= cost, "fee too low");

        b.status = Status.Disputed;
        disputeId = arbitrator.createDispute{value: cost}(RULING_OPTIONS, arbitratorExtraData);
        disputeToBounty[disputeId] = bountyId;

        emit Disputed(bountyId, disputeId);
        emit Dispute(arbitrator, disputeId, metaEvidenceId, bountyId);

        // refund extra ETH fee if any
        if (msg.value > cost) {
            (bool ok, ) = msg.sender.call{value: msg.value - cost}("");
            require(ok, "refund failed");
        }
    }

    function rule(uint256 _disputeID, uint256 _ruling) external override nonReentrant {
        require(msg.sender == address(arbitrator), "only arbitrator");
        uint256 bountyId = disputeToBounty[_disputeID];
        Bounty storage b = bounties[bountyId];
        require(b.status == Status.Disputed, "not disputed");

        if (_ruling == RULING_HUNTER_WINS) {
            b.status = Status.ResolvedPaidHunter;
        } else if (_ruling == RULING_SPONSOR_WINS) {
            b.status = Status.ResolvedRefundSponsor;
        } else {
            revert("invalid ruling");
        }

        b.resolvedAt = uint64(block.timestamp);
        emit Ruling(arbitrator, _disputeID, _ruling);
    }

    function refund(uint256 bountyId) external nonReentrant {
        Bounty storage b = bounties[bountyId];
        require(b.sponsor != address(0), "not found");
        require(b.status == Status.ResolvedRefundSponsor, "not refundable");
        b.status = Status.Refunded;
        _payoutToken(b.token, b.sponsor, b.amount);
        emit Refunded(bountyId, b.sponsor, b.amount);
    }

    function _createBounty(address sponsor, address token, uint256 amount, string calldata metadataURI, bytes32 metadataHash)
        internal
        returns (uint256 bountyId)
    {
        bountyId = ++bountyCount;
        bounties[bountyId] = Bounty({
            sponsor: sponsor,
            token: token,
            amount: amount,
            hunter: address(0),
            status: Status.Created,
            metadataURI: metadataURI,
            metadataHash: metadataHash,
            createdAt: uint64(block.timestamp),
            assignedAt: 0,
            submittedAt: 0,
            resolvedAt: 0,
            payoutAt: 0
        });
        emit BountyCreated(bountyId, sponsor, token, amount, metadataURI, metadataHash);
    }

    function _payoutToken(address token, address to, uint256 amount) internal {
        if (token == address(0)) {
            (bool ok, ) = to.call{value: amount}("");
            require(ok, "eth transfer failed");
        } else {
            _safeTransfer(token, to, amount);
        }
    }

    function _safeTransfer(address token, address to, uint256 amount) internal {
        (bool ok, bytes memory data) = token.call(abi.encodeWithSelector(IERC20.transfer.selector, to, amount));
        require(ok && (data.length == 0 || abi.decode(data, (bool))), "transfer failed");
    }

    function _safeTransferFrom(address token, address from, address to, uint256 amount) internal {
        (bool ok, bytes memory data) = token.call(
            abi.encodeWithSelector(IERC20.transferFrom.selector, from, to, amount)
        );
        require(ok && (data.length == 0 || abi.decode(data, (bool))), "transferFrom failed");
    }

    receive() external payable {}
}

