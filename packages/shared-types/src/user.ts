import type { ISODateString, UUID } from "./common.js";

// Roles

export type UserRole = "owner" | "admin" | "member" | "billing" | "viewer";

// User

export interface User {
  id: UUID;
  email: string;
  displayName: string;
  avatarUrl?: string;
  role: UserRole;
  createdAt: ISODateString;
  updatedAt: ISODateString;
}

// Organization

export interface Organization {
  id: UUID;
  name: string;
  slug: string;
  logoUrl?: string;
  createdAt: ISODateString;
  updatedAt: ISODateString;
}

export interface OrganizationMember {
  user: User;
  organization: Organization;
  role: UserRole;
  joinedAt: ISODateString;
}

// Session / Auth

export interface AuthSession {
  user: User;
  organization: Organization;
  accessToken: string;
  expiresAt: ISODateString;
}
