# go-lab conventions

## Exported vs unexported (rule of thumb)

- Struct: public. Fields: private. Construct via `New`. Lint passes, done.
- Rationale (don't re-debate per case): public structs stay nameable
  (vars, params, godoc, embedding) and satisfy multiple interfaces
  through one constructor; private fields force construction through
  `New`, so misuse fails loud instead of leaking silently.
- Constructors return concrete types (`ireturn` enforced); callers bind
  to narrow interfaces. Exceptions require a `//nolint` reason.
- Strategy selection decides visibility: chosen at compose time (wiring
  picks the implementation) → public struct + constructor returns
  concrete. Chosen by runtime value (flag/config switch in a factory)
  → private structs + factory returns the interface (`nolint:ireturn`
  with reason).
