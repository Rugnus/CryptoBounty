// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test} from "forge-std/Test.sol";

import {BountyEscrow} from "../src/BountyEscrow.sol";
import {MockArbitrator} from "../src/mocks/MockArbitrator.sol";
import {MockERC20} from "../src/mocks/MockERC20.sol";

contract BountyEscrowTest is Test {
    address sponsor = address(0xA11CE);
    address hunter = address(0xB0B);

    MockArbitrator arb;
    BountyEscrow escrow;
    MockERC20 usdc;

    function setUp() public {
        arb = new MockArbitrator(0.01 ether);
        escrow = new BountyEscrow(arb, hex"", 0);
        usdc = new MockERC20("USD Coin", "USDC", 6);
        escrow.setTokenWhitelist(address(usdc), true);

        vm.deal(sponsor, 10 ether);
        vm.deal(hunter, 10 ether);
    }

    function testETH_HappyPathApprovePayout() public {
        vm.prank(sponsor);
        uint256 id = escrow.createBountyETH{value: 1 ether}("ipfs://cid", keccak256("m"));

        vm.prank(hunter);
        escrow.applyBounty(id, "ipfs://apply");

        vm.prank(sponsor);
        escrow.assignHunter(id, hunter);

        vm.prank(hunter);
        escrow.submitWork(id, "ipfs://work", keccak256("w"));

        vm.prank(sponsor);
        escrow.approve(id);

        uint256 balBefore = hunter.balance;
        escrow.payout(id);
        assertEq(hunter.balance, balBefore + 1 ether);
    }

    function testERC20_DisputeHunterWins() public {
        usdc.mint(sponsor, 1_000_000);

        vm.startPrank(sponsor);
        usdc.approve(address(escrow), 500_000);
        uint256 id = escrow.createBountyERC20(address(usdc), 500_000, "https://meta", keccak256("m"));
        vm.stopPrank();

        vm.prank(hunter);
        escrow.applyBounty(id, "apply");

        vm.prank(sponsor);
        escrow.assignHunter(id, hunter);

        vm.prank(hunter);
        escrow.submitWork(id, "work", keccak256("w"));

        vm.prank(sponsor);
        uint256 disputeId = escrow.rejectAndDispute{value: 0.01 ether}(id);

        arb.giveRuling(disputeId, escrow.RULING_HUNTER_WINS());

        uint256 balBefore = usdc.balanceOf(hunter);
        escrow.payout(id);
        assertEq(usdc.balanceOf(hunter), balBefore + 500_000);
    }

    function testDisputeSponsorWinsRefund() public {
        vm.prank(sponsor);
        uint256 id = escrow.createBountyETH{value: 1 ether}("ipfs://cid", keccak256("m"));

        vm.prank(hunter);
        escrow.applyBounty(id, "apply");

        vm.prank(sponsor);
        escrow.assignHunter(id, hunter);

        vm.prank(hunter);
        escrow.submitWork(id, "work", keccak256("w"));

        vm.prank(sponsor);
        uint256 disputeId = escrow.rejectAndDispute{value: 0.01 ether}(id);

        arb.giveRuling(disputeId, escrow.RULING_SPONSOR_WINS());

        uint256 balBefore = sponsor.balance;
        escrow.refund(id);
        assertEq(sponsor.balance, balBefore + 1 ether);
    }
}

