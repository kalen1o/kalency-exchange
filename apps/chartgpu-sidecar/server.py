import hashlib
import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer
from typing import Any, Dict

from renderer import render_svg


def validate_payload(payload: Dict[str, Any]) -> Dict[str, Any]:
    if not isinstance(payload, dict):
        raise ValueError("payload must be an object")

    symbol = str(payload.get("symbol", "")).strip()
    timeframe = str(payload.get("timeframe", "")).strip()
    if not symbol:
        raise ValueError("symbol is required")
    if not timeframe:
        raise ValueError("timeframe is required")

    width = payload.get("width", 800)
    height = payload.get("height", 450)
    try:
        width = int(width)
        height = int(height)
    except (TypeError, ValueError) as exc:
        raise ValueError("width and height must be integers") from exc
    if width <= 0 or height <= 0:
        raise ValueError("width and height must be positive")

    theme = str(payload.get("theme", "light")).strip() or "light"

    return {
        "symbol": symbol,
        "timeframe": timeframe,
        "width": width,
        "height": height,
        "theme": theme,
    }


class Handler(BaseHTTPRequestHandler):
    def _send_json(self, status: int, payload: Dict[str, Any]) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        self.wfile.write(body)

    def do_OPTIONS(self):  # noqa: N802
        self.send_response(204)
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")
        self.end_headers()

    def do_GET(self):  # noqa: N802
        if self.path != "/healthz":
            self._send_json(404, {"error": "not found"})
            return
        self._send_json(200, {"status": "ok"})

    def do_POST(self):  # noqa: N802
        if self.path != "/render":
            self._send_json(404, {"error": "not found"})
            return

        content_length = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(content_length or 0)
        try:
            payload = json.loads(raw or b"{}")
            validated = validate_payload(payload)
            svg = render_svg(validated)
        except (json.JSONDecodeError, ValueError) as exc:
            self._send_json(400, {"error": str(exc)})
            return

        render_id = hashlib.sha256(svg.encode("utf-8")).hexdigest()
        self._send_json(
            200,
            {
                "renderId": render_id,
                "artifactType": "image/svg+xml",
                "artifact": svg,
                "meta": validated,
            },
        )


def run() -> None:
    port = int(os.getenv("PORT", "8086"))
    server = HTTPServer(("0.0.0.0", port), Handler)
    print(f"chartgpu-sidecar listening on :{port}")
    server.serve_forever()


if __name__ == "__main__":
    run()
