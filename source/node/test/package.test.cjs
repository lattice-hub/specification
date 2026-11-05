const assert = require("node:assert/strict");
const test = require("node:test");
const specification = require("@lattice-hub/pole-specification");

test("loads every authoritative proto into gRPC package namespaces", () => {
  assert.equal(specification.protoFileNames.length, 26);
  const grpcObject = specification.loadPoleGrpcObject();
  assert.ok(grpcObject.v1);
  assert.ok(grpcObject.pole.sidecar.v1.SidecarSessionService);
  assert.ok(grpcObject.pole.metric.v2.RateLimitGRPC);
  assert.ok(grpcObject.pole.config.v1.RenderSnapshot);
  assert.ok(grpcObject.pole.config.v1.ReleaseSummary);
  assert.ok(grpcObject.pole.config.v1.CopyReleaseToDraftRequest);
  assert.equal(grpcObject.v1.ConfigFile, undefined);
  assert.equal(grpcObject.polaris, undefined);
  const discover = grpcObject.v1.ConfigGRPC.service.Discover;
  assert.equal(discover.requestType.type.name, "DiscoverRequest");
  assert.equal(discover.responseType.type.name, "Response");
  assert.equal(discover.path, "/v1.ConfigGRPC/Discover");
  assert.equal(discover.requestStream, true);
  assert.equal(discover.responseStream, true);
  assert.deepEqual(Object.keys(grpcObject.v1.ConfigGRPC.service), ["Discover"]);
  assert.equal(grpcObject.pole.config.v1.ResolveRequest, undefined);
  assert.equal(grpcObject.pole.config.v1.WatchRequest, undefined);
});
