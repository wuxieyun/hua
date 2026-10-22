"""鼠标控制器：跨平台（优先 Windows win32api，降级 pynput）。

坐标映射：
- 检测结果是归一化坐标 [0,1]，对应 client 截屏区域；
- absolute 模式：把目标中心映射到屏幕绝对坐标并移动；
- relative 模式：计算目标中心相对截屏中心的偏移，相对移动鼠标。
"""
import sys
import time
from typing import Optional, Tuple

from config import MouseConfig


def _get_screen_size() -> Tuple[int, int]:
    try:
        if sys.platform == "win32":
            import ctypes
            user32 = ctypes.windll.user32
            return user32.GetSystemMetrics(0), user32.GetSystemMetrics(1)
        else:
            from pynput.mouse import Controller as _P
            # pynput 无直接屏幕尺寸，用 tkinter
            import tkinter
            root = tkinter.Tk()
            root.withdraw()
            w = root.winfo_screenwidth()
            h = root.winfo_screenheight()
            root.destroy()
            return w, h
    except Exception:
        return 1920, 1080


class MouseController:
    def __init__(self, cfg: MouseConfig,
                 capture_region: Tuple[int, int, int, int]):
        """
        capture_region: (left, top, width, height) 截屏区域屏幕坐标。
        """
        self.cfg = cfg
        self.left, self.top, self.cap_w, self.cap_h = capture_region
        self.screen_w, self.screen_h = _get_screen_size()
        self._smooth_dx = 0.0
        self._smooth_dy = 0.0
        # 后端
        self._backend = self._init_backend()

    def _init_backend(self):
        if sys.platform == "win32":
            try:
                import win32api  # noqa: F401
                return "win32api"
            except ImportError:
                pass
        try:
            from pynput.mouse import Controller  # noqa: F401
            return "pynput"
        except ImportError:
            return "none"

    def _position(self) -> Tuple[int, int]:
        if self._backend == "win32api":
            import win32api
            p = win32api.GetCursorPos()
            return p[0], p[1]
        elif self._backend == "pynput":
            from pynput.mouse import Controller
            return Controller().position
        return 0, 0

    def _move(self, dx: int, dy: int, absolute: bool = False):
        if self._backend == "win32api":
            import win32con
            if absolute:
                win32api.SetCursorPos((dx, dy))
            else:
                # mouse_event 相对移动
                import ctypes
                ctypes.windll.user32.mouse_event(
                    win32con.MOUSEEVENTF_MOVE, dx, dy, 0, 0)
        elif self._backend == "pynput":
            from pynput.mouse import Controller
            c = Controller()
            if absolute:
                c.position = (dx, dy)
            else:
                c.move(dx, dy)

    def apply_detection(self, target_x: float, target_y: float):
        """
        target_x, target_y: 归一化坐标 [0,1]，目标中心在截屏区域内的位置。
        """
        cfg = self.cfg
        # 截屏中心归一化坐标 = (0.5, 0.5)
        dx_norm = target_x - 0.5
        dy_norm = target_y - 0.5
        # 转成屏幕像素偏移（以截屏区域为参考）
        dx_px = dx_norm * self.cap_w * cfg.sensitivity
        dy_px = dy_norm * self.cap_h * cfg.sensitivity

        # 死区
        if abs(dx_px) < cfg.deadzone and abs(dy_px) < cfg.deadzone:
            return

        # 平滑（EMA）
        if cfg.smoothing > 0:
            self._smooth_dx = (1 - cfg.smoothing) * self._smooth_dx + \
                cfg.smoothing * dx_px
            self._smooth_dy = (1 - cfg.smoothing) * self._smooth_dy + \
                cfg.smoothing * dy_px
            dx_px, dy_px = self._smooth_dx, self._smooth_dy

        if cfg.mode == "absolute":
            # 绝对坐标：截屏区域 left/top + 目标位置
            abs_x = int(self.left + target_x * self.cap_w)
            abs_y = int(self.top + target_y * self.cap_h)
            self._move(abs_x, abs_y, absolute=True)
        else:
            self._move(int(round(dx_px)), int(round(dy_px)), absolute=False)
