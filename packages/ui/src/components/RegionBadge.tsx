import type { GameServerRegion } from "@kleff/shared-types";
import { REGION_LABELS } from "@kleff/shared-types";

interface RegionBadgeProps {
  region: GameServerRegion;
  short?: boolean;
  className?: string;
}

const REGION_FLAGS: Record<GameServerRegion, string> = {
  "us-east-1": "🇺🇸",
  "us-west-2": "🇺🇸",
  "eu-west-1": "🇮🇪",
  "eu-central-1": "🇩🇪",
  "ap-southeast-1": "🇸🇬",
  "ap-northeast-1": "🇯🇵",
  "ca-central-1": "🇨🇦",
  "sa-east-1": "🇧🇷",
};

const REGION_SHORT: Record<GameServerRegion, string> = {
  "us-east-1": "US East",
  "us-west-2": "US West",
  "eu-west-1": "EU West",
  "eu-central-1": "EU Central",
  "ap-southeast-1": "Singapore",
  "ap-northeast-1": "Tokyo",
  "ca-central-1": "Canada",
  "sa-east-1": "São Paulo",
};

export function RegionBadge({ region, short = false, className = "" }: RegionBadgeProps) {
  const label = short ? REGION_SHORT[region] : REGION_LABELS[region];
  const flag = REGION_FLAGS[region];

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium bg-zinc-800 text-zinc-300 ring-1 ring-inset ring-zinc-700/50 ${className}`}
    >
      <span>{flag}</span>
      {label ?? region}
    </span>
  );
}
