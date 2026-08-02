# @lattice-hub/pole-specification

Pole Specification 的官方 Node.js gRPC 定义包。包内包含权威 Proto 副本、
`@grpc/proto-loader` 生成的 TypeScript 类型，以及加载完整 gRPC package object 的入口。

```javascript
const { loadPoleGrpcObject } = require("@lattice-hub/pole-specification");

const pole = loadPoleGrpcObject();
```

该包由 specification GitHub Release 自动生成和发布，不接受手工修改生成目录。
