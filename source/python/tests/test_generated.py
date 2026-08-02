from __future__ import annotations

import importlib
import unittest
from pathlib import Path


class GeneratedPackageTest(unittest.TestCase):
    def test_all_authoritative_protos_are_generated(self) -> None:
        package_directory = Path(__file__).resolve().parents[1]
        proto_files = sorted((package_directory / "src" / "pole_specification" / "proto").glob("*.proto"))
        generated_files = sorted((package_directory / "src" / "pole_specification" / "generated").glob("*_pb2.py"))
        self.assertEqual(26, len(proto_files))
        self.assertEqual(26, len(generated_files))

    def test_grpc_stubs_import(self) -> None:
        modules = [
            "bootstrap_pb2_grpc",
            "grpc_config_api_pb2_grpc",
            "grpcapi_pb2_grpc",
            "grpcapi_ratelimiter_pb2_grpc",
            "workload_identity_pb2_grpc",
        ]
        for module in modules:
            imported = importlib.import_module(f"pole_specification.generated.{module}")
            self.assertIsNotNone(imported)


if __name__ == "__main__":
    unittest.main()
