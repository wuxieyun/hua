"""客户端主程序：截屏-编码-上传 + 接收结果-鼠标控制。

双线程共享同一个 UDP socket（发送端口 = 接收端口，server 回传才能到）：
- send_thread：按 target_fps 截屏、编码、上传；
- recv_thread：阻塞收结果，收到即应用到鼠标（选置信度最高的检测）。
"""
import socket
import threading
import time

from config import CONFIG
from capture import ScreenCapture
from encoder import JPEGEncoder
from mouse_controller import MouseController
from protocol import ResultPacket
from udp_transport import create_socket, split_into_fragments


class Client:
    def __init__(self):
        self.cfg = CONFIG
        self.capture = ScreenCapture(CONFIG.capture.width,
                                     CONFIG.capture.height)
        self.encoder = JPEGEncoder(CONFIG.encode.jpeg_quality,
                                   CONFIG.encode.scale)
        region = (self.capture.region["left"], self.capture.region["top"],
                  self.capture.region["width"], self.capture.region["height"])
        self.mouse = MouseController(CONFIG.mouse, region)
        # 单一 socket：既发送也接收
        self.sock = create_socket()
        self.target = (CONFIG.transport.client_host,
                       CONFIG.transport.server_port)
        self.frame_id = 0
        self.running = True
        # frame_id -> capture_ts，用于真实 E2E 延迟测量
        self._capture_ts = {}
        self._ts_lock = threading.Lock()
        self.stats = {"sent": 0, "recv": 0, "e2e_ms": 0.0,
                      "last_print": time.monotonic()}

    def _send_loop(self):
        interval = 1.0 / self.cfg.capture.target_fps
        while self.running:
            t_start = time.monotonic()
            frame = self.capture.grab()
            capture_ts = time.monotonic()
            jpeg = self.encoder.encode(frame)
            if jpeg:
                self.frame_id += 1
                h, w = frame.shape[:2]
                fid = self.frame_id
                with self._ts_lock:
                    self._capture_ts[fid] = capture_ts
                    # 清理过老的记录
                    if len(self._capture_ts) > 120:
                        old = min(self._capture_ts)
                        del self._capture_ts[old]
                packets = split_into_fragments(
                    fid, capture_ts, w, h, jpeg,
                    CONFIG.transport.mtu)
                for pkt in packets:
                    try:
                        self.sock.sendto(pkt, self.target)
                    except OSError:
                        break
                self.stats["sent"] += 1
            elapsed = time.monotonic() - t_start
            sleep = interval - elapsed
            if sleep > 0:
                time.sleep(sleep)

    def _recv_loop(self):
        while self.running:
            try:
                data, _ = self.sock.recvfrom(65535)
            except socket.timeout:
                continue
            except OSError:
                break
            result = ResultPacket.deserialize(data)
            if result is None:
                continue
            self.stats["recv"] += 1
            now = time.monotonic()
            # 真实 E2E：截屏时刻 -> 收到结果
            with self._ts_lock:
                cap_ts = self._capture_ts.pop(result.frame_id, None)
            if cap_ts is not None:
                self.stats["e2e_ms"] = (now - cap_ts) * 1000
            # 选置信度最高的检测
            if result.detections:
                best = max(result.detections, key=lambda d: d.conf)
                self.mouse.apply_detection(best.x, best.y)

    def run(self):
        print(f"[client] sending to {self.cfg.transport.client_host}:"
              f"{self.cfg.transport.server_port}")
        t_send = threading.Thread(target=self._send_loop, daemon=True)
        t_recv = threading.Thread(target=self._recv_loop, daemon=True)
        t_send.start()
        t_recv.start()
        try:
            while self.running:
                time.sleep(self.cfg.stats_interval)
                now = time.monotonic()
                dt = now - self.stats["last_print"]
                if dt > 0:
                    print(f"[client] send_fps={self.stats['sent']/dt:.0f} "
                          f"recv_fps={self.stats['recv']/dt:.0f} "
                          f"e2e_ms={self.stats['e2e_ms']:.1f}")
                self.stats["sent"] = 0
                self.stats["recv"] = 0
                self.stats["last_print"] = now
        except KeyboardInterrupt:
            self.running = False
        t_send.join(timeout=1)
        t_recv.join(timeout=1)
        try:
            self.sock.close()
        except OSError:
            pass
        print("[client] stopped")


def main():
    client = Client()
    client.run()


if __name__ == "__main__":
    main()
