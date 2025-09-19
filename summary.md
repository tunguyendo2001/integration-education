## 📝 Vấn đề

* **Hiện trạng**: Nhà trường nhập điểm bằng tay vào hệ thống của Sở GD.

  * Quá tốn thời gian.
  * Auth của Sở chỉ sống 5 phút, dễ timeout → mất dữ liệu chưa “save”.
  * Hệ thống Sở chỉ mở khung thời gian nhập điểm (ví dụ 1 ngày/kỳ).
* **Mục tiêu**:

  * Cho phép nhà trường nhập điểm vào file Excel (xlsx).
  * Upload file qua API gateway (đã có).
  * Hệ thống tự động đồng bộ sang Sở GD qua API, đảm bảo dữ liệu đúng, retry được, và user biết trạng thái.

---

## ✅ Giải pháp tạm thời (proposed)

1. **Upload file**

   * Flask API riêng cho điểm nhận file Excel.
   * Lưu file vào S3.
   * Lưu metadata (file\_id, school\_id, semester, …) vào DB.
   * Push job ingestion vào Redis queue.

2. **Ingestion worker**

   * Lấy job ingestion → đọc file từ S3 → parse Excel.
   * Lưu từng record (có `file_id`) vào DB staging (`grades_staging`).
   * Update `files.status=PARSED_OK`.

3. **User action: Đẩy dữ liệu sang Sở**

   * Trên UI có nút "Đẩy dữ liệu".
   * Khi ấn → API check thời gian hợp lệ (có nằm trong khung nhập điểm của Sở không).
   * Nếu hợp lệ → push job sync vào Redis queue: `{file_id, class, subject}`.

4. **Sync worker**

   * Consume job sync.
   * Query dữ liệu từ `grades_staging` theo `file_id, class, subject`.
   * Batch insert sang API Sở GD.
   * Update trạng thái từng record (`SYNCED` hoặc `FAILED`).

5. **Retry & thông báo**

   * Network lỗi → job tự động retry qua DLQ.
   * Business lỗi (sai student\_id, subject không khớp) → mark `FAILED`, báo user.
   * Người dùng có thể tra cứu qua API `/grades/status/{file_id}`.

---

## ⚠️ Những vấn đề còn tồn tại

* **Phụ thuộc vào time window của Sở GD**

  * Nếu hệ thống check nhầm (server local vs server Sở lệch giờ) → có nguy cơ reject toàn bộ.
* **Xử lý lỗi từng phần**

  * Nếu một batch có 100 học sinh mà 5 người sai → tuỳ API Sở, có thể fail cả batch hoặc chỉ 5 người fail.
  * Cần định nghĩa rõ cơ chế retry: retry từng student hay retry cả batch.
* **Concurrency**

  * Nếu nhiều người cùng ấn "Đẩy dữ liệu" cho cùng 1 file/lớp/môn → có thể tạo duplicate jobs. Cần khoá logic (idempotency key).
* **Thông báo cho người dùng**

  * Hiện tại chỉ có status pull (user query). Chưa có push notification/email.
* **Scaling**

  * Redis queue ổn trong phạm vi 1 Sở, nhưng nếu số lượng trường nhiều (hàng ngàn file cùng lúc) có thể cần Kafka để scale và phân vùng theo `school_id`.
* **Data validation**

  * Mới validate format file khi upload, chưa có bước validate semantic (VD: student\_id có tồn tại bên Sở không, điểm hợp lệ 0–10).

---

👉 **Kết luận**:
Cách tiếp cận hiện tại **ổn làm MVP**:

* Đảm bảo dữ liệu không mất (S3 + staging DB).
* Async sync sang Sở GD, có retry.
* User kiểm soát được khi nào “đẩy dữ liệu”.

Nhưng để production thì còn cần giải quyết: **batch error handling, concurrency control, và thông báo lỗi chi tiết cho người dùng**.
