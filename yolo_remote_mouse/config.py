"""全局配置：传输、推理、采集、鼠标控制的所有可调参数。"""
from dataclasses import dataclass, field
from typing import Tuple


@dataclass
class CaptureConfig:
    # 截屏区域：以屏幕中心为基准的宽高。设为 0 表示全屏。
    width: int = 640
    height: int = 480
    # 帧率上限（仅用于 client 侧节流，server 以推理速度为准）
    target_fps: int = 60


@dataclass
class EncodeConfig:
    jpeg_quality: int = 70          # 1-100，越低码率越小但画质越差
    # 可选：缩小后再编码（推理前已 resize，这里可降采样降带宽）
    scale: float = 1.0


@dataclass
class TransportConfig:
    server_host: str = "0.0.0.0"    # server 监听地址
    server_port: int = 9527         # 上传/下行共用同一个 UDP 端口
    client_host: str = "127.0.0.1"  # client 连接的 server 地址（运行时改）
    # UDP 分片大小：留 64 字节给 IP/UDP 头 + 应用头，避免分片
    mtu: int = 1400
    # 帧重组超时（ms），超过则丢弃该帧
    reassemble_timeout_ms: int = 15
    # 最新帧优先：server 收到更新 frame_id 时丢弃未重组完的旧帧
    latest_frame_wins: bool = True


@dataclass
class DetectConfig:
    model: str = "yolov8n.pt"       # 可用 yolov8n.pt / yolov8n.engine(TensorRT)
    imgsz: int = 640
    conf: float = 0.25
    iou: float = 0.45
    device: str = "0"               # cuda:0 / cpu
    half: bool = True               # FP16
    max_det: int = 10
    # 只保留这些类别（COCO id）。0=person。空列表表示不过滤。
    classes: Tuple[int, ...] = (0,)
    # TensorRT 引擎文件后缀（与 model 同名 .engine 时自动加载）
    use_tensorrt: bool = True


@dataclass
class MouseConfig:
    # 鼠标移动模式："relative" 相对当前位置 / "absolute" 绝对坐标
    mode: str = "relative"
    # 灵敏度系数（相对模式下乘以检测偏移）
    sensitivity: float = 1.0
    # 死区：偏移像素小于该值不移动（防抖）
    deadzone: float = 2.0
    # 是否只在按下某个键时才移动（默认 False，始终移动到目标中心）
    require_hotkey: bool = False
    hotkey: str = "mouse_right"
    # 平滑：指数移动平均系数 (0~1)，越小越平滑但延迟增加
    smoothing: float = 0.0


@dataclass
class Config:
    capture: CaptureConfig = field(default_factory=CaptureConfig)
    encode: EncodeConfig = field(default_factory=EncodeConfig)
    transport: TransportConfig = field(default_factory=TransportConfig)
    detect: DetectConfig = field(default_factory=DetectConfig)
    mouse: MouseConfig = field(default_factory=MouseConfig)
    # 打印延迟统计的间隔（秒）
    stats_interval: float = 2.0
    # 是否打印每帧详细日志（调试用，会增加开销）
    verbose: bool = False


CONFIG = Config()
