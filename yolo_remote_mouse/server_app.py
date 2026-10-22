"""服务端主程序：收包-重组-解码-推理-回传结果。

单线程事件循环：
- recv 非阻塞地排空 socket 缓冲区；
- 重组器按 frame_id 聚合，收齐即产出；
- 同一轮若产出多帧，只推理最新帧（latest-frame-wins）；
- 推理完成后把结果回发给对应 client 地址。
"""
import socket
import time
from collections import deque

import cv2
import numpy as np

from config import CONFIG
from detector import YOLODetector
from protocol import parse_upload_packet, ResultPacket
from udp_transport import FrameReassembler, create_socket


class Server:
    def __init__(self):
        cfg = CONFIG.transport
        self.sock = create_socket(bind_host=cfg.server_host,
                                  bind_port=cfg.server_port)
        self.reassembler = FrameReassembler(
            timeout_ms=cfg.reassemble_timeout_ms)
        self.detector = YOLODetector(CONFIG.detect)
        self.running = True
        # 统计
        self.stats = {"frames": 0, "infer_ms": 0.0,
                      "proc_ms": 0.0, "last_print": time.monotonic()}

    def _recv_all(self):
        """非阻塞排空 socket，返回 [(pkt_dict, addr), ...]。"""
        pkts = []
        while True:
            try:
                data, addr = self.sock.recvfrom(65535)
            except socket.timeout:
                break
            except OSError:
                break
            pkt = parse_upload_packet(data)
            if pkt is not None:
                pkts.append((pkt, addr))
        return pkts

    def run(self):
        print(f"[server] listening on {CONFIG.transport.server_host}:"
              f"{CONFIG.transport.server_port}")
        # 用 deque 暂存本轮回收到的完整帧
        ready = deque()
        while self.running:
            # 1. 收包
            pkts = self._recv_all()
            # 2. 重组
            latest_fid = -1
            for pkt, addr in pkts:
                if CONFIG.transport.latest_frame_wins and pkt["frame_id"] > latest_fid:
                    latest_fid = pkt["frame_id"]
                    self.reassembler.drop_older_than(pkt["frame_id"])
                res = self.reassembler.add_fragment(pkt)
                if res is not None:
                    fid, meta, jpeg = res
                    ready.append((fid, meta, jpeg, addr))
            # 3. latest-frame-wins：只保留最新的完整帧
            if ready:
                fid, meta, jpeg, addr = ready[-1]
                ready.clear()
                self._process_frame(fid, meta, jpeg, addr)
            # 4. 清理超时帧
            self.reassembler.sweep_stale()

    def _process_frame(self, frame_id: int, meta: dict,
                       jpeg: bytes, addr):
        t_start = time.monotonic()
        # 解码
        arr = np.frombuffer(jpeg, dtype=np.uint8)
        frame = cv2.imdecode(arr, cv2.IMREAD_COLOR)
        if frame is None:
            return
        # 推理
        dets = self.detector.infer(frame)
        t_done = time.monotonic()
        # 回传
        result = ResultPacket(frame_id, t_done, dets)
        try:
            self.sock.sendto(result.serialize(), addr)
        except OSError:
            pass
        # 统计：服务端处理耗时（解码+推理+发送）
        self.stats["frames"] += 1
        self.stats["infer_ms"] = self.detector.last_infer_ms
        proc_ms = (t_done - t_start) * 1000
        self.stats["proc_ms"] = proc_ms
        if t_done - self.stats["last_print"] > CONFIG.stats_interval:
            print(f"[server] fps~{self.stats['frames'] / CONFIG.stats_interval:.0f} "
                  f"infer={self.stats['infer_ms']:.1f}ms "
                  f"proc={proc_ms:.1f}ms "
                  f"dets={len(dets)}")
            self.stats["frames"] = 0
            self.stats["last_print"] = t_done


def main():
    server = Server()
    try:
        server.run()
    except KeyboardInterrupt:
        print("[server] stopped")


if __name__ == "__main__":
    main()
