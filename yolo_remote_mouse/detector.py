"""YOLOv8 检测器：优先 TensorRT 引擎，降级 PyTorch。

统一输出归一化坐标的 Detection 列表，便于跨分辨率坐标映射。
"""
import os
import time
from typing import List, Optional

import numpy as np

from config import DetectConfig
from protocol import Detection


class YOLODetector:
    def __init__(self, cfg: DetectConfig):
        self.cfg = cfg
        self.model = None
        self._load()

    def _engine_path(self) -> Optional[str]:
        """若 model 同名 .engine 存在则优先返回引擎路径。"""
        model = self.cfg.model
        if model.endswith(".engine"):
            return model
        base = os.path.splitext(model)[0]
        engine = base + ".engine"
        if os.path.exists(engine):
            return engine
        return None

    def _load(self):
        from ultralytics import YOLO

        engine = self._engine_path() if self.cfg.use_tensorrt else None
        if engine:
            print(f"[detector] loading TensorRT engine: {engine}")
            self.model = YOLO(engine, task="detect")
        else:
            print(f"[detector] loading PyTorch model: {self.cfg.model}")
            self.model = YOLO(self.cfg.model, task="detect")
        # warmup
        try:
            dummy = np.zeros((self.cfg.imgsz, self.cfg.imgsz, 3), dtype=np.uint8)
            self.model(dummy, verbose=False, device=self.cfg.device,
                       half=self.cfg.half, imgsz=self.cfg.imgsz)
            print("[detector] warmup done")
        except Exception as e:
            print(f"[detector] warmup warning: {e}")

    def infer(self, frame_bgr: np.ndarray) -> List[Detection]:
        """输入 BGR ndarray，输出归一化坐标的 Detection 列表。"""
        cfg = self.cfg
        t0 = time.perf_counter()
        results = self.model(
            frame_bgr,
            verbose=False,
            device=cfg.device,
            half=cfg.half,
            imgsz=cfg.imgsz,
            conf=cfg.conf,
            iou=cfg.iou,
            max_det=cfg.max_det,
            classes=list(cfg.classes) if cfg.classes else None,
        )
        self.last_infer_ms = (time.perf_counter() - t0) * 1000

        dets: List[Detection] = []
        if not results:
            return dets
        r = results[0]
        if r.boxes is None or len(r.boxes) == 0:
            return dets

        h, w = frame_bgr.shape[:2]
        boxes = r.boxes.xywh.cpu().numpy()          # cx, cy, w, h (像素)
        confs = r.boxes.conf.cpu().numpy()
        clses = r.boxes.cls.cpu().numpy().astype(int)
        for box, conf, cls in zip(boxes, confs, clses):
            cx, cy, bw, bh = box
            dets.append(Detection(
                cls=int(cls), conf=float(conf),
                x=cx / w, y=cy / h, w=bw / w, h=bh / h,
            ))
        return dets
