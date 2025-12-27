# MicroOVN – Kiến trúc, thành phần và cách hoạt động

## 1. MicroOVN là gì?

**MicroOVN** là một thành phần trong hệ sinh thái **MicroCloud (Canonical)**, dùng để:

- Triển khai **OVN (Open Virtual Network)** một cách **đơn giản, HA**
- Quản lý **OVN control plane** (Northbound / Southbound DB)
- Tích hợp **OVN trực tiếp với LXD**
- Cung cấp **SDN networking** cho VM và container

> MicroOVN **không phải** là OVN.
> Nó đóng vai trò **installer + orchestrator + glue** giữa **OVN ↔ LXD**.

---

## 2. Vấn đề MicroOVN giải quyết

Nếu triển khai OVN thủ công:

- Phải tự cài NB/SB DB
- Phải cấu hình RAFT HA
- Phải cấu hình Open vSwitch, tunnel, bridge
- Phải tự tích hợp với LXD

➡ **MicroOVN che giấu toàn bộ độ phức tạp này**.

---

## 3. Kiến trúc tổng thể

```
+------------------+
|       LXD        |
| (API / Network) |
+--------+---------+
         |
         v
+------------------+
|    MicroOVN      |
| (Orchestration) |
+--------+---------+
         |
         v
+---------------------------+
|   OVN Control Plane       |
|  - Northbound DB (NB)     |
|  - Southbound DB (SB)     |
+--------+------------------+
         |
         v
+---------------------------+
| ovn-controller (per node) |
| Open vSwitch (OVS)        |
+--------+------------------+
         |
         v
       NIC
```

---

## 4. Các thành phần chính

### 4.1 OVN Northbound Database (NB DB)

**Vai trò**

- Lưu **ý định mạng (desired state)**
- Không quan tâm node vật lý

**Chứa**

- Logical Switch
- Logical Router
- Logical Port
- ACL (firewall)
- NAT, DHCP, LoadBalancer

➡ Trả lời câu hỏi: _"Mạng logic nên trông như thế nào?"_

---

### 4.2 OVN Southbound Database (SB DB)

**Vai trò**

- Dịch logic → vật lý
- Biết workload chạy trên node nào

**Chứa**

- Chassis (node)
- Port binding (VM ↔ node)
- Tunnel (Geneve)
- Datapath / flow

➡ Trả lời câu hỏi: _"Packet này đi qua node nào?"_

---

### 4.3 ovn-northd

- Đồng bộ dữ liệu từ **NB → SB**
- Không xử lý packet

---

### 4.4 ovn-controller (mỗi node)

- Theo dõi SB DB
- Cấu hình:

  - Open vSwitch
  - OpenFlow rules
  - Geneve tunnel

➡ Là **executor** thực sự của OVN.

---

### 4.5 Open vSwitch (OVS)

- Dataplane thực
- Forward packet
- Thực thi flow rules

---

### 4.6 MicroOVN daemon

**Vai trò**

- Cài đặt OVN
- Khởi tạo cluster OVN DB (RAFT)
- Join / leave node
- Kết nối với LXD

---

## 5. Quy trình hoạt động của MicroOVN

### 5.1 Khởi tạo MicroOVN

```bash
microovn init
```

Thực hiện:

1. Cài OVN packages
2. Tạo OVN NB DB & SB DB
3. Thiết lập RAFT HA
4. Chạy `ovn-northd`

➡ OVN control plane sẵn sàng.

---

### 5.2 Node tham gia cluster

```bash
microovn join
```

Thực hiện:

1. Cài `ovn-controller`
2. Cài Open vSwitch
3. Đăng ký node vào SB DB (chassis)

➡ Node trở thành dataplane.

---

## 6. Binding MicroOVN với LXD

### 6.1 Tạo OVN network

```bash
lxc network create ovn-net --type=ovn
```

MicroOVN:

- Ghi Logical Switch vào NB DB
- northd dịch sang SB DB

---

### 6.2 Cấu hình router & NAT

```bash
lxc network set ovn-net ipv4.address=10.0.0.1/24
lxc network set ovn-net ipv4.nat=true
```

MicroOVN:

- Tạo Logical Router
- Attach Switch ↔ Router
- Bật NAT

---

## 7. Binding OVN vào VM / Container

```bash
lxc launch ubuntu:22.04 vm1 --network ovn-net
```

MicroOVN thực hiện:

1. Tạo Logical Port (NB DB)
2. Bind port vào node (SB DB)
3. ovn-controller tạo veth + OVS port
4. VM nhận IP từ OVN DHCP

---

## 8. Packet flow

```
VM1
 ↓
veth
 ↓
OVS (node-1)
 ↓
Geneve tunnel
 ↓
OVS (node-2)
 ↓
veth
 ↓
VM2
```

---

## 9. Tính năng nổi bật

- HA mặc định (RAFT)
- SDN đầy đủ (Switch, Router, ACL, NAT)
- Tích hợp chặt với LXD
- Multi-node transparent

---

## 10. MicroOVN không làm gì

- Không thay LXD
- Không xử lý packet
- Không thay OVS
- Không làm ingress controller

---

## 11. Tóm tắt

**MicroOVN**:

- Orchestrate OVN
- Hiện thực hóa intent từ LXD
- Cung cấp SDN networking cho VM/Container

> _LXD khai báo mong muốn → MicroOVN triển khai bằng OVN_
