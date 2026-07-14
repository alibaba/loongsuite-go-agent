# Changelog

## Unreleased

### Bug Fixes

- **redis (go-redis v8/v9, rueidis, redigo)**: Do not mark Redis nil sentinels (`redis.Nil`, `rueidis.Nil`, `redigo.ErrNil`) as span errors.
  Missing keys / cache misses stay non-Error for span status. Callers still receive the original sentinel error.

### Upgrade notes

- **Monitoring / SLO**: Redis error-rate panels based on span status `Error` will no longer include nil replies (cache misses). This is intentional. Adjust alerts/SLOs if they previously treated misses as failures.
- **`db.query.text`**: Unchanged (still uses `*redis.Cmd` Stringer output).
