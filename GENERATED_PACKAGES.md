# Node.js 与 Python gRPC 生成包

本仓库以 `api/` 中的 Proto 为唯一权威来源，在 CI 和 Release 阶段生成并发布：

| 语言 | Registry | 包名 | 生成方式 |
| --- | --- | --- | --- |
| Node.js | npm | `@lattice-hub/pole-specification` | `@grpc/proto-loader` 与 `proto-loader-gen-types` |
| Python | PyPI | `pole-specification` | `grpcio-tools` |

生成文件不提交到 Git。普通分支和 Pull Request 由 `.github/workflows/testing.yml`
重新生成、测试和打包；发布 GitHub Release 后，
`.github/workflows/generated-release.yml` 使用 Release tag 设置包版本并发布。
Python 生成器固定版本，且包声明的 `grpcio`、`protobuf` 最低版本与生成代码一致。

## 本地验证

```shell
cd source/node
npm ci
npm test
npm pack --dry-run
```

```shell
cd source/python
python -m pip install ".[dev]"
python scripts/generate.py
PYTHONPATH=src python -m unittest discover -s tests
python -m build
```

## Registry 配置

发布工作流使用 GitHub Actions OIDC，不需要保存长期 npm 或 PyPI Token。

### npm

首次手工发布 `@lattice-hub/pole-specification` 后，在 npm 包设置中配置
Trusted Publisher：

- Organization：`lattice-hub`
- Repository：`specification`
- Workflow：`generated-release.yml`
- Environment：`npm`
- Allowed actions：`npm publish`

### PyPI

在 PyPI 创建 `pole-specification` 的 Pending Publisher，或在首次发布后配置
Trusted Publisher：

- Owner：`lattice-hub`
- Repository：`specification`
- Workflow：`generated-release.yml`
- Environment：`pypi`

GitHub 仓库需要同名的 `npm` 和 `pypi` Environments。可以按需为环境增加审批人；
不需要配置 `NPM_TOKEN` 或 `PYPI_API_TOKEN`。

## 版本映射

Node.js 使用去掉前导 `v` 的 Release tag，例如 `v0.1.0-ALPHA.41` 发布为
`0.1.0-ALPHA.41`。Python 按 PEP 440 规范化同一版本，发布为 `0.1.0a41`。
