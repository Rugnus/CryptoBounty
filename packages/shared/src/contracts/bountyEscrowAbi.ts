export const bountyEscrowAbi = [
  {
    type: "function",
    name: "createBountyETH",
    stateMutability: "payable",
    inputs: [
      { name: "metadataURI", type: "string" },
      { name: "metadataHash", type: "bytes32" }
    ],
    outputs: [{ name: "bountyId", type: "uint256" }]
  },
  {
    type: "function",
    name: "createBountyERC20",
    stateMutability: "nonpayable",
    inputs: [
      { name: "token", type: "address" },
      { name: "amount", type: "uint256" },
      { name: "metadataURI", type: "string" },
      { name: "metadataHash", type: "bytes32" }
    ],
    outputs: [{ name: "bountyId", type: "uint256" }]
  },
  {
    type: "function",
    name: "apply",
    stateMutability: "nonpayable",
    inputs: [
      { name: "bountyId", type: "uint256" },
      { name: "messageURI", type: "string" }
    ],
    outputs: []
  },
  {
    type: "function",
    name: "assignHunter",
    stateMutability: "nonpayable",
    inputs: [
      { name: "bountyId", type: "uint256" },
      { name: "hunter", type: "address" }
    ],
    outputs: []
  },
  {
    type: "function",
    name: "submitWork",
    stateMutability: "nonpayable",
    inputs: [
      { name: "bountyId", type: "uint256" },
      { name: "workURI", type: "string" },
      { name: "workHash", type: "bytes32" }
    ],
    outputs: []
  },
  {
    type: "function",
    name: "approve",
    stateMutability: "nonpayable",
    inputs: [{ name: "bountyId", type: "uint256" }],
    outputs: []
  },
  {
    type: "function",
    name: "payout",
    stateMutability: "nonpayable",
    inputs: [{ name: "bountyId", type: "uint256" }],
    outputs: []
  },
  {
    type: "function",
    name: "cancelBySponsor",
    stateMutability: "nonpayable",
    inputs: [{ name: "bountyId", type: "uint256" }],
    outputs: []
  },
  {
    type: "function",
    name: "rejectAndDispute",
    stateMutability: "payable",
    inputs: [{ name: "bountyId", type: "uint256" }],
    outputs: [{ name: "disputeId", type: "uint256" }]
  },
  {
    type: "function",
    name: "refund",
    stateMutability: "nonpayable",
    inputs: [{ name: "bountyId", type: "uint256" }],
    outputs: []
  },
  {
    type: "function",
    name: "bountyCount",
    stateMutability: "view",
    inputs: [],
    outputs: [{ type: "uint256" }]
  },
  {
    type: "function",
    name: "bounties",
    stateMutability: "view",
    inputs: [{ name: "", type: "uint256" }],
    outputs: [
      { name: "sponsor", type: "address" },
      { name: "token", type: "address" },
      { name: "amount", type: "uint256" },
      { name: "hunter", type: "address" },
      { name: "status", type: "uint8" },
      { name: "metadataURI", type: "string" },
      { name: "metadataHash", type: "bytes32" },
      { name: "createdAt", type: "uint64" },
      { name: "assignedAt", type: "uint64" },
      { name: "submittedAt", type: "uint64" },
      { name: "resolvedAt", type: "uint64" },
      { name: "payoutAt", type: "uint64" }
    ]
  },
  {
    type: "event",
    name: "BountyCreated",
    inputs: [
      { name: "bountyId", type: "uint256", indexed: true },
      { name: "sponsor", type: "address", indexed: true },
      { name: "token", type: "address", indexed: true },
      { name: "amount", type: "uint256", indexed: false },
      { name: "metadataURI", type: "string", indexed: false },
      { name: "metadataHash", type: "bytes32", indexed: false }
    ],
    anonymous: false
  },
  {
    type: "event",
    name: "ApplicationSubmitted",
    inputs: [
      { name: "bountyId", type: "uint256", indexed: true },
      { name: "hunter", type: "address", indexed: true },
      { name: "messageURI", type: "string", indexed: false }
    ],
    anonymous: false
  },
  {
    type: "event",
    name: "HunterAssigned",
    inputs: [
      { name: "bountyId", type: "uint256", indexed: true },
      { name: "hunter", type: "address", indexed: true }
    ],
    anonymous: false
  },
  {
    type: "event",
    name: "WorkSubmitted",
    inputs: [
      { name: "bountyId", type: "uint256", indexed: true },
      { name: "workURI", type: "string", indexed: false },
      { name: "workHash", type: "bytes32", indexed: false }
    ],
    anonymous: false
  },
  {
    type: "event",
    name: "Approved",
    inputs: [{ name: "bountyId", type: "uint256", indexed: true }],
    anonymous: false
  },
  {
    type: "event",
    name: "Disputed",
    inputs: [
      { name: "bountyId", type: "uint256", indexed: true },
      { name: "disputeId", type: "uint256", indexed: true }
    ],
    anonymous: false
  },
  {
    type: "event",
    name: "PaidOut",
    inputs: [
      { name: "bountyId", type: "uint256", indexed: true },
      { name: "to", type: "address", indexed: true },
      { name: "amount", type: "uint256", indexed: false }
    ],
    anonymous: false
  },
  {
    type: "event",
    name: "Refunded",
    inputs: [
      { name: "bountyId", type: "uint256", indexed: true },
      { name: "to", type: "address", indexed: true },
      { name: "amount", type: "uint256", indexed: false }
    ],
    anonymous: false
  }
] as const;

