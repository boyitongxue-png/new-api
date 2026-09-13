# Image Channel Resolution and Cost Routing

Image generation requests can opt into channel routing by resolution. The
feature is deliberately backwards compatible: channels without the metadata
below continue to use the existing priority and weight scheduler.

## Channel metadata

Put the following JSON in a channel's existing `other_info` field:

```json
{
  "supportedResolutions": ["1k", "2k", "4k"],
  "cost1k": 0.02,
  "cost2k": 0.05,
  "cost4k": 0.12,
  "cost8k": 0.30
}
```

`cost1k` through `cost8k` are per-image costs. The compatibility aliases
`cost_2k`, `costPerImage2k`, a nested `costs` object, and
`unitCostPerImage` are also accepted. `supportedResolutions` is optional; when
present, a channel is eligible only for listed tiers.

## Request resolution

OpenAI image requests may use the existing `size` field (`1024x1024`,
`2048x2048`, etc.) or the compatibility `resolution` field (`1k`, `2k`,
`4k`, `8k`). Unknown values do not activate cost routing.

When cost routing is active, eligible channels are ordered by ascending cost;
equal costs are resolved by channel ID. Retries walk that ordered list. If no
eligible channel has a valid cost for the requested tier, the legacy priority
and weight selection remains in effect.