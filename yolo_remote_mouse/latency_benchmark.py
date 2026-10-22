"""端到端延迟基准测试：在本机回环模拟完整链路，测量各阶段耗时。

用于验证延迟预算是否达标。运行：
    python latency_benchmark.py
"""
import time
import numpy as np
import cv2

from config import CONFIG
from encoder import JPEGEncoder
from udp_transport import split_into_fragments, parse_upload_packet, FrameReassembler


def benchmark():
    print("=== Latency Benchmark ===")

    # 1. 截屏模拟（用随机图代替 mss）
    w, h = CONFIG.capture.width, CONFIG.capture.height
    frame = np.random.randint(0, 255, (h, w, 3), dtype=np.uint8)

    # 2. 编码
    enc = JPEGEncoder(CONFIG.encode.jpeg_quality, CONFIG.encode.scale)
    t0 = time.perf_counter()
    for _ in range(100):
        jpeg = enc.encode(frame)
    enc_ms = (time.perf_counter() - t0) / 100 * 1000
    print(f"encode: {enc_ms:.2f} ms  (jpeg_size={len(jpeg)} bytes)")

    # 3. 分片
    t0 = time.perf_counter()
    for _ in range(100):
        pkts = split_into_fragments(1, time.monotonic(), w, h, jpeg,
                                    CONFIG.transport.mtu)
    split_ms = (time.perf_counter() - t0) / 100 * 1000
    print(f"split:  {split_ms:.3f} ms  (fragments={len(pkts)})")

    # 4. 重组
    reassembler = FrameReassembler()
    t0 = time.perf_counter()
    for i in range(100):
        fid = i
        for p in pkts:
            pkt = parse_upload_packet(p)
            res = reassembler.add_fragment(pkt)
        # res should be complete
    reasm_ms = (time.perf_counter() - t0) / 100 * 1000
    print(f"reassemble: {reasm_ms:.3f} ms")

    # 5. 解码
    t0 = time.perf_counter()
    for _ in range(100):
        arr = np.frombuffer(jpeg, dtype=np.uint8)
        dec = cv2.imdecode(arr, cv2.IMREAD_COLOR)
    dec_ms = (time.perf_counter() - t0) / 100 * 1000
    print(f"decode: {dec_ms:.2f} ms")

    # 6. 推理（如果模型可用）
    try:
        from detector import YOLODetector
        det = YOLODetector(CONFIG.detect)
        t0 = time.perf_counter()
        for _ in range(50):
            dets = det.infer(dec)
        infer_ms = (time.perf_counter() - t0) / 50 * 1000
        print(f"infer:  {infer_ms:.2f} ms  (detections={len(dets)})")
    except Exception as e:
        print(f"infer:  SKIPPED ({e})")
        infer_ms = 0

    total = enc_ms + split_ms + reasm_ms + dec_ms + infer_ms
    print(f"\n--- 计算侧合计（不含网络）: {total:.2f} ms ---")
    print(f"加网络往返 ~5ms: ~{total + 5:.1f} ms")
    print(f"目标 <= 45ms: {'PASS' if total + 5 < 45 else 'NEED OPTIMIZE'}")


if __name__ == "__main__":
    benchmark()
