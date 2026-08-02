import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";
import { resolve } from "node:path";
import generatedProtoFileNames from "./proto-files.json";

const protoFileNames = generatedProtoFileNames as readonly string[];

export { protoFileNames };

export function loadPolePackageDefinition(): protoLoader.PackageDefinition {
  const protoDirectory = resolve(__dirname, "../proto");
  return protoLoader.loadSync(
    protoFileNames.map((name) => resolve(protoDirectory, name)),
    {
      includeDirs: [protoDirectory],
      longs: String,
      enums: String,
      defaults: true,
      oneofs: true,
    },
  );
}

export function loadPoleGrpcObject(): grpc.GrpcObject {
  return grpc.loadPackageDefinition(loadPolePackageDefinition());
}
