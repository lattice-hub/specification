const assert = require("node:assert/strict");
const test = require("node:test");
const specification = require("@lattice-hub/pole-specification");

test("loads every authoritative proto into gRPC package namespaces", () => {
  assert.equal(specification.protoFileNames.length, 26);
  const grpcObject = specification.loadPoleGrpcObject();
  assert.ok(grpcObject.v1);
  assert.ok(grpcObject.pole.sidecar.v1.SidecarSessionService);
  assert.ok(grpcObject.polaris.metric.v2.RateLimitGRPC);
});
