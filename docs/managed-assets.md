# Managed assets: development and operations

The durable boundaries are in [the asset contract](../contracts/assets.md).

Admin reads `external_services.asset_storage` and `external_services.cloud_ops` from its explicit configuration file. Copy the shape from `apps/cn/admin/config/server.example.yaml` and put actual development values only in the ignored `config/server.yaml`. Nav Backend has no cloud secret; Nav Web needs public CDN origins only.

Run the opt-in storage acceptance from `apps/cn/admin`:

```powershell
$env:GOFURRY_ASSET_DEV_CONFIG = (Resolve-Path config/server.yaml).Path
go test ./internal/infra/assets -run TestRealDevStorage -count=1 -v
```

It requires development bucket/CDN names, calls real COS and R2 Put/Head/Get, checks both CDN responses, performs mirror repair, and verifies the fixed probe length/hash. It leaves a small set of immutable test objects (including a synthetic, unreferenced site ID); runtime credentials deliberately cannot delete objects. Normal unit tests skip live acceptance without this explicit configuration variable. A skipped test is not cloud acceptance.
