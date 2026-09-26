# Protobuf types — what flies on the wire

> TL;DR: small ints pack tight, `string`/`bytes` cost length,
> everything has a **zero default** (no null), `enum`/`oneof`
> model choices, **well-known types** cover time.

## 1. Scalars (proto → Go)

| Proto                          | Go                                 | Wire cost           | Note                                                    |
| ------------------------------ | ---------------------------------- | ------------------- | ------------------------------------------------------- |
| `int32`/`int64`                | `int32`/`int64`                    | varint (small = 1B) | Negative `int32` costs 10B — use `sint32` for negatives |
| `sint32`/`sint64`              | `int32`/`int64`                    | zigzag varint       | Negatives pack tight                                    |
| `uint32`/`uint64`              | `uint32`/`uint64`                  | varint              | —                                                       |
| `fixed32`/`fixed64`, `sfixed*` | `uint32`/`uint64`, `int32`/`int64` | Always 4/8B         | Big values: fixed beats varint                          |
| `float`/`double`               | `float32`/`float64`                | 4/8B                | —                                                       |
| `bool`                         | `bool`                             | 1B                  | —                                                       |
| `string`                       | `string`                           | UTF-8 + length      | Must be valid UTF-8                                     |
| `bytes`                        | `[]byte`                           | Raw + length        | Images, blobs, ciphertext                               |

- Field numbers (our `id = 1`) are the wire identity — names
  never travel. **Never reuse a number** (reserve deleted ones).
- No null: unset = zero value (`""`, `0`, `false`). Need
  presence? `optional` or wrapper types (`google.protobuf.*Value`).

## 2. Compound

- `repeated T` → `[]T` (packed for numerics by default).
- `map<K,V>` → `map[K]V` (keys: int/string/bool only, unordered).
- `enum` → named `int32` consts; first value **must be 0**
  (the default) — name it `*_UNSPECIFIED`.
- `oneof` → exactly one set (our lab could model `TodoEvent`
  payload variants); Go generates a wrapper interface.

## 3. Well-known types (don't reinvent)

- `google.protobuf.Timestamp` ↔ `time.Time`, `Duration` ↔
  `time.Duration`. Nanosecond precision, JSON-mapped as RFC 3339.
- `Struct`/`Value` for free-form JSON bridges, `Empty` for
  APIs returning nothing, `FieldMask` for partial updates.

## 4. Rules

- Pick by shape: counters → varint ints, money → `int64` cents
  (never float), text → `string`, blobs → `bytes`.
- Evolve by adding fields, never changing numbers or types.
  Renames are free (numbers travel, names don't).
