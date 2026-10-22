"""屏幕采集：基于 mss 的低延迟区域截屏。

以屏幕中心为基准截取 width*height 区域，返回 BGR ndarray。
"""
import numpy as np

try:
    import mss
except ImportError:
    mss = None


class ScreenCapture:
    def __init__(self, width: int = 640, height: int = 480):
        if mss is None:
            raise RuntimeError("mss not installed: pip install mss")
        self.width = width
        self.height = height
        self._sct = mss.mss()
        # 计算中心点裁剪区域
        mon = self._sct.monitors[1]  # 主显示器
        self.screen_w = mon["width"]
        self.screen_h = mon["height"]
        self._region = self._make_region()

    def _make_region(self):
        if self.width <= 0 or self.height <= 0:
            # 全屏
            mon = self._sct.monitors[1]
            return {"left": mon["left"], "top": mon["top"],
                    "width": mon["width"], "height": mon["height"]}
        left = max(0, (self.screen_w - self.width) // 2)
        top = max(0, (self.screen_h - self.height) // 2)
        return {"left": left, "top": top,
                "width": self.width, "height": self.height}

    def grab(self) -> np.ndarray:
        """返回 BGR ndarray。"""
        shot = self._sct.grab(self._region)
        # mss 返回 BGRA，转 BGR
        arr = np.array(shot, dtype=np.uint8)
        return arr[:, :, :3]

    @property
    def region(self):
        return self._region
