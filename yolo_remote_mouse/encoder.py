"""JPEG 编码器：OpenCV imencode，可调质量与缩放。"""
import cv2
import numpy as np


class JPEGEncoder:
    def __init__(self, quality: int = 70, scale: float = 1.0):
        self.quality = quality
        self.scale = scale
        self._params = [int(cv2.IMWRITE_JPEG_QUALITY), int(quality)]

    def encode(self, frame_bgr: np.ndarray) -> bytes:
        if self.scale != 1.0:
            h, w = frame_bgr.shape[:2]
            frame_bgr = cv2.resize(
                frame_bgr, (int(w * self.scale), int(h * self.scale)),
                interpolation=cv2.INTER_AREA)
        ok, buf = cv2.imencode(".jpg", frame_bgr, self._params)
        if not ok:
            return b""
        return buf.tobytes()
