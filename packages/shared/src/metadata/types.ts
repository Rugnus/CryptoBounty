export type BountyCategory =
  | "bugfix"
  | "feature"
  | "audit"
  | "design"
  | "content"
  | "other";

export type BountyDifficulty = "easy" | "medium" | "hard";

export type BountyMetadata = {
  title: string;
  description: string;
  category: BountyCategory;
  tags: string[];
  difficulty: BountyDifficulty;
  payout: {
    tokenSymbol: string;
    amount: string;
  };
  chainId: number;
  createdAt: string;
  externalUrl?: string;
  attachments?: { name: string; url: string }[];
};

