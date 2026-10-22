"""把 YOLOv8 模型导出为 TensorRT FP16 引擎，用于服务端低延迟推理。

用法：
    python export_engine.py --model yolov8n.pt --imgsz 640 --half
导出后生成 yolov8n.engine，detector.py 会自动优先加载。
"""
import argparse


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--model", default="yolov8n.pt")
    parser.add_argument("--imgsz", type=int, default=640)
    parser.add_argument("--half", action="store_true", default=True)
    parser.add_argument("--device", default="0")
    parser.add_argument("--dynamic", action="store_true",
                        help="动态 batch/shape，固定 shape 推理更快")
    args = parser.parse_args()

    from ultralytics import YOLO
    model = YOLO(args.model)
    # export 到 TensorRT
    # imgsz 固定为正方形；不启用 dynamic 以获得最快推理
    model.export(
        format="engine",
        imgsz=args.imgsz,
        half=args.half,
        device=args.device,
        dynamic=args.dynamic,
        simplify=True,
    )
    print(f"[export] engine saved: {args.model.replace('.pt', '.engine')}")


if __name__ == "__main__":
    main()
