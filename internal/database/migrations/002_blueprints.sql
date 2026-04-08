CREATE TABLE IF NOT EXISTS crates (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    category    TEXT NOT NULL,
    description TEXT NOT NULL,
    logo        TEXT NOT NULL DEFAULT '',
    tags        JSONB NOT NULL DEFAULT '[]',
    official    BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS blueprints (
    id               TEXT PRIMARY KEY,
    crate_id         TEXT NOT NULL REFERENCES crates(id) ON DELETE CASCADE,
    name             TEXT NOT NULL,
    description      TEXT NOT NULL,
    long_description TEXT NOT NULL DEFAULT '',
    logo             TEXT NOT NULL DEFAULT '',
    image            TEXT NOT NULL,
    version          TEXT NOT NULL,
    official         BOOLEAN NOT NULL DEFAULT false,
    category         TEXT NOT NULL,
    runtime_hints    JSONB NOT NULL DEFAULT '{}',
    resources        JSONB NOT NULL DEFAULT '{}',
    ports            JSONB NOT NULL DEFAULT '[]',
    config           JSONB NOT NULL DEFAULT '[]',
    outputs          JSONB NOT NULL DEFAULT '[]',
    extensions       JSONB NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Seed: Crates ──────────────────────────────────────────────────────────────

INSERT INTO crates (id, name, category, description, logo, tags, official) VALUES
('minecraft',   'Minecraft',   'game-server', 'The world''s most popular game. Java and Bedrock editions.', '', '["sandbox","survival","multiplayer"]', true),
('fivem',       'FiveM',       'game-server', 'GTA V multiplayer modification platform.',                   '', '["gta","roleplay","multiplayer"]',    true),
('gmod',        'Garry''s Mod','game-server', 'A physics sandbox game with extensive workshop support.',    '', '["sandbox","workshop"]',              true),
('cs2',         'CS2',         'game-server', 'Counter-Strike 2 dedicated server.',                        '', '["fps","competitive"]',               true),
('rust-game',   'Rust',        'game-server', 'Survival multiplayer game with Oxide plugin support.',      '', '["survival","oxide"]',                true),
('valheim',     'Valheim',     'game-server', 'Viking survival and exploration game.',                     '', '["survival","viking"]',               true),
('ark',         'ARK',         'game-server', 'ARK: Survival Evolved and Survival Ascended.',              '', '["survival","dinosaurs"]',            true),
('redis',       'Redis',       'cache',       'In-memory data store, cache, and message broker.',          '', '["cache","in-memory"]',               true),
('postgresql',  'PostgreSQL',  'database',    'The world''s most advanced open source relational database.','', '["sql","relational"]',               true),
('mysql',       'MySQL',       'database',    'Popular open source relational database.',                  '', '["sql","relational"]',                true),
('mongodb',     'MongoDB',     'database',    'General purpose, document-based distributed database.',     '', '["nosql","document"]',               true),
('rabbitmq',    'RabbitMQ',    'messaging',   'Open source message broker with management UI.',            '', '["amqp","messaging"]',               true),
('nginx',       'Nginx',       'web',         'High-performance HTTP server and reverse proxy.',           '', '["proxy","web","http"]',              true),
('caddy',       'Caddy',       'web',         'Modern web server with automatic HTTPS.',                   '', '["proxy","https","web"]',             true)
ON CONFLICT (id) DO NOTHING;

-- ── Seed: Blueprints ──────────────────────────────────────────────────────────

INSERT INTO blueprints (id, crate_id, name, description, image, version, official, category, runtime_hints, resources, ports, config, outputs, extensions) VALUES

-- Minecraft
('minecraft-papermc', 'minecraft', 'PaperMC', 'High-performance Minecraft server fork with plugin support.',
 'itzg/minecraft-server', '1.3.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":true,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":2048,"cpu_millicores":2000,"disk_gb":10}',
 '[{"name":"minecraft","container":25565,"protocol":"tcp","expose":true,"label":"Minecraft (Java)"},{"name":"rcon","container":25575,"protocol":"tcp","expose":false,"label":"RCON"},{"name":"query","container":25565,"protocol":"udp","expose":true,"label":"Query"}]',
 '[{"key":"EULA","label":"Accept EULA","description":"You must accept the Minecraft EULA to run a server.","type":"boolean","required":true,"default":true},{"key":"TYPE","label":"Server Type","type":"select","options":["PAPER"],"default":"PAPER","required":true},{"key":"VERSION","label":"Minecraft Version","type":"select","options":["LATEST","1.21.4","1.21.3","1.20.4","1.20.1"],"default":"LATEST","required":true},{"key":"DIFFICULTY","label":"Difficulty","type":"select","options":["peaceful","easy","normal","hard"],"default":"normal","required":false},{"key":"MAX_PLAYERS","label":"Max Players","type":"number","default":20,"required":false},{"key":"MOTD","label":"MOTD","description":"Message shown in the server list.","type":"string","default":"A Minecraft Server","required":false},{"key":"RCON_PASSWORD","label":"RCON Password","type":"secret","required":false,"auto_generate":true,"auto_generate_length":16},{"key":"MEMORY","label":"JVM Memory","description":"Java heap size, e.g. 2G or 4G.","type":"string","default":"2G","required":false}]',
 '[{"key":"ADDRESS","description":"Internal host:port for other services to connect to.","value_template":"{{ .ContainerName }}:25565"},{"key":"RCON_ADDRESS","description":"RCON host:port.","value_template":"{{ .ContainerName }}:25575"}]',
 '{"plugin":{"enabled":true,"install_method":"jar-drop","install_path":"/data/plugins/","file_extension":".jar","config_path":"/data/plugins/","requires_restart":true,"sources":["modrinth","hangar","spigotmc","github-release","upload"]}}'
),

('minecraft-spigot', 'minecraft', 'Spigot', 'Widely used Minecraft server with plugin support.',
 'itzg/minecraft-server', '1.2.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":true,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":2048,"cpu_millicores":2000,"disk_gb":10}',
 '[{"name":"minecraft","container":25565,"protocol":"tcp","expose":true,"label":"Minecraft (Java)"},{"name":"rcon","container":25575,"protocol":"tcp","expose":false,"label":"RCON"}]',
 '[{"key":"EULA","label":"Accept EULA","type":"boolean","required":true,"default":true},{"key":"TYPE","label":"Server Type","type":"select","options":["SPIGOT"],"default":"SPIGOT","required":true},{"key":"VERSION","label":"Minecraft Version","type":"select","options":["LATEST","1.21.4","1.20.4","1.20.1"],"default":"LATEST","required":true},{"key":"DIFFICULTY","label":"Difficulty","type":"select","options":["peaceful","easy","normal","hard"],"default":"normal","required":false},{"key":"MAX_PLAYERS","label":"Max Players","type":"number","default":20,"required":false},{"key":"MEMORY","label":"JVM Memory","type":"string","default":"2G","required":false}]',
 '[{"key":"ADDRESS","description":"Internal host:port.","value_template":"{{ .ContainerName }}:25565"}]',
 '{"plugin":{"enabled":true,"install_method":"jar-drop","install_path":"/data/plugins/","file_extension":".jar","config_path":"/data/plugins/","requires_restart":true,"sources":["modrinth","hangar","spigotmc","github-release","upload"]}}'
),

('minecraft-forge', 'minecraft', 'Forge', 'Modded Minecraft with Forge mod loader.',
 'itzg/minecraft-server', '1.2.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":4096,"cpu_millicores":2000,"disk_gb":20}',
 '[{"name":"minecraft","container":25565,"protocol":"tcp","expose":true,"label":"Minecraft (Java)"},{"name":"rcon","container":25575,"protocol":"tcp","expose":false,"label":"RCON"}]',
 '[{"key":"EULA","label":"Accept EULA","type":"boolean","required":true,"default":true},{"key":"TYPE","label":"Server Type","type":"select","options":["FORGE"],"default":"FORGE","required":true},{"key":"VERSION","label":"Minecraft Version","type":"select","options":["LATEST","1.21.4","1.20.4","1.20.1"],"default":"LATEST","required":true},{"key":"FORGEVERSION","label":"Forge Version","type":"string","default":"RECOMMENDED","required":false},{"key":"MAX_PLAYERS","label":"Max Players","type":"number","default":20,"required":false},{"key":"MEMORY","label":"JVM Memory","type":"string","default":"4G","required":false}]',
 '[{"key":"ADDRESS","description":"Internal host:port.","value_template":"{{ .ContainerName }}:25565"}]',
 '{"mod":{"enabled":true,"install_method":"jar-drop","install_path":"/data/mods/","file_extension":".jar","requires_restart":true,"sources":["modrinth","curseforge","github-release","upload"]}}'
),

('minecraft-fabric', 'minecraft', 'Fabric', 'Lightweight modded Minecraft with Fabric mod loader.',
 'itzg/minecraft-server', '1.1.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":3072,"cpu_millicores":2000,"disk_gb":15}',
 '[{"name":"minecraft","container":25565,"protocol":"tcp","expose":true,"label":"Minecraft (Java)"},{"name":"rcon","container":25575,"protocol":"tcp","expose":false,"label":"RCON"}]',
 '[{"key":"EULA","label":"Accept EULA","type":"boolean","required":true,"default":true},{"key":"TYPE","label":"Server Type","type":"select","options":["FABRIC"],"default":"FABRIC","required":true},{"key":"VERSION","label":"Minecraft Version","type":"select","options":["LATEST","1.21.4","1.20.4","1.20.1"],"default":"LATEST","required":true},{"key":"MAX_PLAYERS","label":"Max Players","type":"number","default":20,"required":false},{"key":"MEMORY","label":"JVM Memory","type":"string","default":"3G","required":false}]',
 '[{"key":"ADDRESS","description":"Internal host:port.","value_template":"{{ .ContainerName }}:25565"}]',
 '{"mod":{"enabled":true,"install_method":"jar-drop","install_path":"/data/mods/","file_extension":".jar","requires_restart":true,"sources":["modrinth","curseforge","github-release","upload"]}}'
),

('minecraft-arclight', 'minecraft', 'Arclight', 'Hybrid server supporting both plugins and mods simultaneously.',
 'itzg/minecraft-server', '1.1.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":4096,"cpu_millicores":2000,"disk_gb":20}',
 '[{"name":"minecraft","container":25565,"protocol":"tcp","expose":true,"label":"Minecraft (Java)"},{"name":"rcon","container":25575,"protocol":"tcp","expose":false,"label":"RCON"}]',
 '[{"key":"EULA","label":"Accept EULA","type":"boolean","required":true,"default":true},{"key":"TYPE","label":"Server Type","type":"select","options":["ARCLIGHT"],"default":"ARCLIGHT","required":true},{"key":"VERSION","label":"Minecraft Version","type":"select","options":["1.21.4","1.20.4","1.20.1"],"default":"1.21.4","required":true},{"key":"MAX_PLAYERS","label":"Max Players","type":"number","default":20,"required":false},{"key":"MEMORY","label":"JVM Memory","type":"string","default":"4G","required":false}]',
 '[{"key":"ADDRESS","description":"Internal host:port.","value_template":"{{ .ContainerName }}:25565"}]',
 '{"plugin":{"enabled":true,"install_method":"jar-drop","install_path":"/data/plugins/","file_extension":".jar","config_path":"/data/plugins/","requires_restart":true,"sources":["modrinth","hangar","spigotmc","github-release","upload"]},"mod":{"enabled":true,"install_method":"jar-drop","install_path":"/data/mods/","file_extension":".jar","requires_restart":true,"sources":["modrinth","curseforge","github-release","upload"]}}'
),

('minecraft-velocity', 'minecraft', 'Velocity', 'High-performance Minecraft proxy server.',
 'itzg/mc-proxy', '1.2.0', true, 'proxy',
 '{"kubernetes_strategy":"agones","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":1024,"cpu_millicores":1000,"disk_gb":5}',
 '[{"name":"minecraft","container":25565,"protocol":"tcp","expose":true,"label":"Minecraft Proxy"}]',
 '[{"key":"TYPE","label":"Proxy Type","type":"select","options":["VELOCITY"],"default":"VELOCITY","required":true},{"key":"MEMORY","label":"JVM Memory","type":"string","default":"1G","required":false}]',
 '[{"key":"ADDRESS","description":"Proxy address for players to connect to.","value_template":"{{ .ContainerName }}:25565"}]',
 '{"plugin":{"enabled":true,"install_method":"jar-drop","install_path":"/server/plugins/","file_extension":".jar","config_path":"/server/plugins/","requires_restart":true,"sources":["modrinth","hangar","github-release","upload"]}}'
),

('minecraft-bungeecord', 'minecraft', 'BungeeCord', 'Classic Minecraft proxy server.',
 'itzg/mc-proxy', '1.1.0', true, 'proxy',
 '{"kubernetes_strategy":"agones","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":1024,"cpu_millicores":1000,"disk_gb":5}',
 '[{"name":"minecraft","container":25577,"protocol":"tcp","expose":true,"label":"Minecraft Proxy"}]',
 '[{"key":"TYPE","label":"Proxy Type","type":"select","options":["BUNGEECORD"],"default":"BUNGEECORD","required":true},{"key":"MEMORY","label":"JVM Memory","type":"string","default":"1G","required":false}]',
 '[{"key":"ADDRESS","description":"Proxy address for players to connect to.","value_template":"{{ .ContainerName }}:25577"}]',
 '{"plugin":{"enabled":true,"install_method":"jar-drop","install_path":"/server/plugins/","file_extension":".jar","config_path":"/server/plugins/","requires_restart":true,"sources":["github-release","upload"]}}'
),

('minecraft-bedrock', 'minecraft', 'Bedrock', 'Minecraft Bedrock Edition dedicated server.',
 'itzg/minecraft-bedrock-server', '1.0.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":true,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":1024,"cpu_millicores":1000,"disk_gb":10}',
 '[{"name":"bedrock","container":19132,"protocol":"udp","expose":true,"label":"Minecraft (Bedrock)"}]',
 '[{"key":"EULA","label":"Accept EULA","type":"boolean","required":true,"default":true},{"key":"VERSION","label":"Bedrock Version","type":"string","default":"LATEST","required":true},{"key":"DIFFICULTY","label":"Difficulty","type":"select","options":["peaceful","easy","normal","hard"],"default":"normal","required":false},{"key":"MAX_PLAYERS","label":"Max Players","type":"number","default":10,"required":false}]',
 '[{"key":"ADDRESS","description":"Bedrock server address.","value_template":"{{ .ContainerName }}:19132"}]',
 '{}'
),

-- FiveM
('fivem-server', 'fivem', 'FiveM Server', 'Standard FiveM GTA V multiplayer server.',
 'ghcr.io/fersuazo/fivem-server-docker:latest', '1.0.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":true,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":4096,"cpu_millicores":2000,"disk_gb":20}',
 '[{"name":"game","container":30120,"protocol":"tcp","expose":true,"label":"FiveM"},{"name":"game-udp","container":30120,"protocol":"udp","expose":true,"label":"FiveM (UDP)"}]',
 '[{"key":"LICENSE_KEY","label":"FiveM License Key","description":"Your Cfx.re license key.","type":"secret","required":true},{"key":"SERVER_NAME","label":"Server Name","type":"string","default":"FiveM Server","required":false},{"key":"MAX_PLAYERS","label":"Max Players","type":"number","default":32,"required":false},{"key":"STEAM_WEB_API_KEY","label":"Steam Web API Key","type":"secret","required":false}]',
 '[{"key":"ADDRESS","description":"Server address.","value_template":"{{ .ContainerName }}:30120"}]',
 '{"resource":{"enabled":true,"install_method":"folder-drop","install_path":"/txData/resources/","requires_restart":false,"live_commands":{"start":"start {name}","stop":"stop {name}","restart":"restart {name}"},"sources":["github-release","cfx-resource","upload"]}}'
),

-- GMod
('gmod-server', 'gmod', 'Garry''s Mod', 'Garry''s Mod dedicated server with addon and workshop support.',
 'gameservermanagers/gameserver:gmodserver', '1.0.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":true,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":4096,"cpu_millicores":2000,"disk_gb":30}',
 '[{"name":"game","container":27015,"protocol":"udp","expose":true,"label":"Game"},{"name":"game-tcp","container":27015,"protocol":"tcp","expose":true,"label":"Game (TCP)"}]',
 '[{"key":"GAMEMODE","label":"Gamemode","type":"string","default":"sandbox","required":false},{"key":"MAP","label":"Default Map","type":"string","default":"gm_flatgrass","required":false},{"key":"MAXPLAYERS","label":"Max Players","type":"number","default":16,"required":false},{"key":"SERVERNAME","label":"Server Name","type":"string","default":"Garry''s Mod Server","required":false}]',
 '[{"key":"ADDRESS","description":"Server address.","value_template":"{{ .ContainerName }}:27015"}]',
 '{"addon":{"enabled":true,"install_method":"folder-drop","install_path":"/serverdata/garrysmod/addons/","requires_restart":true,"sources":["github-release","upload"]},"workshop":{"enabled":true,"install_method":"workshop-id","config_file":"/serverdata/garrysmod/cfg/workshop.cfg","config_entry_template":"resource.AddWorkshop(\"{id}\")","requires_restart":true,"sources":["steam-workshop"]}}'
),

-- CS2
('cs2-server', 'cs2', 'CS2 Server', 'Counter-Strike 2 dedicated server with MetaMod/SourceMod support.',
 'joedwards32/cs2', '1.0.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":true,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":4096,"cpu_millicores":2000,"disk_gb":30}',
 '[{"name":"game","container":27015,"protocol":"udp","expose":true,"label":"Game"},{"name":"game-tcp","container":27015,"protocol":"tcp","expose":true,"label":"Game (TCP)"}]',
 '[{"key":"CS2_SERVERNAME","label":"Server Name","type":"string","default":"CS2 Server","required":false},{"key":"CS2_MAXPLAYERS","label":"Max Players","type":"number","default":10,"required":false},{"key":"CS2_GAMETYPE","label":"Game Type","type":"number","default":0,"required":false},{"key":"CS2_GAMEMODE","label":"Game Mode","type":"number","default":1,"required":false},{"key":"CS2_STARTMAP","label":"Start Map","type":"string","default":"de_dust2","required":false},{"key":"CS2_RCON_PORT","label":"RCON Port","type":"number","default":27050,"required":false}]',
 '[{"key":"ADDRESS","description":"Server address.","value_template":"{{ .ContainerName }}:27015"}]',
 '{"sourcemod-plugin":{"enabled":true,"install_method":"file-drop","install_path":"/home/steam/cs2-dedicated/game/csgo/addons/sourcemod/plugins/","file_extension":".smx","requires_restart":false,"live_commands":{"load":"sm plugins load {filename}","unload":"sm plugins unload {filename}","reload":"sm plugins reload {filename}"},"sources":["alliedmodders","github-release","upload"]}}'
),

-- Rust
('rust-server', 'rust-game', 'Rust Server', 'Rust dedicated server with Oxide/uMod plugin support.',
 'didstopia/rust-server', '1.0.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":true,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":8192,"cpu_millicores":4000,"disk_gb":30}',
 '[{"name":"game","container":28015,"protocol":"udp","expose":true,"label":"Game"},{"name":"rcon","container":28016,"protocol":"tcp","expose":false,"label":"RCON"}]',
 '[{"key":"RUST_SERVER_NAME","label":"Server Name","type":"string","default":"Rust Server","required":false},{"key":"RUST_SERVER_MAXPLAYERS","label":"Max Players","type":"number","default":100,"required":false},{"key":"RUST_SERVER_WORLDSIZE","label":"World Size","type":"number","default":3500,"required":false},{"key":"RUST_SERVER_SEED","label":"World Seed","type":"number","default":12345,"required":false},{"key":"RUST_RCON_PASSWORD","label":"RCON Password","type":"secret","required":true,"auto_generate":true,"auto_generate_length":16}]',
 '[{"key":"ADDRESS","description":"Server address.","value_template":"{{ .ContainerName }}:28015"}]',
 '{"oxide-plugin":{"enabled":true,"install_method":"file-drop","install_path":"/serverdata/oxide/plugins/","file_extension":".cs","requires_restart":false,"live_commands":{"reload":"oxide.reload {name}"},"sources":["umod","github-release","upload"]}}'
),

-- Valheim
('valheim-server', 'valheim', 'Valheim Server', 'Valheim dedicated server.',
 'lloesche/valheim-server', '1.0.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":true,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":4096,"cpu_millicores":2000,"disk_gb":10}',
 '[{"name":"game","container":2456,"protocol":"udp","expose":true,"label":"Game"},{"name":"game2","container":2457,"protocol":"udp","expose":true,"label":"Game 2"}]',
 '[{"key":"SERVER_NAME","label":"Server Name","type":"string","default":"Valheim Server","required":true},{"key":"WORLD_NAME","label":"World Name","type":"string","default":"Dedicated","required":true},{"key":"SERVER_PASS","label":"Server Password","description":"Minimum 5 characters.","type":"secret","required":true},{"key":"SERVER_PUBLIC","label":"Public Server","type":"boolean","default":true,"required":false}]',
 '[{"key":"ADDRESS","description":"Server address.","value_template":"{{ .ContainerName }}:2456"}]',
 '{}'
),

-- ARK
('ark-ase', 'ark', 'ARK: Survival Evolved', 'ARK: Survival Evolved dedicated server.',
 'hermsi1337/ark-survival-evolved', '1.0.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":true,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":8192,"cpu_millicores":4000,"disk_gb":50}',
 '[{"name":"game","container":7777,"protocol":"udp","expose":true,"label":"Game"},{"name":"query","container":27015,"protocol":"udp","expose":true,"label":"Query"},{"name":"rcon","container":27020,"protocol":"tcp","expose":false,"label":"RCON"}]',
 '[{"key":"SESSIONNAME","label":"Session Name","type":"string","default":"ARK Server","required":true},{"key":"SERVERPASSWORD","label":"Server Password","type":"secret","required":false},{"key":"ADMINPASSWORD","label":"Admin Password","type":"secret","required":true,"auto_generate":true,"auto_generate_length":16},{"key":"MAXPLAYERS","label":"Max Players","type":"number","default":70,"required":false},{"key":"MAP","label":"Map","type":"select","options":["TheIsland","TheCenter","ScorchedEarth_P","Ragnarok","Aberration_P","Extinction","Valguero_P","Genesis","CrystalIsles","Gen2"],"default":"TheIsland","required":false}]',
 '[{"key":"ADDRESS","description":"Server address.","value_template":"{{ .ContainerName }}:7777"}]',
 '{}'
),

('ark-sa', 'ark', 'ARK: Survival Ascended', 'ARK: Survival Ascended dedicated server.',
 'acekorneya/asa-server', '1.0.0', true, 'game-server',
 '{"kubernetes_strategy":"agones","expose_udp":true,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":16384,"cpu_millicores":4000,"disk_gb":60}',
 '[{"name":"game","container":7777,"protocol":"udp","expose":true,"label":"Game"},{"name":"query","container":27015,"protocol":"udp","expose":true,"label":"Query"}]',
 '[{"key":"SESSIONNAME","label":"Session Name","type":"string","default":"ASA Server","required":true},{"key":"SERVERPASSWORD","label":"Server Password","type":"secret","required":false},{"key":"ADMINPASSWORD","label":"Admin Password","type":"secret","required":true,"auto_generate":true,"auto_generate_length":16},{"key":"MAXPLAYERS","label":"Max Players","type":"number","default":70,"required":false}]',
 '[{"key":"ADDRESS","description":"Server address.","value_template":"{{ .ContainerName }}:7777"}]',
 '{}'
),

-- Redis
('redis-standalone', 'redis', 'Redis Standalone', 'Single Redis instance for caching and session storage.',
 'redis:7-alpine', '1.0.0', true, 'cache',
 '{"kubernetes_strategy":"statefulset","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":512,"cpu_millicores":500,"disk_gb":5}',
 '[{"name":"redis","container":6379,"protocol":"tcp","expose":false,"label":"Redis"}]',
 '[{"key":"REDIS_PASSWORD","label":"Password","type":"secret","required":false,"auto_generate":false},{"key":"REDIS_MAXMEMORY","label":"Max Memory","description":"e.g. 256mb or 1gb","type":"string","default":"256mb","required":false},{"key":"REDIS_MAXMEMORY_POLICY","label":"Eviction Policy","type":"select","options":["noeviction","allkeys-lru","volatile-lru","allkeys-random","volatile-random","volatile-ttl"],"default":"allkeys-lru","required":false}]',
 '[{"key":"ADDRESS","description":"Redis host:port for other services.","value_template":"{{ .ContainerName }}:6379"}]',
 '{}'
),

('redis-cluster', 'redis', 'Redis Cluster', 'Redis Cluster with 3+ nodes for high availability.',
 'redis:7-alpine', '1.0.0', true, 'cache',
 '{"kubernetes_strategy":"statefulset","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":1024,"cpu_millicores":1000,"disk_gb":10}',
 '[{"name":"redis","container":6379,"protocol":"tcp","expose":false,"label":"Redis"},{"name":"cluster-bus","container":16379,"protocol":"tcp","expose":false,"label":"Cluster Bus"}]',
 '[{"key":"REDIS_PASSWORD","label":"Password","type":"secret","required":false,"auto_generate":false},{"key":"REDIS_MAXMEMORY","label":"Max Memory Per Node","type":"string","default":"256mb","required":false}]',
 '[{"key":"ADDRESS","description":"Redis cluster seed node address.","value_template":"{{ .ContainerName }}:6379"}]',
 '{}'
),

-- PostgreSQL
('postgresql-16', 'postgresql', 'PostgreSQL 16', 'PostgreSQL 16 relational database.',
 'postgres:16-alpine', '1.0.0', true, 'database',
 '{"kubernetes_strategy":"statefulset","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":1024,"cpu_millicores":1000,"disk_gb":20}',
 '[{"name":"postgres","container":5432,"protocol":"tcp","expose":false,"label":"PostgreSQL"}]',
 '[{"key":"POSTGRES_USER","label":"Username","type":"string","default":"postgres","required":true},{"key":"POSTGRES_PASSWORD","label":"Password","type":"secret","required":true,"auto_generate":true,"auto_generate_length":24},{"key":"POSTGRES_DB","label":"Default Database","type":"string","default":"app","required":true}]',
 '[{"key":"ADDRESS","description":"PostgreSQL host:port.","value_template":"{{ .ContainerName }}:5432"},{"key":"DSN","description":"Full connection string.","value_template":"postgres://{{ .Env.POSTGRES_USER }}:{{ .Env.POSTGRES_PASSWORD }}@{{ .ContainerName }}:5432/{{ .Env.POSTGRES_DB }}"}]',
 '{}'
),

('postgresql-15', 'postgresql', 'PostgreSQL 15', 'PostgreSQL 15 relational database.',
 'postgres:15-alpine', '1.0.0', true, 'database',
 '{"kubernetes_strategy":"statefulset","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":1024,"cpu_millicores":1000,"disk_gb":20}',
 '[{"name":"postgres","container":5432,"protocol":"tcp","expose":false,"label":"PostgreSQL"}]',
 '[{"key":"POSTGRES_USER","label":"Username","type":"string","default":"postgres","required":true},{"key":"POSTGRES_PASSWORD","label":"Password","type":"secret","required":true,"auto_generate":true,"auto_generate_length":24},{"key":"POSTGRES_DB","label":"Default Database","type":"string","default":"app","required":true}]',
 '[{"key":"ADDRESS","description":"PostgreSQL host:port.","value_template":"{{ .ContainerName }}:5432"}]',
 '{}'
),

('postgresql-14', 'postgresql', 'PostgreSQL 14', 'PostgreSQL 14 relational database.',
 'postgres:14-alpine', '1.0.0', true, 'database',
 '{"kubernetes_strategy":"statefulset","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":1024,"cpu_millicores":1000,"disk_gb":20}',
 '[{"name":"postgres","container":5432,"protocol":"tcp","expose":false,"label":"PostgreSQL"}]',
 '[{"key":"POSTGRES_USER","label":"Username","type":"string","default":"postgres","required":true},{"key":"POSTGRES_PASSWORD","label":"Password","type":"secret","required":true,"auto_generate":true,"auto_generate_length":24},{"key":"POSTGRES_DB","label":"Default Database","type":"string","default":"app","required":true}]',
 '[{"key":"ADDRESS","description":"PostgreSQL host:port.","value_template":"{{ .ContainerName }}:5432"}]',
 '{}'
),

-- MySQL / MariaDB
('mysql-8', 'mysql', 'MySQL 8', 'MySQL 8 relational database.',
 'mysql:8', '1.0.0', true, 'database',
 '{"kubernetes_strategy":"statefulset","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":1024,"cpu_millicores":1000,"disk_gb":20}',
 '[{"name":"mysql","container":3306,"protocol":"tcp","expose":false,"label":"MySQL"}]',
 '[{"key":"MYSQL_ROOT_PASSWORD","label":"Root Password","type":"secret","required":true,"auto_generate":true,"auto_generate_length":24},{"key":"MYSQL_DATABASE","label":"Default Database","type":"string","default":"app","required":false},{"key":"MYSQL_USER","label":"Username","type":"string","default":"kleff","required":false},{"key":"MYSQL_PASSWORD","label":"User Password","type":"secret","required":false,"auto_generate":true,"auto_generate_length":16}]',
 '[{"key":"ADDRESS","description":"MySQL host:port.","value_template":"{{ .ContainerName }}:3306"}]',
 '{}'
),

('mariadb-11', 'mysql', 'MariaDB 11', 'MariaDB 11 relational database.',
 'mariadb:11', '1.0.0', true, 'database',
 '{"kubernetes_strategy":"statefulset","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":1024,"cpu_millicores":1000,"disk_gb":20}',
 '[{"name":"mariadb","container":3306,"protocol":"tcp","expose":false,"label":"MariaDB"}]',
 '[{"key":"MARIADB_ROOT_PASSWORD","label":"Root Password","type":"secret","required":true,"auto_generate":true,"auto_generate_length":24},{"key":"MARIADB_DATABASE","label":"Default Database","type":"string","default":"app","required":false},{"key":"MARIADB_USER","label":"Username","type":"string","default":"kleff","required":false},{"key":"MARIADB_PASSWORD","label":"User Password","type":"secret","required":false,"auto_generate":true,"auto_generate_length":16}]',
 '[{"key":"ADDRESS","description":"MariaDB host:port.","value_template":"{{ .ContainerName }}:3306"}]',
 '{}'
),

-- MongoDB
('mongodb-7', 'mongodb', 'MongoDB 7', 'MongoDB 7 document database.',
 'mongo:7', '1.0.0', true, 'database',
 '{"kubernetes_strategy":"statefulset","expose_udp":false,"health_check_path":"","health_check_port":0}',
 '{"memory_mb":1024,"cpu_millicores":1000,"disk_gb":20}',
 '[{"name":"mongodb","container":27017,"protocol":"tcp","expose":false,"label":"MongoDB"}]',
 '[{"key":"MONGO_INITDB_ROOT_USERNAME","label":"Root Username","type":"string","default":"admin","required":true},{"key":"MONGO_INITDB_ROOT_PASSWORD","label":"Root Password","type":"secret","required":true,"auto_generate":true,"auto_generate_length":24},{"key":"MONGO_INITDB_DATABASE","label":"Default Database","type":"string","default":"app","required":false}]',
 '[{"key":"ADDRESS","description":"MongoDB host:port.","value_template":"{{ .ContainerName }}:27017"}]',
 '{}'
),

-- RabbitMQ
('rabbitmq', 'rabbitmq', 'RabbitMQ', 'RabbitMQ message broker with management UI.',
 'rabbitmq:3-management-alpine', '1.0.0', true, 'messaging',
 '{"kubernetes_strategy":"statefulset","expose_udp":false,"health_check_path":"/api/healthchecks/node","health_check_port":15672}',
 '{"memory_mb":512,"cpu_millicores":500,"disk_gb":5}',
 '[{"name":"amqp","container":5672,"protocol":"tcp","expose":false,"label":"AMQP"},{"name":"management","container":15672,"protocol":"tcp","expose":true,"label":"Management UI"}]',
 '[{"key":"RABBITMQ_DEFAULT_USER","label":"Username","type":"string","default":"kleff","required":true},{"key":"RABBITMQ_DEFAULT_PASS","label":"Password","type":"secret","required":true,"auto_generate":true,"auto_generate_length":16},{"key":"RABBITMQ_DEFAULT_VHOST","label":"Default VHost","type":"string","default":"/","required":false}]',
 '[{"key":"ADDRESS","description":"AMQP host:port.","value_template":"{{ .ContainerName }}:5672"},{"key":"MANAGEMENT_URL","description":"Management UI URL.","value_template":"http://{{ .ContainerName }}:15672"}]',
 '{}'
),

-- Nginx
('nginx', 'nginx', 'Nginx', 'Nginx HTTP server and reverse proxy.',
 'nginx:alpine', '1.0.0', true, 'web',
 '{"kubernetes_strategy":"","expose_udp":false,"health_check_path":"/","health_check_port":80}',
 '{"memory_mb":128,"cpu_millicores":250,"disk_gb":5}',
 '[{"name":"http","container":80,"protocol":"tcp","expose":true,"label":"HTTP"},{"name":"https","container":443,"protocol":"tcp","expose":true,"label":"HTTPS"}]',
 '[{"key":"NGINX_HOST","label":"Server Name","type":"string","default":"localhost","required":false},{"key":"NGINX_PORT","label":"HTTP Port","type":"number","default":80,"required":false}]',
 '[{"key":"ADDRESS","description":"HTTP address.","value_template":"http://{{ .ContainerName }}:80"}]',
 '{}'
),

-- Caddy
('caddy', 'caddy', 'Caddy', 'Caddy web server with automatic HTTPS.',
 'caddy:alpine', '1.0.0', true, 'web',
 '{"kubernetes_strategy":"","expose_udp":false,"health_check_path":"/","health_check_port":80}',
 '{"memory_mb":128,"cpu_millicores":250,"disk_gb":5}',
 '[{"name":"http","container":80,"protocol":"tcp","expose":true,"label":"HTTP"},{"name":"https","container":443,"protocol":"tcp","expose":true,"label":"HTTPS"},{"name":"admin","container":2019,"protocol":"tcp","expose":false,"label":"Admin API"}]',
 '[{"key":"CADDY_DOMAIN","label":"Domain","description":"Domain for automatic HTTPS. Leave blank for local use.","type":"string","required":false}]',
 '[{"key":"ADDRESS","description":"HTTP address.","value_template":"http://{{ .ContainerName }}:80"}]',
 '{}'
)

ON CONFLICT (id) DO NOTHING;
