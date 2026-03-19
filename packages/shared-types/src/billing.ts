import type { ISODateString, UUID } from "./common.js";
import type { GameServerPlanTier } from "./gameserver.js";

// Billing Plan

export interface BillingPlan {
  id: UUID;
  tier: GameServerPlanTier;
  name: string;
  description: string;
  pricePerMonth: number;
  pricePerYear: number;
  features: string[];
  maxGameServers: number;
  maxTeamMembers: number;
  supportLevel: "community" | "email" | "priority" | "dedicated";
  isPopular?: boolean;
}

// Subscription

export type SubscriptionStatus =
  | "trialing"
  | "active"
  | "past_due"
  | "canceled"
  | "unpaid"
  | "paused";

export type BillingInterval = "monthly" | "yearly";

export interface Subscription {
  id: UUID;
  organizationId: UUID;
  plan: BillingPlan;
  status: SubscriptionStatus;
  interval: BillingInterval;
  currentPeriodStart: ISODateString;
  currentPeriodEnd: ISODateString;
  cancelAtPeriodEnd: boolean;
  trialEnd?: ISODateString;
  createdAt: ISODateString;
  updatedAt: ISODateString;
}

// Invoice

export type InvoiceStatus = "draft" | "open" | "paid" | "void" | "uncollectible";

export interface InvoiceLineItem {
  id: UUID;
  description: string;
  quantity: number;
  unitAmount: number;
  totalAmount: number;
}

export interface Invoice {
  id: UUID;
  organizationId: UUID;
  subscriptionId?: UUID;
  status: InvoiceStatus;
  number: string;
  lines: InvoiceLineItem[];
  subtotal: number;
  tax: number;
  total: number;
  currency: string;
  dueDate?: ISODateString;
  paidAt?: ISODateString;
  createdAt: ISODateString;
}

// Usage

export interface UsageSummary {
  organizationId: UUID;
  periodStart: ISODateString;
  periodEnd: ISODateString;
  serverHours: number;
  bandwidthGb: number;
  storageGb: number;
  estimatedCost: number;
}
