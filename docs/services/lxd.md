# LXD – Kiến trúc, thành phần và cách hoạt động

## 1. LXD là gì?

**LXD** là một **system container & VM manager** do **Canonical** phát triển, dùng để:

- Quản lý **Linux containers (LXC)** và **Virtual Machines (KVM)**
- Cung cấp trải nghiệm gần giống cloud (API-driven)
- Tích hợp **networking (OVN)** và **storage (Ceph)**
- Phù hợp cho **homelab, private cloud, MicroCloud**

> LXD không chỉ là container runtime.
> Nó là một **host-level hypervisor manager**.

---

## 2. LXD giải quyết vấn đề gì?

So với Docker / libvirt thuần:

- Docker: tập trung application container
- Libvirt: quản lý VM nhưng thiếu API cloud-like

➡ **LXD cung cấp:**

- Unified API cho container + VM
- Networking & storage tích hợp
- Lifecycle management (init, snapshot, migrate)

---

## 3. Kiến trúc tổng thể

```
+----------------------+
|   LXD Client (CLI)   |
|   lxc / REST API    |
+----------+-----------+
           |
           v
+----------------------+
|     LXD Daemon       |
|     (lxd)            |
+----------+-----------+
           |
  +--------+--------+--------+
  |                 |        |
  v                 v        v
Storage          Network    Compute
(ZFS/Ceph)       (OVN)     (LXC/KVM)
```

---

## 4. Các thành phần chính của LXD

### 4.1 LXD Daemon (`lxd`)

**Vai trò**

- REST API server
- Orchestrator cho container & VM
- Giao tiếp kernel (cgroups, namespaces, KVM)

---

### 4.2 LXC (System Container)

- Container cấp OS (init system đầy đủ)
- Gần như VM nhưng nhẹ hơn
- Chạy trực tiếp trên kernel host

---

### 4.3 Virtual Machine (KVM)

- Dùng QEMU + KVM
- Kernel riêng
- Isolation mạnh hơn container

---

### 4.4 Image server

- Quản lý OS images
- Hỗ trợ image remote (images.linuxcontainers.org)

---

### 4.5 Storage backend

Hỗ trợ:

- ZFS
- Ceph (RBD)
- LVM
- Btrfs
- Directory

---

### 4.6 Network backend

Hỗ trợ:

- Bridge
- OVN (SDN)
- Macvlan
- SR-IOV

---

## 5. Quy trình init LXD

```bash
lxd init
```

Quá trình:

1. Khởi tạo daemon
2. Cấu hình storage pool
3. Cấu hình network bridge
4. Enable API & security

➡ LXD sẵn sàng quản lý workload

---

## 6. LXD clustering

```bash
lxc cluster enable node-1
lxc cluster add node-2
```

- Shared database (dqlite)
- Leader election
- VM/container migrate

---

## 7. Binding Network

### 7.1 Bridge network

```bash
lxc network create lxdbr0
```

- NAT đơn giản
- Phù hợp dev/homelab nhỏ

---

### 7.2 OVN network (MicroOVN)

```bash
lxc network create ovn-net --type=ovn
```

- SDN
- Multi-node
- HA

---

## 8. Binding Storage

### 8.1 Local storage

```bash
lxc storage create local zfs
```

---

### 8.2 Ceph storage (MicroCeph)

```bash
lxc storage create ceph-pool ceph
```

- Distributed block storage
- VM migrate không downtime

---

## 9. Launch workload

### 9.1 Container

```bash
lxc launch ubuntu:22.04 c1
```

---

### 9.2 Virtual Machine

```bash
lxc launch ubuntu:22.04 vm1 --vm
```

---

## 10. Snapshot & Migration

```bash
lxc snapshot vm1 snap1
lxc move vm1 node-2
```

---

## 11. Security

- AppArmor
- Seccomp
- cgroups v2
- UID/GID mapping

---

## 12. Tính năng nổi bật

- Container + VM chung API
- Live migration
- Snapshot nhanh
- Network & storage tích hợp
- Cloud-like UX

---

## 13. LXD không làm gì

- Không phải Kubernetes
- Không phải Docker runtime
- Không orchestration app

---

## 14. Tóm tắt

**LXD**:

- Là hypervisor manager cấp OS
- Cung cấp nền tảng cho MicroCloud
- Kết hợp hoàn hảo với MicroOVN & MicroCeph

> \*LXD quản lý compute → MicroOVN quản lý network → MicroCeph quản lý stor
