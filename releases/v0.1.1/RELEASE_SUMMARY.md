# Tóm tắt phiên bản v0.1.1

## Tổng quan nhanh
Phiên bản đơn giản hóa API để hợp lý hóa truy cập dịch vụ thông qua giao diện manager thống nhất, loại bỏ mẫu `*.manager` cho trải nghiệm phát triển tốt hơn.

## Điểm nổi bật chính
- 🎉 **API đơn giản**: Truy cập dịch vụ thống nhất thông qua giao diện manager loại bỏ nhiều đăng ký dịch vụ
- 🚀 **DX tốt hơn**: Thiết kế API trực quan hơn với điểm truy cập duy nhất qua `container.MustMake("queue")`
- 🔧 **Trải nghiệm phát triển**: Code sạch hơn với các phương thức `manager.Client()` và `manager.Server()`

## Thống kê
- **Issues đã đóng**: 1 (Vấn đề nhất quán API)
- **Pull Requests đã merge**: 2
- **Contributors mới**: 0
- **Files đã thay đổi**: 5 (provider.go, README.md, docs/, tests)
- **Dòng đã thêm**: 20
- **Dòng đã xóa**: 20

## Tác động
Phiên bản này cải thiện trải nghiệm phát triển bằng cách cung cấp API sạch hơn, nhất quán hơn để truy cập các dịch vụ queue. Người dùng hiện tại sẽ cần cập nhật mẫu truy cập dịch vụ nhưng sẽ có được giao diện trực quan hơn.

## Bước tiếp theo
- Tính năng giám sát và metrics nâng cao
- WebUI dashboard để quản lý queue
- Tối ưu hóa hiệu suất cho các tình huống throughput cao

---
**Ghi chú phiên bản đầy đủ**: [RELEASE_NOTES.md](./RELEASE_NOTES.md)  
**Hướng dẫn Migration**: [MIGRATION.md](./MIGRATION.md)  
**Ngày phát hành**: 2025-06-05
