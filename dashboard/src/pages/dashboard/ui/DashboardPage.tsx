import { useMemo, useState } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/shared/ui/alert";
import { Avatar, AvatarFallback, AvatarImage } from "@/shared/ui/avatar";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { ButtonGroup } from "@/shared/ui/button-group";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import {
    Carousel,
    CarouselContent,
    CarouselItem,
    CarouselNext,
    CarouselPrevious,
} from "@/shared/ui/carousel";
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/shared/ui/chart";
import { Checkbox } from "@/shared/ui/checkbox";
import {
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
} from "@/shared/ui/collapsible";
import {
    ContextMenu,
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuTrigger,
} from "@/shared/ui/context-menu";
import {
    Drawer,
    DrawerContent,
    DrawerDescription,
    DrawerFooter,
    DrawerHeader,
    DrawerTitle,
    DrawerTrigger,
} from "@/shared/ui/drawer";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu";
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/shared/ui/empty";
import { Field, FieldContent, FieldDescription, FieldLabel } from "@/shared/ui/field";
import { HoverCard, HoverCardContent, HoverCardTrigger } from "@/shared/ui/hover-card";
import { Input } from "@/shared/ui/input";
import { Item, ItemDescription, ItemTitle } from "@/shared/ui/item";
import { Label } from "@/shared/ui/label";
import {
    Menubar,
    MenubarContent,
    MenubarItem,
    MenubarMenu,
    MenubarTrigger,
} from "@/shared/ui/menubar";
import {
    Pagination,
    PaginationContent,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious,
} from "@/shared/ui/pagination";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/shared/ui/popover";
import { Progress } from "@/shared/ui/progress";
import { RadioGroup, RadioGroupItem } from "@/shared/ui/radio-group";
import { ScrollArea } from "@/shared/ui/scroll-area";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { Separator } from "@/shared/ui/separator";
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
    SheetTrigger,
} from "@/shared/ui/sheet";
import { Skeleton } from "@/shared/ui/skeleton";
import { Slider } from "@/shared/ui/slider";
import { toast } from "sonner";
import { Toaster } from "@/shared/ui/sonner";
import { Switch } from "@/shared/ui/switch";
import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
} from "@/shared/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";
import { ResponsiveContainer, BarChart, Bar, XAxis, CartesianGrid } from "recharts";
import { Bell, ChevronDown, Database, Server, Shield, Sparkles } from "lucide-react";

const chartData = [
    { name: "Mon", deployments: 12 },
    { name: "Tue", deployments: 18 },
    { name: "Wed", deployments: 9 },
    { name: "Thu", deployments: 22 },
    { name: "Fri", deployments: 16 },
];

const chartConfig = {
    deployments: {
        label: "Deployments",
    },
};

export default function DashboardPage() {
    const [progress, setProgress] = useState(64);
    const [cpuLimit, setCpuLimit] = useState([45]);
    const [emailEnabled, setEmailEnabled] = useState(true);
    const [agreed, setAgreed] = useState(true);
    const stats = useMemo(
        () => [
            { title: "Active Nodes", value: "12", icon: Server },
            { title: "Protected Apps", value: "38", icon: Shield },
            { title: "Databases", value: "24", icon: Database },
        ],
        []
    );

    return (
        <div className="min-h-screen bg-zinc-950 text-zinc-50 p-6 md:p-10">
            <div className="mx-auto max-w-7xl space-y-8">
                <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
                    <div>
                        <Badge className="mb-3">Kleff UI Showcase</Badge>
                        <h1 className="text-4xl font-semibold tracking-tight">Premium dashboard component preview</h1>
                        <p className="mt-2 max-w-3xl text-zinc-400">
                            A quick all-in-one page using your generated shadcn components so you can verify styling,
                            imports, and overall dashboard feel.
                        </p>
                    </div>

                    <div className="flex flex-wrap gap-3">
                        <Button variant="outline">Secondary action</Button>
                        <Button onClick={() => toast("Deployment queued", { description: "Your app is being prepared." })}>
                            Trigger toast
                        </Button>
                    </div>
                </div>

                <Menubar className="bg-zinc-900 border-zinc-800">
                    <MenubarMenu>
                        <MenubarTrigger>Platform</MenubarTrigger>
                        <MenubarContent>
                            <MenubarItem>Overview</MenubarItem>
                            <MenubarItem>Projects</MenubarItem>
                        </MenubarContent>
                    </MenubarMenu>
                    <MenubarMenu>
                        <MenubarTrigger>Infrastructure</MenubarTrigger>
                        <MenubarContent>
                            <MenubarItem>Nodes</MenubarItem>
                            <MenubarItem>Regions</MenubarItem>
                        </MenubarContent>
                    </MenubarMenu>
                    <MenubarMenu>
                        <MenubarTrigger>Billing</MenubarTrigger>
                        <MenubarContent>
                            <MenubarItem>Invoices</MenubarItem>
                            <MenubarItem>Usage</MenubarItem>
                        </MenubarContent>
                    </MenubarMenu>
                </Menubar>

                <Alert>
                    <Sparkles className="h-4 w-4" />
                    <AlertTitle>Kleff premium theme ready</AlertTitle>
                    <AlertDescription>
                        This page is only a component playground, but it already mirrors a premium dark dashboard layout.
                    </AlertDescription>
                </Alert>

                <div className="grid gap-6 lg:grid-cols-3">
                    {stats.map((stat) => {
                        const Icon = stat.icon;
                        return (
                            <Card key={stat.title} className="bg-zinc-900/80 border-zinc-800 rounded-2xl">
                                <CardHeader className="flex flex-row items-center justify-between space-y-0">
                                    <div>
                                        <CardTitle className="text-sm text-zinc-400">{stat.title}</CardTitle>
                                        <CardDescription>Cluster snapshot</CardDescription>
                                    </div>
                                    <div className="rounded-xl bg-zinc-800 p-2">
                                        <Icon className="h-4 w-4 text-amber-400" />
                                    </div>
                                </CardHeader>
                                <CardContent>
                                    <p className="text-3xl font-semibold">{stat.value}</p>
                                </CardContent>
                            </Card>
                        );
                    })}
                </div>

                <Tabs defaultValue="inputs" className="space-y-6">
                    <TabsList className="grid w-full grid-cols-4">
                        <TabsTrigger value="inputs">Inputs</TabsTrigger>
                        <TabsTrigger value="navigation">Navigation</TabsTrigger>
                        <TabsTrigger value="display">Display</TabsTrigger>
                        <TabsTrigger value="feedback">Feedback</TabsTrigger>
                    </TabsList>

                    <TabsContent value="inputs" className="space-y-6">
                        <div className="grid gap-6 lg:grid-cols-2">
                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Form controls</CardTitle>
                                    <CardDescription>Core input components in one place.</CardDescription>
                                </CardHeader>
                                <CardContent className="space-y-5">
                                    <Field>
                                        <FieldLabel>Project name</FieldLabel>
                                        <FieldContent>
                                            <Input defaultValue="kleff-platform" />
                                        </FieldContent>
                                        <FieldDescription>This input uses your generated input component.</FieldDescription>
                                    </Field>

                                    <div className="grid gap-4 md:grid-cols-2">
                                        <div className="space-y-2">
                                            <Label>Region</Label>
                                            <Select defaultValue="ca-east">
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select region" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    <SelectItem value="ca-east">Canada East</SelectItem>
                                                    <SelectItem value="us-east">US East</SelectItem>
                                                    <SelectItem value="eu-west">EU West</SelectItem>
                                                </SelectContent>
                                            </Select>
                                        </div>

                                        <div className="space-y-2">
                                            <Label>Plan</Label>
                                            <RadioGroup defaultValue="premium" className="space-y-2">
                                                <div className="flex items-center gap-2">
                                                    <RadioGroupItem value="starter" id="starter" />
                                                    <Label htmlFor="starter">Starter</Label>
                                                </div>
                                                <div className="flex items-center gap-2">
                                                    <RadioGroupItem value="premium" id="premium" />
                                                    <Label htmlFor="premium">Premium</Label>
                                                </div>
                                            </RadioGroup>
                                        </div>
                                    </div>

                                    <div className="space-y-3">
                                        <div className="flex items-center justify-between">
                                            <Label>CPU limit</Label>
                                            <Badge variant="outline">{cpuLimit[0]}%</Badge>
                                        </div>
                                        <Slider value={cpuLimit} onValueChange={setCpuLimit} max={100} step={1} />
                                    </div>

                                    <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between rounded-xl border border-zinc-800 p-4">
                                        <div className="space-y-1">
                                            <p className="font-medium">Email alerts</p>
                                            <p className="text-sm text-zinc-400">Notify owners when deployments fail.</p>
                                        </div>
                                        <Switch checked={emailEnabled} onCheckedChange={setEmailEnabled} />
                                    </div>

                                    <div className="flex items-center gap-3">
                                        <Checkbox checked={agreed} onCheckedChange={(v) => setAgreed(Boolean(v))} />
                                        <span className="text-sm text-zinc-300">Apply configuration immediately after save.</span>
                                    </div>
                                </CardContent>
                            </Card>

                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Interactive overlays</CardTitle>
                                    <CardDescription>Popover, hover card, dropdown, drawer, sheet and context menu.</CardDescription>
                                </CardHeader>
                                <CardContent className="flex flex-wrap gap-3">
                                    <Popover>
                                        <PopoverTrigger asChild>
                                            <Button variant="outline">Open popover</Button>
                                        </PopoverTrigger>
                                        <PopoverContent className="w-72">
                                            Quick actions and contextual controls can live here.
                                        </PopoverContent>
                                    </Popover>

                                    <HoverCard>
                                        <HoverCardTrigger asChild>
                                            <Button variant="outline">Hover for details</Button>
                                        </HoverCardTrigger>
                                        <HoverCardContent>
                                            Hover cards are useful for lightweight summaries and metadata previews.
                                        </HoverCardContent>
                                    </HoverCard>

                                    <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                            <Button variant="outline">
                                                Menu <ChevronDown className="ml-2 h-4 w-4" />
                                            </Button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent>
                                            <DropdownMenuLabel>Deployment</DropdownMenuLabel>
                                            <DropdownMenuSeparator />
                                            <DropdownMenuItem>Redeploy</DropdownMenuItem>
                                            <DropdownMenuItem>Pause</DropdownMenuItem>
                                            <DropdownMenuItem>Delete</DropdownMenuItem>
                                        </DropdownMenuContent>
                                    </DropdownMenu>

                                    <Drawer>
                                        <DrawerTrigger asChild>
                                            <Button variant="outline">Open drawer</Button>
                                        </DrawerTrigger>
                                        <DrawerContent>
                                            <DrawerHeader>
                                                <DrawerTitle>Mobile style drawer</DrawerTitle>
                                                <DrawerDescription>Useful for compact workflows and settings panels.</DrawerDescription>
                                            </DrawerHeader>
                                            <DrawerFooter>
                                                <Button>Confirm</Button>
                                            </DrawerFooter>
                                        </DrawerContent>
                                    </Drawer>

                                    <Sheet>
                                        <SheetTrigger asChild>
                                            <Button variant="outline">Open sheet</Button>
                                        </SheetTrigger>
                                        <SheetContent>
                                            <SheetHeader>
                                                <SheetTitle>Right-side inspector</SheetTitle>
                                                <SheetDescription>Perfect for logs, resource details, and quick edits.</SheetDescription>
                                            </SheetHeader>
                                        </SheetContent>
                                    </Sheet>

                                    <ContextMenu>
                                        <ContextMenuTrigger asChild>
                                            <div className="rounded-xl border border-dashed border-zinc-700 px-4 py-3 text-sm text-zinc-400">
                                                Right click this box
                                            </div>
                                        </ContextMenuTrigger>
                                        <ContextMenuContent>
                                            <ContextMenuItem>Restart container</ContextMenuItem>
                                            <ContextMenuItem>Open logs</ContextMenuItem>
                                        </ContextMenuContent>
                                    </ContextMenu>
                                </CardContent>
                            </Card>
                        </div>
                    </TabsContent>

                    <TabsContent value="navigation" className="space-y-6">
                        <div className="grid gap-6 lg:grid-cols-2">
                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Pagination and button groups</CardTitle>
                                    <CardDescription>Typical dashboard controls.</CardDescription>
                                </CardHeader>
                                <CardContent className="space-y-6">
                                    <ButtonGroup>
                                        <Button variant="secondary">Overview</Button>
                                        <Button variant="outline">Containers</Button>
                                        <Button variant="outline">Logs</Button>
                                    </ButtonGroup>

                                    <Pagination>
                                        <PaginationContent>
                                            <PaginationItem>
                                                <PaginationPrevious href="#" />
                                            </PaginationItem>
                                            <PaginationItem>
                                                <PaginationLink href="#" isActive>
                                                    1
                                                </PaginationLink>
                                            </PaginationItem>
                                            <PaginationItem>
                                                <PaginationLink href="#">2</PaginationLink>
                                            </PaginationItem>
                                            <PaginationItem>
                                                <PaginationNext href="#" />
                                            </PaginationItem>
                                        </PaginationContent>
                                    </Pagination>
                                </CardContent>
                            </Card>

                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Collapsible content</CardTitle>
                                    <CardDescription>Compact detail panels for operators.</CardDescription>
                                </CardHeader>
                                <CardContent>
                                    <Collapsible>
                                        <CollapsibleTrigger asChild>
                                            <Button variant="outline">Toggle diagnostics</Button>
                                        </CollapsibleTrigger>
                                        <CollapsibleContent className="mt-4 rounded-xl border border-zinc-800 p-4 text-sm text-zinc-300">
                                            Diagnostics: ingress healthy, database latency stable, queue backlog under threshold.
                                        </CollapsibleContent>
                                    </Collapsible>
                                </CardContent>
                            </Card>
                        </div>
                    </TabsContent>

                    <TabsContent value="display" className="space-y-6">
                        <div className="grid gap-6 xl:grid-cols-3">
                            <Card className="xl:col-span-2 rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Chart</CardTitle>
                                    <CardDescription>Using the chart wrapper with Recharts.</CardDescription>
                                </CardHeader>
                                <CardContent>
                                    <ChartContainer config={chartConfig} className="h-[280px] w-full">
                                        <ResponsiveContainer width="100%" height="100%">
                                            <BarChart data={chartData}>
                                                <CartesianGrid vertical={false} />
                                                <XAxis dataKey="name" tickLine={false} axisLine={false} />
                                                <ChartTooltip content={<ChartTooltipContent />} />
                                                <Bar dataKey="deployments" radius={8} />
                                            </BarChart>
                                        </ResponsiveContainer>
                                    </ChartContainer>
                                </CardContent>
                            </Card>

                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Carousel</CardTitle>
                                    <CardDescription>Good for onboarding or feature banners.</CardDescription>
                                </CardHeader>
                                <CardContent>
                                    <Carousel className="w-full">
                                        <CarouselContent>
                                            {["Compute", "Storage", "Networking"].map((item) => (
                                                <CarouselItem key={item}>
                                                    <div className="rounded-2xl border border-zinc-800 bg-zinc-950 p-8 text-center">
                                                        <p className="text-sm text-zinc-400">Module</p>
                                                        <p className="mt-2 text-2xl font-semibold">{item}</p>
                                                    </div>
                                                </CarouselItem>
                                            ))}
                                        </CarouselContent>
                                        <CarouselPrevious />
                                        <CarouselNext />
                                    </Carousel>
                                </CardContent>
                            </Card>
                        </div>

                        <div className="grid gap-6 lg:grid-cols-2">
                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Item rows and table</CardTitle>
                                </CardHeader>
                                <CardContent className="space-y-6">
                                    <div className="space-y-3">
                                        <Item>
                                            <Bell className="h-4 w-4" />
                                            <div>
                                                <ItemTitle>Node maintenance scheduled</ItemTitle>
                                                <ItemDescription>One worker node will reboot at 02:00.</ItemDescription>
                                            </div>
                                        </Item>
                                        <Separator />
                                        <Item>
                                            <Shield className="h-4 w-4" />
                                            <div>
                                                <ItemTitle>WAF policy applied</ItemTitle>
                                                <ItemDescription>Premium tenants are protected by the latest rule set.</ItemDescription>
                                            </div>
                                        </Item>
                                    </div>

                                    <Table>
                                        <TableHeader>
                                            <TableRow>
                                                <TableHead>Service</TableHead>
                                                <TableHead>Status</TableHead>
                                                <TableHead>Region</TableHead>
                                            </TableRow>
                                        </TableHeader>
                                        <TableBody>
                                            <TableRow>
                                                <TableCell>platform-api</TableCell>
                                                <TableCell>Healthy</TableCell>
                                                <TableCell>CA-East</TableCell>
                                            </TableRow>
                                            <TableRow>
                                                <TableCell>dashboard</TableCell>
                                                <TableCell>Healthy</TableCell>
                                                <TableCell>CA-East</TableCell>
                                            </TableRow>
                                        </TableBody>
                                    </Table>
                                </CardContent>
                            </Card>

                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Scroll area + avatars</CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <ScrollArea className="h-64 rounded-xl border border-zinc-800 p-4">
                                        <div className="space-y-4">
                                            {Array.from({ length: 8 }).map((_, i) => (
                                                <div key={i} className="flex items-center gap-3 rounded-xl bg-zinc-950 p-3">
                                                    <Avatar>
                                                        <AvatarImage src={`https://i.pravatar.cc/100?img=${i + 1}`} />
                                                        <AvatarFallback>KF</AvatarFallback>
                                                    </Avatar>
                                                    <div>
                                                        <p className="font-medium">Operator {i + 1}</p>
                                                        <p className="text-sm text-zinc-400">Reviewed deployment logs and metrics.</p>
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    </ScrollArea>
                                </CardContent>
                            </Card>
                        </div>
                    </TabsContent>

                    <TabsContent value="feedback" className="space-y-6">
                        <div className="grid gap-6 lg:grid-cols-3">
                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Progress</CardTitle>
                                </CardHeader>
                                <CardContent className="space-y-4">
                                    <Progress value={progress} />
                                    <Button variant="outline" onClick={() => setProgress((p) => Math.min(p + 10, 100))}>
                                        Increase progress
                                    </Button>
                                </CardContent>
                            </Card>

                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Skeleton loaders</CardTitle>
                                </CardHeader>
                                <CardContent className="space-y-3">
                                    <Skeleton className="h-5 w-1/3" />
                                    <Skeleton className="h-20 w-full" />
                                    <Skeleton className="h-10 w-2/3" />
                                </CardContent>
                            </Card>

                            <Card className="rounded-2xl bg-zinc-900/80 border-zinc-800">
                                <CardHeader>
                                    <CardTitle>Empty state</CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <Empty>
                                        <EmptyHeader>
                                            <EmptyTitle>No deployments yet</EmptyTitle>
                                            <EmptyDescription>Create your first workload to see activity here.</EmptyDescription>
                                        </EmptyHeader>
                                    </Empty>
                                </CardContent>
                            </Card>
                        </div>
                    </TabsContent>
                </Tabs>
            </div>

            <Toaster />
        </div>
    );
}
