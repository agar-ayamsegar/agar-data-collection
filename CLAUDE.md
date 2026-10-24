# agar-data-collection

Go API (`cmd/server`, Echo + MongoDB) for C05: MQ-137 + MQ-135 gas readings from raw whole
chickens, collected to train a freshness model. Data collection plan: Obsidian vault
`c5/data-collection-plan.md`.

Project-wide context is in `../AGENTS.md` (rules, product, routine) and `../MEMORY.md`
(history of A01–A14, results, design files, open issues). Read them before analysing data.

**EDA:** `eda/FINDINGS.md` holds the current findings and every number behind them
(features, thresholds, held-out accuracy, the humidity confound). Regenerate the report
with `python3 tools/eda_report.py` → `eda/index.html` after each new chicken
(`--json` for raw numbers, `--demo` for the self-check). Stdlib only, no pandas needed.

## Where the data stands (2026-10-24)

| | |
|---|---|
| Analysed (proper EDA) | **A01–A06** only. `eda/FINDINGS.md` and `eda/index.html` stop here |
| JSON made | **A03–A14 all converted**: `cmd/iot/data/a03/A03_h1.json` … `a14/A14_h7.json` (55 checkpoints, `smell_level`/`texture`/`color` filled). `a15/` exists but is empty. |
| Quick check on A03–A14 | `../MEMORY.md`, "Update 2026-10-24": MQ-135 `rel300` LOCO 89 % vs label, 93 % vs smell ≥ 4, cut ≈ 0.305. Consistent with the 6-chicken EDA |
| Stray file | `data/A10_A10_20261023_192013.json` (doubled name, mostly empty), superseded by `cmd/iot/data/a10/` |

The thresholds quoted below are still the 6-chicken EDA numbers. Re-run `tools/eda_report.py`
on A01–A14 before quoting a final accuracy. **A09–A14 JSONs start at `hours_since_buy: 2`** (no h1),
which matters for the MQ-137-relative-to-h1 feature.

- **Chickens now run in parallel**, several per day: A07+A08 (22 Oct), A09+A10 (23 Oct),
  **A11+A12+A13+A14 (24 Oct)** with `h1` staggered 09:45 / 10:15 / 10:45 / 11:10. That is
  the alternate-start-time fix for the humidity confound (`../MEMORY.md` to-do,
  `AGENTS.md` §6b) — the reversed-clock evening chicken is still **not** done.
- **IDs were sent bare** (`s sample A11`, not `A11_2h`), so checkpoints are only told
  apart by the file timestamp. Before converting, read the `# CMD` lines and work out the
  hour yourself. Known mix-ups: `s baseline A010` and `s s baseline A09` (typos), and
  `data/A09_20261023_154956.csv` contains an `s baseline A10` line.
- `data/` is in `.gitignore`. The logs are local only — never assume git has a copy.

## Design files (`../design/`)

| File | What |
|---|---|
| `../design/agar-sniffer-3d.html` | 3D handheld sniffer concept (three.js; open in a browser). Funnel Ø7 cm on 1 cm legs → ~150 ml chamber (MQ-137, MQ-135, DHT) → fan **after** the sensors. Steps: reference → sniff → clear. Parts, power (~2–2.5 W, 2×18650 ≈ 7–9 h), and the data rule: `"method": "funnel"`, collect box + funnel pairs. Published copy: https://claude.ai/artifact/Afs9d3kDW67VMS1TYdjHYG |
| `../design/lakban_setup.png` / `.svg` | Cable/Arduino taping: lakban outside the box only, with a slack loop. No tape inside the box or on the sensors (glue VOCs raise the MQ-135) |
| iPad app canvas "Lalalla" | https://claude.ai/artifact/CiYDYzYrE4sAR5LQ5RKnLn. Owner dashboard, batch detail, 8 live scan screens + clickable prototype. Scan output is **BAGUS / RUSAK** only, plus grey **ULANGI SCAN** for invalid readings |

## Working with the user

- The user is a beginner. Give short, simple, concrete answers, using tables and numbered
  steps. One clear next step is better than many options.
- The user mixes English and Indonesian (bagus, rusak, apek, asem, kecut, pesing, tawar,
  gurih, ayam).
- **Look at the real log files** (`data/*.csv`) before explaining anything "weird". Most
  surprises so far were hardware or handling problems, visible in the logs.
- Never edit or delete raw logs or original JSONs. Cleaned copies go in a separate `clean/`
  folder.

## Arduino sketches — `cmd/iot/` (Uno, 115200 baud)

| Sketch | Use |
|---|---|
| `enose_logger` | Measurements. Raw only, no calibration on the Uno |
| `enose_calibration` | 5-hour test: warm-up, calibrates once, prints `t_s,raw137,rs137,rel137,raw135,rs135,rel135` |

Wiring: MQ-137 AO → A0, MQ-135 AO → A1, DHT11 data → D3. VCC → 5V, GND → GND. Uno ADC is
0–1023. RL on the modules: MQ-137 1 kΩ, MQ-135 2 kΩ.

- **ADC reference:**
  - A01 used `analogReference(INTERNAL)`, `VREF = 1.1`. It saturated at 1023 on spoiled
    chicken.
  - **Since A02: `analogReference(DEFAULT)`, `VREF = 5.0`.** Raw counts are about 4.5×
    smaller. Never compare raw across vrefs; compare `v` or ratios.
- Logger row: `t_s,sample_id,label,raw137,v137,raw135,v135,temp,humidity`.
  - The sketch currently prints `32 C, 48 %` for temp/humidity. `log_to_json.py` copes;
    plain numbers would be cleaner.
  - Rows come about every 1.5 s, because the DHT11 read blocks.
- **ESP32-S3 port (planned):**
  - MQ AO can reach 5 V, so each AO needs a **voltage divider: 10 kΩ (AO → pin) + 15 kΩ
    (pin → GND)**. That gives `v_pin = 0.6 × v_sensor`. 10k/20k gives 3.33 V, which is
    above the ADC range.
  - Use ADC1 pins (GPIO1/2). DHT11 on GPIO4.
  - Power from a 5 V 2 A wall charger.
  - Switch hardware only at a new chicken. Record the device in `notes`.

## Recording (use `tools/serial_log.py`, not the Serial Monitor)

```
cd ~/Development/Apple/Agar/agar-data-collection
make serial-log NAME=A03        # = caffeinate -i python3 tools/serial_log.py A03
```

- Close the Arduino IDE Serial Monitor first. Only one program can open the port.
- Opening the port resets the Arduino. Wait for `# READY`.
- Writes `data/<NAME>_<YYYYmmdd_HHMMSS>.csv`, with a clock on every row and `# CMD/START/STOP`
  lines.
- **Commands** (typed in the same terminal): `s baseline <id>`, `s sample <id>`, `x` (stop),
  `m <note>` (marker), `q` (quit).
- **IDs:** one per checkpoint, chicken + hour, for example `A03_2h`. A bare `A03` makes the
  checkpoints hard to tell apart.

### Measurement routine (one checkpoint per hour)

**Always record an `h1` checkpoint for every chicken.** It is the per-chicken reference
that makes the MQ-137 usable (`eda/FINDINGS.md` §3).

1. **Once per day:**
   - Plug in the USB and **don't touch the wires or cable again**.
   - Tape the Arduino and the cable down, so opening the lid can't pull them.
   - Warm up ~20 min.
2. The chicken stays in its own closed container, **away from the box**.
3. Air the box for 2 min, then close it **empty**. Wait until it's flat: ±2 counts for
   1 min, with **no steps**.
4. `s baseline A03_2h`, about 1 min.
5. `s sample A03_2h` straight away. Prefer **no `x` before it** — it keeps the sample clock
   honest and the runs comparable. If you do type one, the run is still usable: 11 runs have
   a gap and they track smell as well as the clean ones (`eda/FINDINGS.md` §5).
6. Take the chicken out of its container, **1 min on a tray in open air** (no fanning, never
   open the container next to the box). Put it in and close the lid within 5 s.
7. After **5 min**, `x`. (Was 2 min. `rel` at 300 s is 96 % held-out vs 88 % for 120 s —
   three extra minutes, +8 points, no extra work. See `eda/FINDINGS.md` §2.)
8. **Smell (never taste raw chicken)** and record `smell_level`, `texture` and `color`.
   Use gloves once the chicken is off.
9. Stop the chicken after the checkpoint where smell is 4–5 or the texture is slimy/mushy.

The old method (Serial Monitor → copy into `data/h<hours>.txt`) still converts with
`log_to_json.py`, but it drops lines and has no clock. Use it only as a fallback.

## Math and features

```
v   = raw * VREF / 1023
Rs  ∝ (5 - v) / v
rel = Rs_sample / Rs_baseline      # 1.0 = empty box; LOWER = MORE GAS
```

- Baseline value: the median of the last ~20 **flat** baseline rows. Drop lid/hand spikes.
- Main feature: `rel` at **300 s after `s sample`** on the **MQ-135** (threshold `< 0.285`,
  96 % leave-one-chicken-out against smell ≥ 4). Extras: `rel` at 120 s and the minimum
  `rel`, for both sensors.
- **Combining the sensors:** an `AND`/`OR` rule on the two raw `rel` values buys nothing
  (96 %, same as MQ-135 alone). What works is the **ratio** `rel300_137 / rel300_135`,
  which fires **high**: `> 2.01` is also 96 % held-out, needs no h1 reference, and is less
  humidity-tangled (r +0.70 vs −0.80). Its errors differ from MQ-135's, so run both rules:
  **agree → call it (21/21 correct); disagree → ULANGI SCAN** (2 of 23 runs).
  See `eda/FINDINGS.md` §2b.
- **MQ-137 is a relative sensor, not an absolute one.** It responds as hard as the MQ-135
  (min `rel` 0.14 vs 0.13), but 0.65 means a different thing on each chicken. Judge it
  against that chicken's own h1 reading. Pooled it is only 69 % held-out; rescaled by h1
  it goes from ρ −0.54 to −0.71.
- ⚠ **Humidity is confounded with age (r = +0.94)** — and it is **the room, not the
  chicken**: box RH starts at 35 % in the morning and 77 % at night, while the chicken
  adds only +1–13 %. A05 and A06 ran in parallel and match within 1–6 points at the same
  clock time. The baseline does **not** cancel it (`rel300_135` ~ box RH = −0.76). Always
  carry `humidity` as a feature. Unresolved until someone runs a **reversed-clock
  chicken** (`h1` at ~21:00). See `eda/FINDINGS.md` §4.
- One checkpoint = one data point. Hold out **whole chickens** for validation.
- The absolute baseline varies between days and rooms (open air ≫ closed box). That's
  normal; only flatness matters.
- Rough ranges over all 26 usable runs, MQ-135 `rel` at 2 min:
  - smell 1 (tawar) ≈ 0.60–1.08 (median 0.87),
  - smell 2–3 (asin/apek/kecut) ≈ 0.35–0.93,
  - smell 4–5 (pesing/rotten) ≈ 0.13–0.40.
  - The 2-min bands overlap. At **5 min** they do not: `rel300_135 < 0.285` splits
    smell ≥ 4 from the rest at 96 % held-out.

## Reading a log: diagnosis

| Symptom | Meaning |
|---|---|
| Smooth slow climb after the chicken goes in | Real chicken gas |
| **Step** (jump within 1–2 rows), both sensors at once, box closed | Lid/box touched, or a power/connection glitch |
| `# serial error … Device not configured` | USB dropped. **Even a split second cools the heaters.** The first rows after reconnecting dip (e.g. 50/28), then values drift for minutes. Baseline before + sample after the drop are **not comparable**: wait 15–20 min and take a new baseline. |
| Sample reads **below** baseline (`rel` > 1) | Not physically possible for a chicken, so the baseline was contaminated, or the sensor state changed (drop/cold) |
| A1 stays within ~1–2 counts of A0 for minutes while A0 drifts | **MQ-135 not connected: A1 is floating** and copies A0 (ADC sample-and-hold). "Both sensors move together" is NOT proof the wiring is OK. Test: unplug MQ-135 VCC for 10 s; if A1 doesn't drop toward 0, it's floating. |
| A sensor stuck at 1017–1023 with VREF 5 | Loose GND on that module |
| 1023 at VREF 1.1 | Saturated |
| Big jump between `x` and `s sample` | Happened while not recording. **Not a defect** — tested over 11 runs, they track smell better than clean runs. Keep the run |
| Very strong jump with an "asin" smell (e.g. A02 h3) | Chicken came straight from a closed container (built-up gas burst) |
| Weak response with a "kecut" smell (e.g. A02 h4) | Chicken was fanned first (gas blown away) |
| Humidity jumps (+5–10 %) when the chicken goes in | Normal. The MQ-135 is humidity-sensitive, so the DHT columns matter |
| Box reading creeps up while empty | Damp box (dry it fully), smell residue, or an outside source |

## .txt / .csv → JSON

**Always convert with the script — never retype the numbers by hand.**

```
python3 tools/log_to_json.py data/A03_20261020_200303.csv   # serial_log.py output
python3 tools/log_to_json.py data/h2.txt                    # old Serial Monitor paste
```

- Report every `warning:` to the user: they mean rows were skipped or missing (e.g. copy
  started late, Serial Monitor dropped lines). Don't silently accept them.
- The script fills:
  - the sensor arrays,
  - `interval_ms` (measured from the rows, not assumed: the DHT11 read blocks, so a logger
    tick is ~1500 ms, not 1000),
  - `vref` (1.1 or 5.0, back-computed from the raw + volt columns; raw counts are only
    comparable between runs at the same vref),
  - `lid_closed_at_s` (= baseline length in seconds),
  - `part: "WHOLE"`,
  - `chicken_id` (from the command ID).
- The user fills by hand:
  - `date` (RFC 3339, e.g. `2026-10-16T09:00:00+07:00`),
  - `supplier`,
  - `label` (`FRESH` / `UNFRESH` / `UNKNOWN`),
  - `author`,
  - `weight_g` (**grams**, e.g. 1158, not 1.158 — A01–A06 are all filled in). Note it is
    currently **one weighing per chicken, copied to every checkpoint**, so treat it as the
    *initial* weight. Per-checkpoint weighing needs a 0.1 g scale to be meaningful.
  - `hours_since_buy`.
- `temp` and `humidity` come from the DHT11 columns (sample phase). They stay `null` on logs
  recorded before the DHT11 was added.
- `hours_since_buy` must be ≥ 1. The first measurement is recorded as `1`, never `0`
  (`required` rejects 0 on a `float64`).
- Field names follow `internal/domain/gas.go` (`baseline_137`, not `baseline_mq_137`). If the
  domain changes, update `tools/log_to_json.py` to match.
- The user POSTs the JSON via Postman. Every field the user fills is required by `domain.Gas`.
- **Sensory fields in `domain.Gas`:**

  | Field | Values |
  |---|---|
  | `smell_level` | 1 normal/tawar · 2 asin/metallic · 3 apek/asem/kecut · 4 pesing/ammonia · 5 rotten |
  | `texture` | `BOUNCY` / `SOFT` (stays pushed in) / `MUSHY` (slimy) |
  | `color` | `PINK_PEACH` / `PEACH_YELLOWISH` / `PINK_YELLOWISH` / `PALE` / `GREYISH` |

  - Derive `label` from these: smell ≥ 3 or `SOFT`/`MUSHY` → `UNFRESH`.
  - `taste_notes` is the field name. `tasting_notes` is not a field and gets dropped.
    Taste is being phased out (unsafe).
- **Not yet in the domain**, so put them in `notes[]` as `key: value` until added:
  - `prep:` (`tray 1 min` standard / `from closed container` / `fanned`),
  - `quality:` (`ok`, `usb_drop_cold`, `mq135_floating`, `mq135_invalid`, `cold_sensor`,
    `damp_box`, `rewired_high_baseline`, `saturated`, `rel_above_1`). `gap_jump` was
    dropped: tested, no effect on the data.
  - `room:` (e.g. `no AC 33 °C`).
- **Before converting, check the log for:**
  - a restarted sample with the same ID (A02 h4 had an aborted sample glued in front),
  - lid spikes at the end of a baseline (A02 h1).

  Keep only the clean segment in the cleaned copy.

## 5-hour test (`enose_calibration`)

Serial Monitor drops old lines, so record with the terminal instead (Arduino IDE closed):

```
caffeinate -i sh -c '(stty 115200 raw; cat) < /dev/cu.usbmodemXXX | tee data/test_5jam.csv'
```
