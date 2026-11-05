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

    def test_unified_configuration_contract(self) -> None:
        from pole_specification.generated import config_pb2, grpc_config_api_pb2, grpcapi_ratelimiter_pb2
        service = grpc_config_api_pb2.DESCRIPTOR.services_by_name["ConfigGRPC"]
        self.assertEqual(["Discover"], [method.name for method in service.methods])
        self.assertEqual("pole.config.v1.DiscoverRequest", service.methods[0].input_type.full_name)
        self.assertEqual("v1.Response", service.methods[0].output_type.full_name)
        self.assertTrue(service.methods[0].client_streaming)
        self.assertTrue(service.methods[0].server_streaming)
        self.assertNotIn("ResolveRequest", config_pb2.DESCRIPTOR.message_types_by_name)
        self.assertNotIn("WatchRequest", config_pb2.DESCRIPTOR.message_types_by_name)
        self.assertEqual("pole.metric.v2", grpcapi_ratelimiter_pb2.DESCRIPTOR.package)
        self.assertNotIn("definition", config_pb2.ReleaseSummary.DESCRIPTOR.fields_by_name)
        self.assertNotIn("observation_cursor", config_pb2.RenderSnapshot.DESCRIPTOR.fields_by_name)
        self.assertEqual(4, config_pb2.VALUE_TYPE_DECIMAL)

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
