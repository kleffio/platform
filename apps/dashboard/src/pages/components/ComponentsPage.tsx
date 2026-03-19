import { useState } from "react";
import type { GameServer, Invoice, Subscription } from "@kleff/shared-types";
import {
    // Domain components
    MetricCard,
    PlanBadge,
    RegionBadge,
    ServerCard,
    StatusBadge,
    // Primitives
    Alert,
    AlertDescription,
    AlertTitle,
    Badge,
    Button,
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    ChartContainer,
    ChartTooltip,
    ChartTooltipContent,
    Separator,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
    Toaster,
} from "@kleff/ui";
import { ResponsiveContainer, BarChart, Bar, XAxis, CartesianGrid } from "recharts";
import { Activity, AlertTriangle, Server, Users } from "lucide-react";
import { toast } from "sonner";

// ─── Mock data ────────────────────────────────────────────────────────────────

const MOCK_SERVERS: GameServer[] = [
    {
        id: "srv-001",
        organizationId: "org-001",
        name: "mc-survival-na",
        gameType: "Minecraft",
        region: "us-east-1",
        status: "running",
        plan: { id: "plan-pro", tier: "pro", name: "Pro", vcpu: 4, memoryGb: 8, storageGb: 100, bandwidthGb: 2000, maxPlayers: 50, pricePerHour: 0.12 },
        resources: { cpuPercent: 42, memoryPercent: 67, diskPercent: 23, networkInMbps: 12, networkOutMbps: 8 },
        ipAddress: "34.201.12.88",
        port: 25565,
        currentPlayers: 18,
        createdAt: "2024-11-01T00:00:00Z",
        updatedAt: "2025-03-18T10:00:00Z",
        lastStartedAt: "2025-03-18T08:00:00Z",
    },
    {
        id: "srv-002",
        organizationId: "org-001",
        name: "valheim-eu-pvp",
        gameType: "Valheim",
        region: "eu-central-1",
        status: "running",
        plan: { id: "plan-business", tier: "business", name: "Business", vcpu: 8, memoryGb: 16, storageGb: 250, bandwidthGb: 5000, maxPlayers: 100, pricePerHour: 0.28 },
        resources: { cpuPercent: 71, memoryPercent: 55, diskPercent: 40, networkInMbps: 45, networkOutMbps: 22 },
        ipAddress: "18.185.44.9",
        port: 2456,
        currentPlayers: 63,
        createdAt: "2024-12-15T00:00:00Z",
        updatedAt: "2025-03-18T10:00:00Z",
        lastStartedAt: "2025-03-17T20:00:00Z",
    },
    {
        id: "srv-003",
        organizationId: "org-001",
        name: "csgo-casual-ap",
        gameType: "CS2",
        region: "ap-southeast-1",
        status: "stopped",
        plan: { id: "plan-starter", tier: "starter", name: "Starter", vcpu: 2, memoryGb: 4, storageGb: 50, bandwidthGb: 500, maxPlayers: 20, pricePerHour: 0.05 },
        currentPlayers: 0,
        createdAt: "2025-01-10T00:00:00Z",
        updatedAt: "2025-03-17T22:00:00Z",
    },
    {
        id: "srv-004",
        organizationId: "org-001",
        name: "ark-rag-us",
        gameType: "ARK: Survival Evolved",
        region: "us-west-2",
        status: "provisioning",
        plan: { id: "plan-pro", tier: "pro", name: "Pro", vcpu: 4, memoryGb: 8, storageGb: 100, bandwidthGb: 2000, maxPlayers: 50, pricePerHour: 0.12 },
        currentPlayers: 0,
        createdAt: "2025-03-18T09:45:00Z",
        updatedAt: "2025-03-18T09:45:00Z",
    },
    {
        id: "srv-005",
        organizationId: "org-001",
        name: "rust-official-eu",
        gameType: "Rust",
        region: "eu-west-1",
        status: "crashed",
        plan: { id: "plan-business", tier: "business", name: "Business", vcpu: 8, memoryGb: 16, storageGb: 250, bandwidthGb: 5000, maxPlayers: 100, pricePerHour: 0.28 },
        resources: { cpuPercent: 0, memoryPercent: 0, diskPercent: 61, networkInMbps: 0, networkOutMbps: 0 },
        currentPlayers: 0,
        createdAt: "2025-02-01T00:00:00Z",
        updatedAt: "2025-03-18T07:12:00Z",
        lastStartedAt: "2025-03-18T06:00:00Z",
    },
    {
        id: "srv-006",
        organizationId: "org-001",
        name: "terraria-jp",
        gameType: "Terraria",
        region: "ap-northeast-1",
        status: "running",
        plan: { id: "plan-free", tier: "free", name: "Free", vcpu: 1, memoryGb: 1, storageGb: 10, bandwidthGb: 100, maxPlayers: 8, pricePerHour: 0 },
        resources: { cpuPercent: 8, memoryPercent: 31, diskPercent: 5, networkInMbps: 1, networkOutMbps: 1 },
        currentPlayers: 3,
        createdAt: "2025-03-01T00:00:00Z",
        updatedAt: "2025-03-18T10:00:00Z",
        lastStartedAt: "2025-03-18T09:00:00Z",
    },
];

const MOCK_SUBSCRIPTION: Subscription = {
    id: "sub-001",
    organizationId: "org-001",
    plan: {
        id: "plan-business",
        tier: "business",
        name: "Business",
        description: "For serious hosting operations",
        pricePerMonth: 149,
        pricePerYear: 1490,
        features: ["Up to 20 game servers", "Priority support", "Custom domains", "Advanced analytics", "DDoS protection"],
        maxGameServers: 20,
        maxTeamMembers: 15,
        supportLevel: "priority",
        isPopular: true,
    },
    status: "active",
    interval: "monthly",
    currentPeriodStart: "2025-03-01T00:00:00Z",
    currentPeriodEnd: "2025-04-01T00:00:00Z",
    cancelAtPeriodEnd: false,
    createdAt: "2024-11-01T00:00:00Z",
    updatedAt: "2025-03-01T00:00:00Z",
};

const MOCK_INVOICES: Invoice[] = [
    {
        id: "inv-003", organizationId: "org-001", subscriptionId: "sub-001",
        status: "paid", number: "INV-2025-003",
        lines: [{ id: "li-1", description: "Business Plan — March 2025", quantity: 1, unitAmount: 14900, totalAmount: 14900 }],
        subtotal: 14900, tax: 1937, total: 16837, currency: "usd",
        paidAt: "2025-03-01T10:00:00Z", createdAt: "2025-03-01T00:00:00Z",
    },
    {
        id: "inv-002", organizationId: "org-001", subscriptionId: "sub-001",
        status: "paid", number: "INV-2025-002",
        lines: [{ id: "li-2", description: "Business Plan — February 2025", quantity: 1, unitAmount: 14900, totalAmount: 14900 }],
        subtotal: 14900, tax: 1937, total: 16837, currency: "usd",
        paidAt: "2025-02-01T10:00:00Z", createdAt: "2025-02-01T00:00:00Z",
    },
    {
        id: "inv-001", organizationId: "org-001", subscriptionId: "sub-001",
        status: "open", number: "INV-2025-001",
        lines: [{ id: "li-3", description: "Business Plan — January 2025", quantity: 1, unitAmount: 14900, totalAmount: 14900 }],
        subtotal: 14900, tax: 1937, total: 16837, currency: "usd",
        createdAt: "2025-01-01T00:00:00Z",
    },
];

const deploymentChartData = [
    { day: "Mon", deployments: 4 },
    { day: "Tue", deployments: 7 },
    { day: "Wed", deployments: 3 },
    { day: "Thu", deployments: 9 },
    { day: "Fri", deployments: 6 },
    { day: "Sat", deployments: 2 },
    { day: "Sun", deployments: 1 },
];

const chartConfig = { deployments: { label: "Deployments" } };

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatCents(cents: number, currency = "usd") {
    return new Intl.NumberFormat("en-US", { style: "currency", currency: currency.toUpperCase() }).format(cents / 100);
}

function formatDate(iso: string) {
    return new Date(iso).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function ComponentsPage() {
    const [servers, setServers] = useState<GameServer[]>(MOCK_SERVERS);

    const runningCount = servers.filter((s) => s.status === "running").length;
    const crashedServers = servers.filter((s) => s.status === "crashed" || s.status === "error");
    const totalPlayers = servers.reduce((sum, s) => sum + s.currentPlayers, 0);

    function handleStart(id: string) {
        setServers((prev) =>
            prev.map((s) => (s.id === id ? { ...s, status: "starting" as const, updatedAt: new Date().toISOString() } : s))
        );
        toast.success("Server starting", { description: `Starting ${servers.find((s) => s.id === id)?.name}…` });
    }

    function handleStop(id: string) {
        setServers((prev) =>
            prev.map((s) => (s.id === id ? { ...s, status: "stopping" as const, currentPlayers: 0, updatedAt: new Date().toISOString() } : s))
        );
        toast("Server stopping", { description: `Gracefully stopping ${servers.find((s) => s.id === id)?.name}…` });
    }

    return (
        <div className="min-h-screen bg-zinc-950 text-zinc-50 p-6 md:p-10">
            <div className="mx-auto max-w-7xl space-y-8">

                {/* ── Header ── */}
                <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
                    <div>
                        <Badge className="mb-3">Game Hosting Platform</Badge>
                        <h1 className="text-3xl font-semibold tracking-tight">Dashboard</h1>
                        <p className="mt-1 text-zinc-400">Manage your hosted game servers across all regions.</p>
                    </div>
                    <Button onClick={() => toast.success("Coming soon", { description: "Server provisioning UI is on the roadmap." })}>
                        <Server className="mr-2 h-4 w-4" />
                        New server
                    </Button>
                </div>

                {/* ── Crashed server alert ── */}
                {crashedServers.length > 0 && (
                    <Alert variant="destructive">
                        <AlertTriangle className="h-4 w-4" />
                        <AlertTitle>{crashedServers.length} server{crashedServers.length > 1 ? "s" : ""} need attention</AlertTitle>
                        <AlertDescription>
                            {crashedServers.map((s) => s.name).join(", ")} {crashedServers.length > 1 ? "have" : "has"} crashed and may require a restart.
                        </AlertDescription>
                    </Alert>
                )}

                {/* ── Metric cards ── */}
                <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                    <MetricCard label="Total Servers" value={servers.length} icon={<Server className="h-4 w-4" />} />
                    <MetricCard
                        label="Running"
                        value={runningCount}
                        icon={<Activity className="h-4 w-4" />}
                        delta={{ value: `${servers.length - runningCount} offline`, direction: servers.length - runningCount > 0 ? "down" : "neutral" }}
                    />
                    <MetricCard
                        label="Players Online"
                        value={totalPlayers}
                        icon={<Users className="h-4 w-4" />}
                        delta={{ value: "+12 vs yesterday", direction: "up" }}
                    />
                    <MetricCard
                        label="Est. Monthly Cost"
                        value={formatCents(MOCK_SUBSCRIPTION.plan.pricePerMonth * 100)}
                        delta={{ value: "Business plan", direction: "neutral" }}
                    />
                </div>

                {/* ── Tabs ── */}
                <Tabs defaultValue="overview" className="space-y-6">
                    <TabsList className="grid w-full grid-cols-4">
                        <TabsTrigger value="overview">Overview</TabsTrigger>
                        <TabsTrigger value="servers">Servers</TabsTrigger>
                        <TabsTrigger value="billing">Billing</TabsTrigger>
                        <TabsTrigger value="components">Components</TabsTrigger>
                    </TabsList>

                    {/* ── Overview ── */}
                    <TabsContent value="overview" className="space-y-6">
                        <div className="grid gap-6 xl:grid-cols-3">
                            <Card className="xl:col-span-2 rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Deployments this week</CardTitle>
                                    <CardDescription>Server deploys and restarts per day.</CardDescription>
                                </CardHeader>
                                <CardContent>
                                    <ChartContainer config={chartConfig} className="h-60 w-full">
                                        <ResponsiveContainer width="100%" height="100%">
                                            <BarChart data={deploymentChartData}>
                                                <CartesianGrid vertical={false} />
                                                <XAxis dataKey="day" tickLine={false} axisLine={false} />
                                                <ChartTooltip content={<ChartTooltipContent />} />
                                                <Bar dataKey="deployments" radius={6} fill="#f5b517" />
                                            </BarChart>
                                        </ResponsiveContainer>
                                    </ChartContainer>
                                </CardContent>
                            </Card>

                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Service health</CardTitle>
                                    <CardDescription>All servers at a glance.</CardDescription>
                                </CardHeader>
                                <CardContent>
                                    <div className="space-y-3">
                                        {servers.map((server) => (
                                            <div key={server.id} className="flex items-center justify-between gap-2">
                                                <div className="min-w-0">
                                                    <p className="text-sm font-medium text-zinc-200 truncate">{server.name}</p>
                                                    <RegionBadge region={server.region} short className="mt-0.5" />
                                                </div>
                                                <StatusBadge status={server.status} showDot />
                                            </div>
                                        ))}
                                    </div>
                                </CardContent>
                            </Card>
                        </div>

                        <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                            <CardHeader>
                                <CardTitle>All servers</CardTitle>
                                <CardDescription>Quick reference — switch to the Servers tab to manage.</CardDescription>
                            </CardHeader>
                            <CardContent>
                                <Table>
                                    <TableHeader>
                                        <TableRow>
                                            <TableHead>Name</TableHead>
                                            <TableHead>Game</TableHead>
                                            <TableHead>Region</TableHead>
                                            <TableHead>Plan</TableHead>
                                            <TableHead>Status</TableHead>
                                            <TableHead className="text-right">Players</TableHead>
                                        </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                        {servers.map((server) => (
                                            <TableRow key={server.id}>
                                                <TableCell className="font-mono text-sm">{server.name}</TableCell>
                                                <TableCell>{server.gameType}</TableCell>
                                                <TableCell><RegionBadge region={server.region} short /></TableCell>
                                                <TableCell><PlanBadge tier={server.plan.tier} /></TableCell>
                                                <TableCell><StatusBadge status={server.status} /></TableCell>
                                                <TableCell className="text-right tabular-nums">
                                                    {server.currentPlayers} / {server.plan.maxPlayers}
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                            </CardContent>
                        </Card>
                    </TabsContent>

                    {/* ── Servers ── */}
                    <TabsContent value="servers" className="space-y-4">
                        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
                            {servers.map((server) => (
                                <ServerCard key={server.id} server={server} onStart={handleStart} onStop={handleStop} />
                            ))}
                        </div>
                    </TabsContent>

                    {/* ── Billing ── */}
                    <TabsContent value="billing" className="space-y-6">
                        <div className="grid gap-6 lg:grid-cols-2">
                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader className="flex flex-row items-start justify-between">
                                    <div>
                                        <CardTitle>Current subscription</CardTitle>
                                        <CardDescription>Your active plan and billing details.</CardDescription>
                                    </div>
                                    <PlanBadge tier={MOCK_SUBSCRIPTION.plan.tier} />
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    <div className="flex items-baseline gap-1">
                                        <span className="text-3xl font-semibold">{formatCents(MOCK_SUBSCRIPTION.plan.pricePerMonth * 100)}</span>
                                        <span className="text-zinc-400">/ month</span>
                                    </div>
                                    <Separator className="border-zinc-800" />
                                    <ul className="space-y-2">
                                        {MOCK_SUBSCRIPTION.plan.features.map((feature) => (
                                            <li key={feature} className="flex items-center gap-2 text-sm text-zinc-300">
                                                <span className="text-amber-400">✓</span>
                                                {feature}
                                            </li>
                                        ))}
                                    </ul>
                                    <Separator className="border-zinc-800" />
                                    <div className="grid grid-cols-2 gap-2 text-sm">
                                        <div><p className="text-zinc-500">Period start</p><p className="font-medium">{formatDate(MOCK_SUBSCRIPTION.currentPeriodStart)}</p></div>
                                        <div><p className="text-zinc-500">Next billing</p><p className="font-medium">{formatDate(MOCK_SUBSCRIPTION.currentPeriodEnd)}</p></div>
                                        <div><p className="text-zinc-500">Servers</p><p className="font-medium">{servers.length} / {MOCK_SUBSCRIPTION.plan.maxGameServers}</p></div>
                                        <div><p className="text-zinc-500">Support</p><p className="font-medium capitalize">{MOCK_SUBSCRIPTION.plan.supportLevel}</p></div>
                                    </div>
                                    <Button variant="outline" className="w-full" onClick={() => toast("Coming soon", { description: "Plan management is on the roadmap." })}>
                                        Manage plan
                                    </Button>
                                </CardContent>
                            </Card>

                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Invoice history</CardTitle>
                                    <CardDescription>Your recent billing statements.</CardDescription>
                                </CardHeader>
                                <CardContent>
                                    <Table>
                                        <TableHeader>
                                            <TableRow>
                                                <TableHead>Invoice</TableHead>
                                                <TableHead>Date</TableHead>
                                                <TableHead>Status</TableHead>
                                                <TableHead className="text-right">Total</TableHead>
                                            </TableRow>
                                        </TableHeader>
                                        <TableBody>
                                            {MOCK_INVOICES.map((invoice) => (
                                                <TableRow key={invoice.id}>
                                                    <TableCell className="font-mono text-xs">{invoice.number}</TableCell>
                                                    <TableCell>{formatDate(invoice.createdAt)}</TableCell>
                                                    <TableCell>
                                                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${
                                                            invoice.status === "paid" ? "bg-emerald-500/10 text-emerald-400 ring-emerald-500/20"
                                                            : invoice.status === "open" ? "bg-amber-400/10 text-amber-400 ring-amber-400/20"
                                                            : "bg-zinc-500/10 text-zinc-400 ring-zinc-500/20"
                                                        }`}>
                                                            {invoice.status.charAt(0).toUpperCase() + invoice.status.slice(1)}
                                                        </span>
                                                    </TableCell>
                                                    <TableCell className="text-right tabular-nums">{formatCents(invoice.total, invoice.currency)}</TableCell>
                                                </TableRow>
                                            ))}
                                        </TableBody>
                                    </Table>
                                </CardContent>
                            </Card>
                        </div>
                    </TabsContent>

                    {/* ── Components showcase ── */}
                    <TabsContent value="components" className="space-y-8">

                        {/* StatusBadge — all 8 statuses */}
                        <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                            <CardHeader>
                                <CardTitle>StatusBadge</CardTitle>
                                <CardDescription>All GameServerStatus variants. Transitional states pulse.</CardDescription>
                            </CardHeader>
                            <CardContent className="flex flex-wrap gap-3">
                                <StatusBadge status="running" />
                                <StatusBadge status="stopped" />
                                <StatusBadge status="starting" />
                                <StatusBadge status="stopping" />
                                <StatusBadge status="restarting" />
                                <StatusBadge status="provisioning" />
                                <StatusBadge status="crashed" />
                                <StatusBadge status="error" />
                            </CardContent>
                        </Card>

                        {/* PlanBadge — all 5 tiers */}
                        <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                            <CardHeader>
                                <CardTitle>PlanBadge</CardTitle>
                                <CardDescription>All GameServerPlanTier variants.</CardDescription>
                            </CardHeader>
                            <CardContent className="flex flex-wrap gap-3">
                                <PlanBadge tier="free" />
                                <PlanBadge tier="starter" />
                                <PlanBadge tier="pro" />
                                <PlanBadge tier="business" />
                                <PlanBadge tier="enterprise" />
                            </CardContent>
                        </Card>

                        {/* RegionBadge — all 8 regions, both modes */}
                        <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                            <CardHeader>
                                <CardTitle>RegionBadge</CardTitle>
                                <CardDescription>All 8 regions — full label and short mode.</CardDescription>
                            </CardHeader>
                            <CardContent className="space-y-3">
                                <div className="flex flex-wrap gap-2">
                                    <RegionBadge region="us-east-1" />
                                    <RegionBadge region="us-west-2" />
                                    <RegionBadge region="eu-west-1" />
                                    <RegionBadge region="eu-central-1" />
                                    <RegionBadge region="ap-southeast-1" />
                                    <RegionBadge region="ap-northeast-1" />
                                    <RegionBadge region="ca-central-1" />
                                    <RegionBadge region="sa-east-1" />
                                </div>
                                <p className="text-xs text-zinc-500">Short mode:</p>
                                <div className="flex flex-wrap gap-2">
                                    <RegionBadge region="us-east-1" short />
                                    <RegionBadge region="eu-central-1" short />
                                    <RegionBadge region="ap-northeast-1" short />
                                    <RegionBadge region="sa-east-1" short />
                                </div>
                            </CardContent>
                        </Card>

                        {/* MetricCard — all delta directions */}
                        <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                            <CardHeader>
                                <CardTitle>MetricCard</CardTitle>
                                <CardDescription>With icon, delta up/down/neutral, and plain.</CardDescription>
                            </CardHeader>
                            <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                                <MetricCard label="Running Servers" value={4} icon={<Server className="h-4 w-4" />} delta={{ value: "+1 today", direction: "up" }} />
                                <MetricCard label="Crashed" value={1} icon={<AlertTriangle className="h-4 w-4" />} delta={{ value: "1 unresolved", direction: "down" }} />
                                <MetricCard label="Players Online" value={84} icon={<Users className="h-4 w-4" />} delta={{ value: "same as yesterday", direction: "neutral" }} />
                                <MetricCard label="Monthly Cost" value="$149.00" />
                            </CardContent>
                        </Card>

                        {/* ServerCard — representative states */}
                        <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                            <CardHeader>
                                <CardTitle>ServerCard</CardTitle>
                                <CardDescription>Running with resources, stopped, provisioning, and crashed.</CardDescription>
                            </CardHeader>
                            <CardContent className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                                <ServerCard
                                    server={MOCK_SERVERS[0]}
                                    onStart={handleStart}
                                    onStop={handleStop}
                                />
                                <ServerCard
                                    server={MOCK_SERVERS[2]}
                                    onStart={handleStart}
                                    onStop={handleStop}
                                />
                                <ServerCard
                                    server={MOCK_SERVERS[3]}
                                    onStart={handleStart}
                                    onStop={handleStop}
                                />
                                <ServerCard
                                    server={MOCK_SERVERS[4]}
                                    onStart={handleStart}
                                    onStop={handleStop}
                                />
                            </CardContent>
                        </Card>

                    </TabsContent>
                </Tabs>
            </div>

            <Toaster />
        </div>
    );
}
