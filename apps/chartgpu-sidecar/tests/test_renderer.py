import unittest

from renderer import render_svg


class RenderSvgTests(unittest.TestCase):
    def test_render_svg_contains_symbol_and_timeframe(self):
        payload = {
            "symbol": "BTC-USD",
            "timeframe": "1m",
            "width": 640,
            "height": 360,
            "theme": "dark",
        }
        svg = render_svg(payload)

        self.assertIn("BTC-USD", svg)
        self.assertIn("1m", svg)
        self.assertIn("width='640'", svg)
        self.assertIn("height='360'", svg)

    def test_render_svg_defaults_dimensions(self):
        payload = {"symbol": "ETH-USD", "timeframe": "5m"}
        svg = render_svg(payload)
        self.assertIn("width='800'", svg)
        self.assertIn("height='450'", svg)

    def test_render_svg_uses_candlestick_bars_not_line_polyline(self):
        payload = {"symbol": "ETH-USD", "timeframe": "5m", "theme": "dark"}
        svg = render_svg(payload)
        self.assertIn("candle-body", svg)
        self.assertIn("candle-wick", svg)
        self.assertNotIn("<polyline", svg)


if __name__ == "__main__":
    unittest.main()
