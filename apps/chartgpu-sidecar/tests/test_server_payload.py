import unittest

from server import validate_payload


class ValidatePayloadTests(unittest.TestCase):
    def test_validate_payload_accepts_minimal(self):
        payload = {"symbol": "BTC-USD", "timeframe": "1m"}
        validated = validate_payload(payload)
        self.assertEqual(validated["symbol"], "BTC-USD")
        self.assertEqual(validated["timeframe"], "1m")

    def test_validate_payload_rejects_missing_symbol(self):
        with self.assertRaises(ValueError):
            validate_payload({"timeframe": "1m"})


if __name__ == "__main__":
    unittest.main()
