# M2M

### 🏛️ Bức tranh tổng thể về Kiến trúc Xác thực (Auth Architecture)

Chúng ta phân chia mọi hành động trong hệ thống ra làm 2 nhóm chính dựa trên sự xuất hiện của **Con người (User)** [1]:

#### Nhóm 1: Các hành động có Con người (User-initiated Actions)

- **Ví dụ:** User bấm nút đặt hàng, xem profile, chuyển tiền [1].
- **Giải pháp chuẩn mực:** **mTLS (Tầng mạng) + JWT Propagation (Tầng code)** [1, 2].
  - _mTLS:_ Để các service tin tưởng lẫn nhau ở tầng hạ tầng (Chỉ cho phép Service A gọi Service B) [1, 2].
  - _JWT:_ Chuyển tiếp (forward) token của User đi xuyên suốt các service để biết **AI** đang thực hiện hành động và ghi log chính xác [1].
  - ➔ Ở nhóm này, bạn **không cần** dùng đến Client Credentials nữa [1].

#### Nhóm 2: Các hành động KHÔNG có Con người (System-initiated Actions - M2M)

- **M2M Nội bộ (Background Jobs / ETL / Cron Jobs):**
  - _Ví dụ:_ Máy chủ tự động chạy ETL đồng bộ dữ liệu lúc 12h đêm, tự dọn dẹp database [1].
  - _Giải pháp:_ Dùng **mTLS** hoặc phân vùng mạng nội bộ (Internal Network) là đủ bảo mật [1].
- **M2M Đối ngoại (B2B Integrations - Tích hợp hệ thống với đối tác bên ngoài):**
  - _Ví dụ:_ Hệ thống của bạn là cổng thanh toán. Bạn mở cổng API cho máy chủ của các đối tác (như Shopee, Tiki) tự động gọi sang hệ thống của bạn để kiểm tra trạng thái đơn hàng [1].
  - _Tại sao là M2M?_ Vì máy chủ của đối tác tự động gọi sang máy chủ của bạn qua API, hoàn toàn không có "User" nào của hệ thống bạn đăng nhập ở đây cả [1].
  - _Giải pháp:_ Đây chính là nơi **OAuth 2.0 Client Credentials Grant** phát huy sức mạnh tối đa [1]. Bạn cấp cho đối tác cặp `client_id` và `client_secret` để máy chủ của họ tự động lấy Access Token gọi API của bạn [1].

---

### 🏆 Đúc kết cuối cùng:

- **Có User:** Dùng **mTLS + JWT** [1].
- **Không có User (Nội bộ):** Dùng **mTLS / Background Jobs** [1].
- **Không có User (Đối ngoại / Đối tác / Webhooks):** Dùng **OAuth 2.0 Client Credentials** [1].
