"""文件用途：P01 防止泄密和错误推断 DERP。创建者：HIHU-COCO；日期：2026-10-19。"""
import json
import unittest
from probe import summarize_status


class StatusSummaryTests(unittest.TestCase):
    def test_summary_excludes_identity_addresses_and_raw_secrets(self):
        result = summarize_status({"BackendState": "Running", "Self": {"Online": True, "HostName": "SECRET", "TailscaleIPs": ["SECRET"]}, "Peer": {"SECRET": {"Relay": "SECRET", "CurAddr": "SECRET"}}, "AuthURL": "SECRET"})
        self.assertNotIn("SECRET", json.dumps(result))
        self.assertEqual(result["peer_count"], 1)
        self.assertEqual(result["connection_type"], "unknown")
        self.assertFalse(result["derp_verified"])

    def test_no_peers_or_self_does_not_claim_connectivity(self):
        result = summarize_status({"BackendState": "NeedsLogin", "Peer": None, "Self": None})
        self.assertFalse(result["self_online"])
        self.assertEqual(result["peer_count"], 0)
        self.assertFalse(result["cross_network_verified"])

    def test_unknown_backend_is_not_echoed(self):
        self.assertEqual(summarize_status({"BackendState": "sensitive unexpected value"})["backend_state"], "Unknown")

    def test_malformed_status_is_rejected(self):
        for raw in ([], {"Peer": [1]}, {"Self": "wrong"}):
            with self.assertRaises(ValueError):
                summarize_status(raw)


if __name__ == "__main__":
    unittest.main()
