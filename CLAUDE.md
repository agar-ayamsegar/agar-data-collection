# agar-data-collection

Go API (`cmd/server`, Echo + MongoDB) for C05: MQ-137 + MQ-135 gas readings from raw whole
chickens, collected to train a freshness model. Data collection plan: Obsidian vault
`c5/data-collection-plan.md`.

## Arduino sketches — `cmd/iot/` (Uno, 115200 baud)

| Sketch | Use |
|---|---|
| `enose_logger` | Measurements. Raw only, no calibration on the Uno |
| `enose_calibration` | 5-hour test: warm-up, calibrates once, prints `t_s,raw137,rs137,rel137,raw135,rs135,rel135` |

Wiring: MQ-137 AO → A0, MQ-135 AO → A1. Uno ADC is 0–1023. RL on the modules: MQ-137 1 kΩ,
MQ-135 2 kΩ.

## Recording one measurement (`enose_logger`)

In Arduino IDE Serial Monitor (115200, line ending Newline), sensors warmed up ≥ 10 min:

1. Empty box, lid closed → send `s baseline A01`, wait 3 min, send `x`
2. Chicken in, lid closed → send `s sample A01`, wait 5 min, send `x`
3. Repeat for the next chicken with its own ID (`A02`, …)
4. Copy the whole output into `data/h<hours>.txt` (e.g. `data/h2.txt`), then Clear Output

Rows look like `12.0,A01,baseline,451,2.204,402,1.965` (`t_s,id,phase,raw137,v137,raw135,v135`).
The ID is the chicken ID — no spaces or commas. Without an ID (`s baseline`) a .txt may hold
only one measurement.

## .txt → JSON

**Always convert with the script — never retype the numbers by hand.**

```
python3 tools/log_to_json.py data/h2.txt     # writes data/A01_h2.json, data/A02_h2.json, ...
```

- Report every `warning:` to the user: they mean rows were skipped or missing (e.g. copy
  started late, Serial Monitor dropped lines). Don't silently accept them.
- The script fills the sensor arrays, `interval_ms`, `lid_closed_at_s` (= baseline length in
  seconds), `part: "WHOLE"`, `chicken_id` (from the command ID). The user fills `date`
  (RFC 3339, e.g. `2026-10-16T09:00:00+07:00`), `supplier`, `label` (`FRESH` / `UNFRESH`),
  `author`, `weight_g`, `hours_since_buy` by hand. `temp`, `humidity` stay `null` (no sensor).
- `hours_since_buy` must be ≥ 1 — the first measurement is recorded as `1`, never `0`
  (`required` rejects 0 on a `float64`).
- Field names follow `internal/domain/gas.go` (`baseline_137`, not `baseline_mq_137`). If the
  domain changes, update `tools/log_to_json.py` to match.
- The user POSTs the JSON via Postman. Every field the user fills is required by `domain.Gas`.

## 5-hour test (`enose_calibration`)

Serial Monitor drops old lines, so record with the terminal instead (Arduino IDE closed):

```
caffeinate -i sh -c '(stty 115200 raw; cat) < /dev/cu.usbmodemXXX | tee data/test_5jam.csv'
```
