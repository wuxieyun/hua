"""传输协议定义：上行帧分片包 + 下行结果包。

设计原则：
- 固定二进制头，避免 JSON 解析开销；
- 每个分片携带 frame_id，server 端可 latest-frame-wins；
- 下行结果足够小，单包即可。
"""
import struct
import json
from dataclasses import dataclass
from typing import List, Optional

# magic 标识
MAGIC_UPLOAD = b"YRMU"   # YOLO Remote Mouse Upload
MAGIC_RESULT = b"YRMR"   # YOLO Remote Mouse Result

# 上行包头: magic(4) + frame_id(4) + frag_idx(2) + frag_total(2)
#         + capture_ts(8) + width(2) + height(2) + payload_len(2) = 26
UPLOAD_HEADER_FMT = "<4sIHHdHHH"
UPLOAD_HEADER_SIZE = struct.calcsize(UPLOAD_HEADER_FMT)

# 下行结果头: magic(4) + frame_id(4) + server_ts(8) + payload_len(2) = 18
RESULT_HEADER_FMT = "<4sIdH"
RESULT_HEADER_SIZE = struct.calcsize(RESULT_HEADER_FMT)


@dataclass
class Detection:
    cls: int
    conf: float
    # 相对于捕获区域的归一化坐标 [0,1]
    x: float   # 中心 x
    y: float   # 中心 y
    w: float   # 宽
    h: float   # 高

    def to_dict(self):
        return {"cls": self.cls, "conf": round(self.conf, 4),
                "x": round(self.x, 5), "y": round(self.y, 5),
                "w": round(self.w, 5), "h": round(self.h, 5)}

    @classmethod
    def from_dict(cls, d):
        return cls(d["cls"], d["conf"], d["x"], d["y"], d["w"], d["h"])


@dataclass
class ResultPacket:
    frame_id: int
    server_ts: float            # server 完成推理的 monotonic 时间戳
    detections: List[Detection]

    def serialize(self) -> bytes:
        payload = json.dumps(
            {"detections": [d.to_dict() for d in self.detections]},
            separators=(",", ":")
        ).encode("utf-8")
        header = struct.pack(RESULT_HEADER_FMT, MAGIC_RESULT,
                             self.frame_id, self.server_ts, len(payload))
        return header + payload

    @classmethod
    def deserialize(cls, data: bytes) -> Optional["ResultPacket"]:
        if len(data) < RESULT_HEADER_SIZE:
            return None
        magic, frame_id, server_ts, plen = struct.unpack(
            RESULT_HEADER_FMT, data[:RESULT_HEADER_SIZE])
        if magic != MAGIC_RESULT:
            return None
        body = data[RESULT_HEADER_SIZE:RESULT_HEADER_SIZE + plen]
        try:
            obj = json.loads(body)
            dets = [Detection.from_dict(d) for d in obj.get("detections", [])]
            return cls(frame_id, server_ts, dets)
        except (json.JSONDecodeError, KeyError, TypeError):
            return None


def build_upload_packet(frame_id: int, frag_idx: int, frag_total: int,
                        capture_ts: float, width: int, height: int,
                        payload: bytes) -> bytes:
    header = struct.pack(UPLOAD_HEADER_FMT, MAGIC_UPLOAD, frame_id,
                         frag_idx, frag_total, capture_ts, width, height,
                         len(payload))
    return header + payload


def parse_upload_packet(data: bytes):
    """返回 dict 或 None（magic 不匹配/长度不足）。"""
    if len(data) < UPLOAD_HEADER_SIZE:
        return None
    magic, frame_id, frag_idx, frag_total, capture_ts, width, height, plen = \
        struct.unpack(UPLOAD_HEADER_FMT, data[:UPLOAD_HEADER_SIZE])
    if magic != MAGIC_UPLOAD:
        return None
    payload = data[UPLOAD_HEADER_SIZE:UPLOAD_HEADER_SIZE + plen]
    return {
        "frame_id": frame_id,
        "frag_idx": frag_idx,
        "frag_total": frag_total,
        "capture_ts": capture_ts,
        "width": width,
        "height": height,
        "payload": payload,
    }
