from hashlib import sha256
from typing import Any, Dict, List, Tuple


def _generate_ohlc(symbol: str, timeframe: str, count: int) -> List[Tuple[float, float, float, float]]:
    digest = sha256(f"{symbol}:{timeframe}".encode("utf-8")).digest()
    seed = int.from_bytes(digest[:8], byteorder="big", signed=False)

    price = 80.0 + (seed % 450) / 10.0
    candles: List[Tuple[float, float, float, float]] = []

    for _ in range(count):
        seed = (1664525 * seed + 1013904223) % (2**32)
        move = ((seed % 2001) - 1000) / 620.0
        open_price = price
        close_price = max(1.0, open_price + move)

        seed = (1664525 * seed + 1013904223) % (2**32)
        wick_up = (seed % 700) / 420.0
        seed = (1664525 * seed + 1013904223) % (2**32)
        wick_down = (seed % 700) / 420.0

        high_price = max(open_price, close_price) + wick_up
        low_price = max(0.1, min(open_price, close_price) - wick_down)
        candles.append((open_price, high_price, low_price, close_price))
        price = close_price

    return candles


def render_svg(payload: Dict[str, Any]) -> str:
    symbol = str(payload.get("symbol", "UNKNOWN")).strip() or "UNKNOWN"
    timeframe = str(payload.get("timeframe", "1m")).strip() or "1m"

    width = int(payload.get("width") or 800)
    height = int(payload.get("height") or 450)
    theme = str(payload.get("theme", "light")).strip().lower() or "light"

    dark = theme == "dark"
    bg = "#0B0E11" if dark else "#F7F9FC"
    fg = "#EAECEF" if dark else "#1E2329"
    muted = "#5E6673" if dark else "#8892A0"
    grid = "#1E2329" if dark else "#D9DEE7"
    up = "#0ECB81"
    down = "#F6465D"

    chart_left = 52
    chart_right = width - 24
    chart_top = 96
    chart_bottom = height - 30
    chart_width = max(120, chart_right - chart_left)
    chart_height = max(100, chart_bottom - chart_top)

    candle_count = max(18, min(80, chart_width // 14))
    candles = _generate_ohlc(symbol, timeframe, candle_count)

    low_bound = min(low for _, _, low, _ in candles)
    high_bound = max(high for _, high, _, _ in candles)
    price_span = max(0.1, high_bound - low_bound)

    def price_to_y(price: float) -> float:
        pct = (price - low_bound) / price_span
        return chart_bottom - (pct * chart_height)

    step = chart_width / candle_count
    body_width = max(2.0, min(8.0, step * 0.66))

    grid_lines = []
    for idx in range(6):
        y = chart_top + (chart_height / 5) * idx
        grid_lines.append(
            f"<line x1='{chart_left}' y1='{y:.2f}' x2='{chart_right}' y2='{y:.2f}' stroke='{grid}' stroke-width='1' opacity='0.55'/>"
        )

    candle_shapes = []
    for idx, (open_price, high_price, low_price, close_price) in enumerate(candles):
        center_x = chart_left + (idx + 0.5) * step
        y_open = price_to_y(open_price)
        y_close = price_to_y(close_price)
        y_high = price_to_y(high_price)
        y_low = price_to_y(low_price)

        candle_color = up if close_price >= open_price else down
        body_top = min(y_open, y_close)
        body_height = max(1.8, abs(y_close - y_open))
        body_x = center_x - body_width / 2

        candle_shapes.append(
            f"<line class='candle-wick' x1='{center_x:.2f}' y1='{y_high:.2f}' x2='{center_x:.2f}' y2='{y_low:.2f}' stroke='{candle_color}' stroke-width='1.2'/>"
        )
        candle_shapes.append(
            f"<rect class='candle-body' x='{body_x:.2f}' y='{body_top:.2f}' width='{body_width:.2f}' height='{body_height:.2f}' fill='{candle_color}' rx='1'/>"
        )

    return (
        f"<svg xmlns='http://www.w3.org/2000/svg' width='{width}' height='{height}' viewBox='0 0 {width} {height}'>"
        f"<rect width='{width}' height='{height}' fill='{bg}'/>"
        f"<text x='24' y='44' fill='{fg}' font-size='24' font-family='monospace'>{symbol}</text>"
        f"<text x='24' y='74' fill='{muted}' font-size='16' font-family='monospace'>timeframe {timeframe}</text>"
        f"<rect x='{chart_left}' y='{chart_top}' width='{chart_width}' height='{chart_height}' fill='none' stroke='{grid}' stroke-width='1'/>"
        f"{''.join(grid_lines)}"
        f"{''.join(candle_shapes)}"
        "</svg>"
    )
