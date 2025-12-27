# MicroCeph – Kiến trúc, thành phần và cách hoạt động

## 1. MicroCeph là gì?

**MicroCeph** là công cụ do **Canonical** phát triển để:

- Triển khai **Ceph cluster** một cách **đơn giản, HA**
- Quản lý vòng đời Ceph (init / join / scale)
- Tích hợp **Ceph trực tiếp với LXD**
- Cung cấp **distributed storage** cho VM và container

> MicroCeph **không phải** là Ceph.
> Nó đóng vai trò **installer + orchestrator + glue** giữa **Ceph ↔ LXD**.

---

## 2. Vấn đề MicroCeph giải quyết

Triển khai Ceph truyền thống rất phức tạp:

- Phải tự cài MON, MGR, OSD
- Phải cấu hình CRUSH map
- Phải lo quorum, HA, recovery
- Phải tích hợp thủ công với LXD / OpenStack

➡ **MicroCeph che giấu toàn bộ độ phức tạp đó**, phù hợp cho:

- Homelab
- Private cloud nhỏ
- MicroCloud

---

## 3. Kiến trúc tổng thể

```
+------------------+
|       LXD        |
|  (VM / Volume)  |
+--------+---------+
         |
         v
+------------------+
|    MicroCeph     |
| (Orchestration) |
+--------+---------+
         |
         v
+----------------------------------+
|        Ceph Cluster               |
|  - MON (Monitor)                  |
|  - MGR (Manager)                  |
|  - OSD (Object Storage Daemon)    |
+--------+--------------------------+
         |
         v
     Disks / NVMe / SSD / HDD
```

---

## 4. Các thành phần chính của Ceph (được MicroCeph quản lý)

### 4.1 MON – Ceph Monitor

**Vai trò**

- Giữ **cluster map**
- Quản lý quorum
- Xác định trạng thái cluster

> MON = control plane của Ceph

---

### 4.2 MGR – Ceph Manager

**Vai trò**

- Thu thập metrics
- Quản lý module (dashboard, balancer)
- Điều phối background task

---

### 4.3 OSD – Object Storage Daemon

**Vai trò**

- Lưu dữ liệu thật trên disk
- Replication / recovery
- Tham gia CRUSH algorithm

> 1 disk = 1 OSD (thường)

---

### 4.4 CRUSH Map

- Thuật toán phân phối dữ liệu
- Không cần metadata server
- Quyết định object nằm ở node/disk nào

---

## 5. MicroCeph daemon

**Vai trò**

- Cài đặt Ceph packages
- Bootstrap cluster
- Join / leave node
- Quản lý disk → OSD
- Kết nối với LXD storage backend

---

## 6. Quy trình hoạt động của MicroCeph

### 6.1 Khởi tạo MicroCeph cluster

```bash
microceph init
```

MicroCeph sẽ:

1. Cài Ceph packages
2. Khởi tạo MON đầu tiên
3. Khởi tạo MGR
4. Tạo cluster keyring

➡ Ceph cluster sẵn sàng nhận node mới

---

### 6.2 Node tham gia cluster

```bash
microceph join
```

MicroCeph:

1. Thêm MON (nếu cần HA)
2. Thêm MGR (active/standby)
3. Đồng bộ cluster config

➡ Cluster đạt HA quorum

---

### 6.3 Thêm disk làm OSD

```bash
microceph disk add /dev/sdb
```

MicroCeph:

- Zap disk
- Tạo OSD
- Gắn vào CRUSH map

➡ Storage capacity tăng ngay lập tức

---

## 7. Binding MicroCeph với LXD

### 7.1 Tạo storage pool trong LXD

```bash
lxc storage create ceph-pool ceph
```

MicroCeph:

- Cấp quyền Ceph user cho LXD
- Map pool vào Ceph RBD

---

### 7.2 Gắn volume vào VM

```bash
lxc launch ubuntu:22.04 vm1 -s ceph-pool
```

Hoặc attach disk:

```bash
lxc storage volume attach ceph-pool vol1 vm1
```

➡ VM sử dụng **Ceph RBD** làm block device

---

## 8. Data flow (VM → Disk)

```
VM
 ↓
RBD block device
 ↓
Ceph client
 ↓
CRUSH algorithm
 ↓
OSD replication
 ↓
Physical disks
```

---

## 9. Tính năng nổi bật của MicroCeph

- HA mặc định (MON quorum)
- Scale ngang dễ dàng
- Tự động replication
- Self-healing
- Tích hợp chặt với LXD

---

## 10. MicroCeph không làm gì

- Không thay Ceph
- Không thay filesystem trong VM
- Không quản lý application data

---

## 11. So sánh nhanh

| Thành phần | Vai trò                |
| ---------- | ---------------------- |
| LXD        | VM / container manager |
| MicroCeph  | Ceph orchestrator      |
| Ceph       | Distributed storage    |
| OSD        | Disk daemon            |

---

## 12. Tóm tắt

**MicroCeph**:

- Đơn giản hóa Ceph
- Che giấu phức tạp của distributed storage
- Cung cấp block storage HA cho VM/container

> _LXD cần storage → MicroCeph hiện thực bằng Ceph_
