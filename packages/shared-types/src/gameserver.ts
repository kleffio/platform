import type { ISODateString, UUID } from "./common.js";

// Region

export type GameServerRegion =
  | "us-east-1"
  | "us-west-2"
  | "eu-west-1"
  | "eu-central-1"
  | "ap-southeast-1"
  | "ap-northeast-1"
  | "ca-central-1"
  | "sa-east-1";

export const REGION_LABELS: Record<GameServerRegion, string> = {
  "us-east-1": "US East (N. Virginia)",
  "us-west-2": "US West (Oregon)",
  "eu-west-1": "EU West (Ireland)",
  "eu-central-1": "EU Central (Frankfurt)",
  "ap-southeast-1": "Asia Pacific (Singapore)",
  "ap-northeast-1": "Asia Pacific (Tokyo)",
  "ca-central-1": "Canada (Central)",
  "sa-east-1": "South America (São Paulo)",
};

// Status

export type GameServerStatus =
  | "provisioning"
  | "starting"
  | "running"
  | "stopping"
  | "stopped"
  | "restarting"
  | "crashed"
  | "error";

// Plan / Tier

export type GameServerPlanTier = "free" | "starter" | "pro" | "business" | "enterprise";

export interface GameServerPlan {
  id: UUID;
  tier: GameServerPlanTier;
  name: string;
  vcpu: number;
  memoryGb: number;
  storageGb: number;
  bandwidthGb: number;
  maxPlayers: number;
  pricePerHour: number;
}

// Resources (live metrics)

export interface GameServerResources {
  cpuPercent: number;
  memoryPercent: number;
  diskPercent: number;
  networkInMbps: number;
  networkOutMbps: number;
}

// Game Server

export interface GameServer {
  id: UUID;
  organizationId: UUID;
  name: string;
  gameType: string;
  region: GameServerRegion;
  status: GameServerStatus;
  plan: GameServerPlan;
  resources?: GameServerResources;
  ipAddress?: string;
  port?: number;
  currentPlayers: number;
  createdAt: ISODateString;
  updatedAt: ISODateString;
  lastStartedAt?: ISODateString;
}

// Deployment

export type DeploymentStatus = "pending" | "in_progress" | "succeeded" | "failed" | "rolled_back";

export interface Deployment {
  id: UUID;
  gameServerId: UUID;
  version: string;
  status: DeploymentStatus;
  initiatedBy: UUID;
  startedAt: ISODateString;
  finishedAt?: ISODateString;
  logs?: string;
}
