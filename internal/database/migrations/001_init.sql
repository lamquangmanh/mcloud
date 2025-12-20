-- 1. Cluster information
CREATE TABLE IF NOT EXISTS clusters (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  state TEXT NOT NULL CHECK(state IN ('init', 'active', 'degraded')),
  
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  create_user_id TEXT,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_user_id TEXT
);

-- 2. Node information
CREATE TABLE IF NOT EXISTS nodes (
  id TEXT PRIMARY KEY,
  cluster_id TEXT NOT NULL,
  hostname TEXT NOT NULL,
  ip TEXT NOT NULL,
  role TEXT NOT NULL CHECK(role IN ('leader', 'worker')),
  status TEXT NOT NULL CHECK(status IN ('joining', 'online', 'offline')),
  joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  last_heartbeat DATETIME,

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  create_user_id TEXT,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_user_id TEXT,

  FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE,
  UNIQUE (cluster_id, hostname),
  UNIQUE (cluster_id, ip)
);
CREATE INDEX IF NOT EXISTS idx_nodes_cluster_id ON nodes(cluster_id);
CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status);

-- 3. Bootstrap tokens
CREATE TABLE IF NOT EXISTS bootstrap_tokens (
  token TEXT PRIMARY KEY,
  cluster_id TEXT NOT NULL,
  expires_at DATETIME NOT NULL,
  used INTEGER DEFAULT 0,

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  create_user_id TEXT,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_user_id TEXT,


  FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_tokens_expires_at ON bootstrap_tokens(expires_at);

-- 4. Certificate authorities
CREATE TABLE IF NOT EXISTS certificate_authorities (
  id TEXT PRIMARY KEY,
  cluster_id TEXT NOT NULL,
  cert_pem TEXT NOT NULL,
  key_pem TEXT NOT NULL,

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  create_user_id TEXT,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_user_id TEXT,

  FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE
);

-- 5. Node certificates
CREATE TABLE IF NOT EXISTS node_certificates (
  id TEXT PRIMARY KEY,
  node_id TEXT NOT NULL,
  cert_pem TEXT NOT NULL,
  issued_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME NOT NULL,

  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  create_user_id TEXT,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_user_id TEXT,

  FOREIGN KEY (node_id) REFERENCES nodes(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_node_certs_node_id ON node_certificates(node_id);

-- 6. Node health metrics
CREATE TABLE IF NOT EXISTS node_health (
  node_id TEXT PRIMARY KEY,
  cpu_usage REAL,
  memory_usage REAL,
  disk_usage REAL,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY (node_id) REFERENCES nodes(id) ON DELETE CASCADE
);

-- 7. Workloads
CREATE TABLE IF NOT EXISTS workloads (
  id TEXT PRIMARY KEY,
  cluster_id TEXT NOT NULL,
  node_id TEXT,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK(kind IN ('vm', 'container', 'job')),
  status TEXT NOT NULL CHECK(status IN ('pending', 'running', 'stopped', 'failed')),
  
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  create_user_id TEXT,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_user_id TEXT,

  FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE,
  FOREIGN KEY (node_id) REFERENCES nodes(id)
);
CREATE INDEX IF NOT EXISTS idx_workloads_cluster_id ON workloads(cluster_id);
CREATE INDEX IF NOT EXISTS idx_workloads_node_id ON workloads(node_id);

-- 8. Events
CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  cluster_id TEXT,
  node_id TEXT,
  type TEXT NOT NULL,
  message TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY (cluster_id) REFERENCES clusters(id),
  FOREIGN KEY (node_id) REFERENCES nodes(id)
);
CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at);

-- 9. Key-value store (config / state )
CREATE TABLE IF NOT EXISTS kv_store (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

