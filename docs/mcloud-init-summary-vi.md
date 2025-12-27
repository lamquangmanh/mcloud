# MicroCloud init – Phân tích chi tiết & Gợi ý build MCloud init

Tài liệu này gồm 2 phần:

1. **MicroCloud `init` làm gì (chi tiết theo từng phase)**
2. **Gợi ý các step tương đương để bạn build `mcloud init` (learning-friendly)**

Mục tiêu: giúp bạn **hiểu đúng kiến trúc**, không build lại MicroCloud, mà **xây MCloud như một control-plane orchestration layer**.

---

## Phần 1 – MicroCloud `init` chi tiết theo phase

### Tổng quan

```
microcloud init
│
├─ Phase 0: Preflight & Validation
├─ Phase 1: Bootstrap LXD Cluster
├─ Phase 2: Bootstrap Networking (MicroOVN)
├─ Phase 3: Bootstrap Storage (MicroCeph)
├─ Phase 4: Attach Network + Storage to LXD
└─ Phase 5: Finalize & Expose Cluster
```

MicroCloud **không thay thế** LXD / OVN / Ceph.
Nó chỉ **orchestrate** và **đảm bảo thứ tự, version, state** giữa các thành phần.

---

### Phase 0 – Preflight & Validation

MicroCloud kiểm tra:

- OS (Ubuntu LTS)
- Kernel features (vxlan, bridge, nftables)
- Network interface khả dụng
- Disk trống cho Ceph OSD
- Snap đã cài:

  - lxd
  - microovn
  - microceph

> Fail ở phase này → `microcloud init` dừng ngay

---

### Phase 1 – Bootstrap LXD Cluster

Đây là phần tương ứng với `lxd.InitCluster()`.

MicroCloud:

- Start LXD daemon
- Generate cluster certificate
- Init **LXD cluster database (dqlite)**
- Node hiện tại trở thành **cluster leader**

Conceptually:

```
lxd init --cluster
```

Sau phase này:

- Có LXD cluster
- Chưa network
- Chưa storage
- Chưa usable

---

### Phase 2 – Bootstrap Networking (MicroOVN)

MicroCloud gọi **MicroOVN** để:

- Bootstrap OVN Northbound / Southbound DB
- Deploy OVN controller trên node
- Create logical switch + router
- Bind uplink NIC

Sau đó:

- Register OVN network vào LXD

Kết quả:

```
lxc network list
# thấy network type = ovn
```

---

### Phase 3 – Bootstrap Storage (MicroCeph)

MicroCloud gọi **MicroCeph** để:

- Bootstrap Ceph MON
- Detect raw disks
- Create Ceph OSD
- Create Ceph pool (RBD)

Sau đó:

- Register Ceph pool vào LXD

Kết quả:

```
lxc storage list
# backend = ceph
```

---

### Phase 4 – Attach Network + Storage to LXD

Đây là orchestration logic chính.

MicroCloud:

- Update LXD `default` profile
- Attach:

  - OVN network → eth0
  - Ceph pool → root disk

Conceptually:

```
lxc profile device add default eth0 nic network=ovn-net
lxc profile device add default root disk pool=ceph-pool
```

Sau phase này:

- VM / container có IP
- Có persistent storage
- HA-ready

---

### Phase 5 – Finalize & Expose Cluster

MicroCloud:

- Validate toàn bộ stack
- Save cluster metadata
- Expose join token

Output:

```
MicroCloud initialized successfully
Run `microcloud join` on other nodes
```

---

# MCloud Init – Kiến trúc chuẩn & các bước triển khai (Learning nhưng chạy THẬT)

Tài liệu này **tổng hợp kiến trúc và các bước `mcloud init`** được đề xuất để:

- Học **control plane / distributed system** đúng bản chất
- Chạy **infra thật** (LXD, MicroCeph, MicroOVN)
- **Không chắp vá**, không fake
- Có thể mở rộng lên HA sau này

---

## 1. Mục tiêu thiết kế

- Build **orchestration layer**, không reimplement infra
- Có **leader control plane** rõ ràng
- Có **auth + trust** ngay từ đầu
- State nhất quán, không hack sync

Không nhằm clone MicroCloud, mà là:

> _MicroCloud-inspired learning control plane_

---

## 2. Kiến trúc tổng thể

```
+-----------------------------+
| Leader Node                 |
|                             |
|  mcloudd                    |
|  ├─ SQLite (cluster state)  |
|  ├─ CA + Cert Store         |
|  ├─ Join Token Store        |
|  ├─ LXD Orchestrator        |
|  ├─ MicroCeph Orchestrator  |
|  └─ MicroOVN Orchestrator   |
|                             |
+--------------▲--------------+
               │ mTLS
+--------------┴--------------+
| Worker Node                  |
|                              |
|  mcloud-agent                |
|  ├─ Join logic               |
|  ├─ mTLS client              |
|  └─ Infra hooks              |
|     ├─ LXD                   |
|     ├─ MicroCeph             |
|     └─ MicroOVN              |
|                              |
+------------------------------+
```

### Nguyên tắc cốt lõi

- **Single leader control plane**
- SQLite **chỉ chạy trên leader**
- Worker **không giữ cluster DB**
- Giao tiếp **mTLS-only** sau bootstrap

---

## 3. Phân tách trách nhiệm

### mcloudd (leader)

- Quyền quyết định cluster state
- Ghi SQLite
- Sinh cert, token
- Điều phối infra theo thứ tự

### mcloud-agent (worker)

- Không quyết định state
- Thực thi lệnh local (LXD/Ceph/OVN)
- Báo kết quả về leader

---

## 4. State Management (SQLite – chuẩn, không chắp vá)

- SQLite **leader-only**
- Tables tối thiểu:

  - `cluster`
  - `nodes`
  - `join_tokens`
  - `certs`

### Lý do không dùng SQLite distributed

- SQLite không multi-writer
- Đồng bộ DB = hack
- Leader-based = đúng control plane

---

## 5. Auth & Trust Model (BẮT BUỘC)

### Bootstrap flow

1. Leader tạo CA
2. Leader tạo join token (TTL)
3. Worker dùng token để join
4. Leader cấp client cert
5. Từ đây → **mTLS only**

> Không auth = không phải cloud control plane

---

## 6. Các bước `mcloud init` (CHI TIẾT)

### Phase 0 – Preflight

- Load config
- Validate:

  - OS / permission
  - port
  - disk path
  - snap availability

---

### Phase 1 – Init Control Plane

- Generate cluster ID
- Init SQLite
- Mark node = leader
- Generate CA + server cert

---

### Phase 2 – Init LXD Cluster (REAL)

- Ensure LXD running
- Run LXD cluster init
- Verify cluster leader

> Đây là compute layer thật

---

### Phase 3 – Init MicroCeph (REAL)

- Bootstrap MicroCeph
- Detect raw disks
- Create OSD
- Create Ceph pool
- Register pool vào LXD

> Đây là storage layer thật

---

### Phase 4 – Init MicroOVN (REAL)

- Bootstrap OVN DB
- Deploy OVN controller
- Create logical network
- Register OVN network vào LXD

> Networking overlay thật (không fake)

---

### Phase 5 – Attach Network + Storage

- Update LXD default profile
- Attach:

  - OVN network → eth0
  - Ceph pool → root disk

---

### Phase 6 – Finalize

- Save cluster metadata
- Generate join token
- Expose `mcloud join` instruction

---

## 7. `mcloud join` (tóm tắt để liên kết)

1. Worker có join token
2. Call leader `/join`
3. Verify token
4. Issue client cert
5. Join LXD cluster
6. Add Ceph OSD
7. Deploy OVN agent
8. Register node

---

## 8. Những gì CHỦ ĐÍCH bỏ qua (giai đoạn học)

- HA control plane (dqlite / raft)
- UI
- Auto-upgrade
- Multi-tenant

> Những thứ này để Phase sau

---

## 9. Roadmap mở rộng hợp lý

### Phase A – Core

- mcloud init
- mcloud join
- node / cluster list

### Phase B – Reliability

- health check
- node drain / leave

### Phase C – HA

- dqlite hoặc raft

### Phase D – Kubernetes (optional)

- MicroK8s inside LXD
- Ceph as PV

---

# Thiết kế các function bootstrap / init / join cho LXD, MicroOVN, MicroCeph (MCloud)

Tài liệu này **xác định rõ các function cần có** trong MCloud để:

- Bootstrap cluster
- Init node đầu tiên
- Join node mới

Mục tiêu:

- Dùng **infra thật** (LXD, MicroCeph, MicroOVN)
- Không chắp vá
- Có thể triển khai từng bước

---

## 1. Nguyên tắc thiết kế

1. **MCloud KHÔNG reimplement infra**
2. MCloud chỉ:

   - gọi CLI / API chính thức
   - kiểm tra trạng thái
   - rollback khi lỗi

3. Mỗi infra có 3 nhóm function:

   - Bootstrap (node đầu)
   - Join (node mới)
   - Validate / Status

---

## 2. Phân tầng module trong MCloud

```
internal/
├── infra/
│   ├── lxd/
│   │   ├── bootstrap.go
│   │   ├── join.go
│   │   └── status.go
│   ├── microceph/
│   │   ├── bootstrap.go
│   │   ├── join.go
│   │   └── status.go
│   └── microovn/
│       ├── bootstrap.go
│       ├── join.go
│       └── status.go
```

---

## 3. LXD – Function design

### 3.1 Bootstrap LXD cluster (node đầu tiên)

```go
func BootstrapCluster(cfg LXDConfig) error
```

**Nhiệm vụ**:

- Kiểm tra LXD daemon
- Init LXD cluster (leader)
- Verify cluster status

**Thực thi thực tế**:

- `lxd init --cluster`
- hoặc LXD REST API

---

### 3.2 Join LXD cluster (node mới)

```go
func JoinCluster(cfg LXDJoinConfig) error
```

**Nhiệm vụ**:

- Nhận join token / cert từ leader
- Join LXD cluster
- Verify node xuất hiện trong cluster

**Thực thi**:

- `lxd cluster add`
- `lxd init --cluster-join`

---

### 3.3 Status / Validate

```go
func Status() (LXDStatus, error)
```

---

## 4. MicroCeph – Function design

### 4.1 Bootstrap MicroCeph (node đầu)

```go
func Bootstrap(cfg CephConfig) error
```

**Nhiệm vụ**:

- Ensure microceph running
- Bootstrap Ceph MON
- Detect raw disks
- Create OSD
- Create default pool

**Thực thi**:

- `microceph init`
- `microceph disk add`

---

### 4.2 Join MicroCeph cluster (node mới)

```go
func Join(cfg CephJoinConfig) error
```

**Nhiệm vụ**:

- Join Ceph cluster
- Add OSD từ disk

**Thực thi**:

- `microceph join`
- `microceph disk add`

---

### 4.3 Register Ceph storage vào LXD

```go
func RegisterToLXD(pool string) error
```

---

## 5. MicroOVN – Function design

### 5.1 Bootstrap MicroOVN (node đầu)

```go
func Bootstrap(cfg OVNConfig) error
```

**Nhiệm vụ**:

- Bootstrap OVN Northbound / Southbound DB
- Start OVN controller
- Create logical network

**Thực thi**:

- `microovn init`

---

### 5.2 Join MicroOVN (node mới)

```go
func Join(cfg OVNJoinConfig) error
```

**Nhiệm vụ**:

- Join OVN cluster
- Start OVN controller

**Thực thi**:

- `microovn join`

---

### 5.3 Register OVN network vào LXD

```go
func RegisterToLXD(network string) error
```

---

## 6. Orchestration flow (mcloud init)

```text
mcloud init
│
├─ lxd.BootstrapCluster()
├─ microceph.Bootstrap()
├─ microovn.Bootstrap()
├─ microceph.RegisterToLXD()
├─ microovn.RegisterToLXD()
└─ finalize
```

---

## 7. Orchestration flow (mcloud join)

```text
mcloud join
│
├─ authenticate (token + mTLS)
├─ lxd.JoinCluster()
├─ microceph.Join()
├─ microovn.Join()
└─ register node
```

---

## 8. Vì sao thiết kế này là CHUẨN

- Mỗi infra có module riêng
- Không hard-code CLI rải rác
- Có thể replace CLI bằng API
- Giữ đúng abstraction boundary

---

## 9. Bước tiếp theo nên làm

1. Viết interface chung `InfraBootstrapper`
2. Implement LXD bootstrap thật (CLI wrapper)
3. Implement MicroCeph bootstrap thật
4. Implement MicroOVN bootstrap thật

---
