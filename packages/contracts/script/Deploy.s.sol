// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Script} from "forge-std/Script.sol";

import {BountyEscrow} from "../src/BountyEscrow.sol";
import {IArbitrator} from "../src/interfaces/IArbitrator.sol";
import {MockArbitrator} from "../src/mocks/MockArbitrator.sol";
import {MockERC20} from "../src/mocks/MockERC20.sol";

contract Deploy is Script {
    function run() external returns (BountyEscrow escrow, address arbitrator, MockERC20 usdc, MockERC20 usdt) {
        uint256 pk = vm.envUint("DEPLOYER_PRIVATE_KEY");
        vm.startBroadcast(pk);

        address arb = vm.envOr("ARBITRATOR_ADDRESS", address(0));
        if (arb == address(0)) {
            arb = address(new MockArbitrator(0.01 ether));
        }
        arbitrator = arb;
        escrow = new BountyEscrow(IArbitrator(arb), hex"", 0);

        usdc = new MockERC20("USD Coin", "USDC", 6);
        usdt = new MockERC20("Tether USD", "USDT", 6);
        escrow.setTokenWhitelist(address(usdc), true);
        escrow.setTokenWhitelist(address(usdt), true);

        vm.stopBroadcast();
    }
}

