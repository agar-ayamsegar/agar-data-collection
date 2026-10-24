# EDA findings — 31 checkpoints, 6 chickens (A01–A06)

**Date of analysis:** 2026-10-22 · **Data:** `cmd/iot/data/*/*.json` (A01 18 Oct → A06 21 Oct)
**Interactive report:** https://claude.ai/artifact/5GR6FM2fGUr6wzuT4KLhtm
**Regenerate:** `python3 tools/eda_report.py` → `eda/index.html` · `--json` for numbers · `--demo` for the self-check

Stdlib only, no pandas. Re-run it after every new chicken; every number below comes out of it.

---

## 1. What changes in the routine

| Change | Why | Evidence |
|---|---|---|
| **Sample for 5 min, not 2** | `rel300_135` is 96 % held-out vs 88 % for `rel120_135` | §4 |
| **Always record an h1 checkpoint** | It is the per-chicken reference that makes MQ-137 usable | §3 |
| **Stop flagging `gap_jump`** | Tested, no effect. Gap runs track smell *better* | §5 |
| **Record `humidity` as a model feature** | It is confounded with age and must be reported alongside | §4 |
| **Use two rules, not two sensors** | MQ-135 absolute + the 137/135 ratio. Agree → call it; disagree → ULANGI SCAN | §2b |

## 2. Which feature to use

`rel = Rs_sample / Rs_baseline`, baseline = median of the last 20 flat baseline rows.
Lower = more gas. 26 runs at VREF 5.0 (A01 excluded: VREF 1.1, two runs railed the ADC).

| feature | r vs smell | ρ vs smell | best threshold | in-sample | **leave-one-chicken-out** |
|---|---|---|---|---|---|
| **`rel300_135`** | −0.86 | −0.88 | **< 0.285** | 100 % (n=23) | **96 %** |
| `rel120_135` | −0.85 | −0.84 | < 0.420 | 96 % (n=26) | 88 % |
| `relmin_135` | −0.83 | −0.86 | < 0.300 | 96 % (n=26) | 88 % |
| `ratio300` = `rel300_137`/`rel300_135` | +0.82 | +0.84 | **> 2.010** | 100 % (n=23) | **96 %** |
| `ratio120` | +0.67 | — | > 2.030 | 92 % (n=26) | 88 % |
| `rel120_137` | −0.44 | −0.42 | < 0.625 | 81 % (n=26) | 69 % |

The `ratio` features fire **high** (they rise with spoilage); every `rel` feature fires
**low**. Combining the two sensors is covered in §2b.

Target is `smell_level >= 4`, **not** `label` — `label` is derived from the clock, so
predicting it just re-predicts `hours_since_buy`.

Leave-one-chicken-out = fit the threshold on 5 chickens, test on the 6th. That is the
number that survives a new chicken; in-sample is always optimistic.

The only in-sample miss anywhere is **A02_h3** at 120 s, and the log already explains it:
that chicken came straight out of a closed container, so a burst of built-up gas hit the
box at once.

## 2b. Using BOTH sensors — what works and what does not

Tested with the same leave-one-chicken-out protocol, 5-min reading, n=23.

| rule | held-out | verdict |
|---|---|---|
| MQ-135 alone (`rel300_135 < 0.285`) | 96 % | baseline |
| MQ-137 alone (`rel300_137 < t`) | 83 % | worse |
| **BOTH must fire** (`135 AND 137`) | 96 % | **no gain** — degenerates to MQ-135 |
| **EITHER fires** (`135 OR 137`) | 96 % | no gain (and only 81 % at 2 min — *worse*) |
| **ratio `rel300_137 / rel300_135 > 2.01`** | **96 %** | **works, and errors differ** |

**An AND/OR rule on the two raw `rel` values buys nothing.** MQ-137's mistakes are a
superset of MQ-135's, so the combination just falls back on MQ-135.

**The ratio does work.** It rises as the chicken spoils, because MQ-135 falls faster than
MQ-137. r = +0.82 against smell. Two properties make it worth having:

1. **Its errors are different.** MQ-135 alone misses A02_h3; the ratio misses A04_h5. Two
   independent 96 % rules, not one rule twice.
2. **It is less tangled with humidity** — r +0.70 vs −0.80 at 5 min (+0.56 vs −0.78 at
   2 min). Not clean, but better. §4 still stands.
3. **It needs no h1 reference.** Both `rel` values are already baseline-normalised inside
   the run, so the ratio works on a single scan. The h1 trick in §3 is a lab tool;
   **the ratio is the one that survives into the handheld**, where you scan a chicken
   once and have no history for it.

### The useful two-sensor rule: agreement = confidence

| | |
|---|---|
| Rule A | `rel300_135 < 0.285` (MQ-135 absolute) |
| Rule B | `rel300_137 / rel300_135 > 2.01` (both sensors, as a ratio) |

| outcome | runs | result |
|---|---|---|
| **A and B agree** | 21 / 23 | **21/21 correct — zero errors** |
| **A and B disagree** | 2 / 23 (9 %) | A02_h3, A04_h5 — exactly the two hard runs |

This is the honest version of "if both, we proceed": don't AND the two *sensors*, AND the
two *rules*, and send the disagreements to **ULANGI SCAN** — which `AGENTS.md` §9 already
specifies as the product's third outcome.

⚠ **Caveat:** 21/21 is 21 runs from 5 chickens with only 2 disagreements. The thresholds
were fitted leave-one-chicken-out, so the protocol is honest, but "100 %" on a sample this
small is encouraging, not proven. Both rules still sit under the §4 humidity confound.

## 3. MQ-137 is not weak — it has no absolute scale

It responds as hard as MQ-135 (min `rel` 0.14 vs 0.13; median 0.68 vs 0.60). The low
*pooled* correlation means "0.65 means a different thing on A03 than on A05", not
"it cannot smell".

Spearman ρ against hours computed **inside each chicken**, which cancels that chicken's
own offset:

| feature | pooled | A02 | A03 | A04 | A05 | A06 | mean within | rescaled by that chicken's h1 |
|---|---|---|---|---|---|---|---|---|
| `relmin_137` | −0.54 | **+0.40** | −0.70 | −0.90 | −0.83 | −0.66 | −0.54 | **−0.71** |
| `rel120_137` | −0.42 | +0.20 | −0.30 | −0.90 | −0.71 | −0.66 | −0.47 | −0.59 |
| `relmin_135` | −0.85 | −0.60 | −1.00 | −1.00 | −0.66 | −0.89 | −0.83 | −0.84 |
| `rel120_135` | −0.84 | 0.00 | −0.70 | −1.00 | −0.89 | −0.89 | −0.69 | −0.82 |

Rescaling by h1 lifts `relmin_137` from −0.54 to −0.71. MQ-135 barely moves (−0.85 →
−0.84) because it never needed the help.

**A02 is the only chicken that inverts**, and the log says why: h3 from a closed
container, h4 fanned first. Two known handling errors, one chicken.

**Rule:** MQ-135 calls the hour on an absolute threshold. MQ-137 is a *relative* sensor —
judge it against that chicken's own h1 reading, never as a standalone absolute cut-off.
Physically consistent: MQ-137 is tuned for ammonia, which only arrives at smell 4–5
(`pesing`), while MQ-135 answers the whole VOC mix and the `apek`/`asem` stage comes early.

## 4. ⚠ The humidity confound — unresolved, and it gates everything

Box humidity climbs ~40 % → ~80 % inside **every single chicken**, never once reversing.

| | r | ρ |
|---|---|---|
| humidity ~ hours | **+0.94** | +0.98 |
| humidity ~ smell | +0.90 | +0.92 |
| `rel120_135` ~ humidity | −0.78 | −0.76 |
| `rel120_135` ~ hours | −0.81 | −0.77 |

MQ-135 is humidity-sensitive. Humidity and age move together almost perfectly, so
"`rel` tracks spoilage" and "`rel` tracks water vapour" fit every number here equally well.

### It is the ROOM, not the chicken

Splitting each run's humidity into the box state *before* the chicken went in and the rise
*during* the sample separates the two sources:

| | range | r vs hours |
|---|---|---|
| **box RH at start** (room / box state) | **35 % → 77 %** | **+0.91** |
| **rise during the sample** (the chicken itself) | +1 % → +13 % | +0.64 |

The climb is overwhelmingly the box starting damper each hour, not the chicken breathing
out. **A05 and A06 ran in parallel on 2026-10-21, same room**, and their box humidity at
matched clock times agrees to within 1–6 points:

| clock | A05 | A06 |
|---|---|---|
| 12:44 / 13:04 | 39 % | 35 % |
| 14:28 / 14:49 | 47 % | 48 % |
| 18:41 / 19:09 | 63 % | 67 % |
| 20:44 / 21:12 | 71 % | 75 % |

corr(box RH at start, **clock time**) = **+0.97**. Two independent chickens cannot drive
one shared curve — so the driver is **the evening room**, and the box never drying out
between checkpoints.

### The baseline does not cancel it

The baseline is taken in the same box seconds before the sample, so box humidity *should*
divide out of `rel`. It does not, fully:

| feature | vs box RH at start | vs the chicken's own rise |
|---|---|---|
| `rel300_135` | **−0.76** | −0.58 |
| `rel120_135` | **−0.74** | −0.56 |

`rel` still tracks the box state more strongly than the chicken's contribution. So the
confound survives the ratio.

This does not mean the model is wrong. It means **this dataset cannot prove it right** —
and it localises the problem to the room, which is fixable.

### The experiment that settles it

**Run one chicken on a reversed clock.** Buy it in the evening and take `h1` at ~21:00,
when ambient RH is already ~75 %.

- If `h1` still reads fresh (`rel` ≈ 0.85) in a humid room → humidity is **not** driving
  the signal, and the whole result stands.
- If `h1` reads like a spoiled chicken → the model has been reading the clock, not the meat.

This beats the wet-cloth control: it is a real chicken, it costs one normal run, and it
breaks the age↔clock confound directly instead of only testing the sensor.

Cheaper partial checks, still worth doing: a **wet-cloth control** (damp cloth, ~75 % RH,
empty box, one full checkpoint) and **drying the box properly** between checkpoints — 2 min
of airing is clearly not enough to go from 35 % to 77 % and back.

Temperature is a weaker second confound (r = −0.62 with smell), drifting 34 → 31 °C.

## 5. `gap_jump` was tested and is NOT a data-quality problem

11 of the 26 usable runs have one. Previously flagged; the flag was never tested.

| | gap_jump (n=11) | no gap (n=15) |
|---|---|---|
| ρ (`rel120_135` vs smell) | **−0.91** | −0.80 |
| median `rel`, smell 1 | 0.86 | 0.87 |
| median `rel`, smell 5 | 0.17 | 0.20 |

Gap runs track smell *better* and match at every smell level. The chicken is already
sealed in the box during the gap, so no gas is lost — only the sample clock starts late,
and the curve has flattened by 120 s. Removed from the quality flags.

Typing `x` before `s sample` is still untidy, but it does not invalidate a run.

## 6. Real data-quality problems (these do kill a run)

| Problem | Runs | Why fatal |
|---|---|---|
| ADC railed at 1023 | A01_h5, A01_h6 (MQ-135, VREF 1.1) | Magnitude censored, not measured |
| `rel` > 1.0 — impossible | A02_h2 (135: 1.08), A03_h4 (137: 1.28) | Baseline was contaminated |
| ~~`weight_g` placeholder~~ | ~~all 31~~ | **Fixed 2026-10-22.** Note it is one weighing per chicken copied to every checkpoint, so it is the *initial* weight. Not currently used as a feature — no mechanism and only 6 birds. |
| No DHT11 | all 5 A01 runs | `temp`/`humidity` are `null` |
| Sample < 300 s | A03_h3, A04_h1, A04_h8 | Cannot supply `rel300`, the best feature |

**VREF split:** A01 at 1.1, everything after at 5.0. Raw counts are ~4.5× apart and must
never be pooled — but `rel` divides the scaling out, so the 3 unsaturated A01 runs stay
comparable. All fitting above uses the 26 VREF 5.0 runs only (the conservative choice).

## 7. Honest status

31 checkpoints, 26 usable, 6 chickens. Both sensors respond to a spoiling chicken; one of
them (MQ-135) does so on a scale that transfers to a new chicken, separating fresh from
spoiled at 96 % held-out. That is a real result at this sample size.

It is **not yet a validated result**:

- humidity tracks age too closely to rule out (§4),
- smell 4–5 holds only 7 of 26 usable runs, from 4 chickens,
- nothing was ever recorded past hour 8,
- every chicken but A01 came from the same supplier on consecutive days.

Treat it as a promising pilot, not a model. `AGENTS.md` §10 still stands: a real model
needs 10–20+ chickens.
