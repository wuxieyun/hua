"""UDP 传输层：分片上传 + 单包下行，latest-frame-wins。

设计要点：
- 上传分片：单帧 JPEG 可能超过 MTU，按 mtu 切片发送；
- 重组超时：server 端若某帧超过 reassemble_timeout_ms 未收齐则丢弃；
- 最新帧优先：收到更高 frame_id 时，丢弃旧 frame_id 的未完成重组；
- 结果下行：检测结果 JSON 远小于 MTU，单包发送即可。
"""
import socket
import time
import threading
from typing import Optional, Tuple

from config import TransportConfig
from protocol import (UPLOAD_HEADER_SIZE, build_upload_packet,
                      parse_upload_packet, ResultPacket)


class FrameReassembler:
    """服务端帧重组器：按 frame_id 聚合分片，收齐后返回完整 JPEG。"""

    def __init__(self, timeout_ms: int = 15):
        self.timeout_s = timeout_ms / 1000.0
        self._frames = {}          # frame_id -> {"frags": {}, "total": n, "ts": t, "w": w, "h": h}
        self._lock = threading.Lock()

    def add_fragment(self, pkt: dict) -> Optional[Tuple[int, dict, bytes]]:
        """添加分片。若该帧收齐，返回 (frame_id, meta, jpeg_bytes)；否则 None。"""
        fid = pkt["frame_id"]
        now = time.monotonic()
        with self._lock:
            frame = self._frames.get(fid)
            if frame is None:
                frame = {"frags": {}, "total": pkt["frag_total"],
                         "ts": now, "w": pkt["width"], "h": pkt["height"],
                         "capture_ts": pkt["capture_ts"]}
                self._frames[fid] = frame
            else:
                frame["ts"] = now  # 刷新活跃时间

            frame["frags"][pkt["frag_idx"]] = pkt["payload"]

            if len(frame["frags"]) >= frame["total"]:
                # 收齐
                ordered = [frame["frags"][i] for i in range(frame["total"])
                           if i in frame["frags"]]
                if len(ordered) != frame["total"]:
                    return None
                jpeg = b"".join(ordered)
                meta = {"width": frame["w"], "height": frame["h"],
                        "capture_ts": frame["capture_ts"]}
                del self._frames[fid]
                return fid, meta, jpeg
            return None

    def drop_older_than(self, frame_id: int) -> None:
        """丢弃所有 frame_id < frame_id 的未完成重组（latest-frame-wins）。"""
        with self._lock:
            for fid in list(self._frames.keys()):
                if fid < frame_id:
                    del self._frames[fid]

    def sweep_stale(self) -> None:
        """清理超时未收齐的帧。"""
        now = time.monotonic()
        with self._lock:
            for fid in list(self._frames.keys()):
                if now - self._frames[fid]["ts"] > self.timeout_s:
                    del self._frames[fid]


def create_socket(bind_host: str = "0.0.0.0", bind_port: int = 0,
                  send_buf: int = 4 * 1024 * 1024,
                  recv_buf: int = 8 * 1024 * 1024) -> socket.socket:
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_SNDBUF, send_buf)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_RCVBUF, recv_buf)
    # 降低延迟：禁用 Nagle/缓冲对 UDP 影响小，但设下阻塞读超时便于 sweep
    s.settimeout(0.05)
    if bind_port:
        s.bind((bind_host, bind_port))
    return s


def split_into_fragments(frame_id: int, capture_ts: float,
                         width: int, height: int,
                         jpeg: bytes, mtu: int):
    """把一帧 JPEG 切成若干 UDP 包。"""
    payload_size = mtu - UPLOAD_HEADER_SIZE
    total = (len(jpeg) + payload_size - 1) // payload_size
    packets = []
    for i in range(total):
        chunk = jpeg[i * payload_size:(i + 1) * payload_size]
        packets.append(build_upload_packet(
            frame_id, i, total, capture_ts, width, height, chunk))
    return packets


class UdpSender:
    """客户端发送器：负责把帧分片发出。"""

    def __init__(self, cfg: TransportConfig):
        self.cfg = cfg
        self.sock = create_socket()
        self.target = (cfg.client_host, cfg.server_port)

    def send_frame(self, frame_id: int, capture_ts: float,
                   width: int, height: int, jpeg: bytes) -> int:
        packets = split_into_fragments(
            frame_id, capture_ts, width, height, jpeg, self.cfg.mtu)
        sent = 0
        for pkt in packets:
            try:
                sent += self.sock.sendto(pkt, self.target)
            except OSError:
                break
        return sent

    def send_result(self, result: ResultPacket, addr: Tuple[str, int]) -> int:
        """server 侧下行结果。"""
        data = result.serialize()
        return self.sock.sendto(data, addr)

    def close(self):
        try:
            self.sock.close()
        except OSError:
            pass
