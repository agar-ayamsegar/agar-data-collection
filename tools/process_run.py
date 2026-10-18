#!/usr/bin/env python3
"""
Make a processed copy of a logger CSV (original is never modified).

    python3 tools/process_run.py data/runs/<run>/raw/<file>.csv        # writes to data/runs/<run>/processed/
    python3 tools/process_run.py <file>.csv <out_dir>                 # or choose the output folder

Adds: rel vs empty-box baseline, shake (step) correction, smoothing,
MQ-137/MQ-135 ratio, hours since buy, minutes since chicken in, label.
Run-specific settings are in RUN below.
"""
import sys
from pathlib import Path
import pandas as pd

RUN = {
    "empty_rs137": 27.53, "empty_rs135": 38.65,           # empty closed box (reference_baselines.json)
    "bought": "2026-10-17 19:30:00",
    "chicken_in": "2026-10-17 20:01:32",
    "shakes": ["2026-10-17 23:33:23"],                     # instant steps on both sensors
}

src = Path(sys.argv[1])
rows, marks = [], []
for line in src.read_text().splitlines():
    if line.startswith("# MARK"):
        p = line.split(" ", 4); marks.append((pd.Timestamp(p[2] + " " + p[3]), p[4]))
    elif line and not line.startswith(("#", "clock")):
        p = line.split(",")
        if len(p) == 8:
            try: rows.append([pd.Timestamp(p[0]), int(p[1]), int(p[2]), float(p[3]), float(p[4]), int(p[5]), float(p[6]), float(p[7])])
            except ValueError: pass
d = pd.DataFrame(rows, columns=["clock","t_s","raw137","rs137","rel137_logger","raw135","rs135","rel135_logger"]).set_index("clock")

# shake correction: multiply Rs after each step so the 2 min after match the 2 min before
d["after_shake"] = 0
d["rs137_corr"], d["rs135_corr"] = d.rs137, d.rs135
for s in RUN["shakes"]:
    t = pd.Timestamp(s)
    pre = d.loc[t - pd.Timedelta("2min"): t - pd.Timedelta("3s")]
    post = d.loc[t + pd.Timedelta("7s"): t + pd.Timedelta("2min")]
    k137 = pre.rs137_corr.mean() / post.rs137_corr.mean()
    k135 = pre.rs135_corr.mean() / post.rs135_corr.mean()
    m = d.index >= t
    d.loc[m, "rs137_corr"] *= k137; d.loc[m, "rs135_corr"] *= k135; d.loc[m, "after_shake"] += 1
    print(f"shake {s}: rs137 x{k137:.3f}, rs135 x{k135:.3f}")
    d.loc[t - pd.Timedelta("1s"): t + pd.Timedelta("6s"), ["rs137_corr", "rs135_corr"]] = float("nan")  # drop the jump seconds

d["rel137_vs_empty"] = (d.rs137_corr / RUN["empty_rs137"]).round(4)
d["rel135_vs_empty"] = (d.rs135_corr / RUN["empty_rs135"]).round(4)
d["rel137_vs_empty_uncorr"] = (d.rs137 / RUN["empty_rs137"]).round(4)
d["rel135_vs_empty_uncorr"] = (d.rs135 / RUN["empty_rs135"]).round(4)
d["rel137_smooth30s"] = d.rel137_vs_empty.rolling("30s").mean().round(4)
d["rel135_smooth30s"] = d.rel135_vs_empty.rolling("30s").mean().round(4)
d["rel137_smooth10min"] = d.rel137_vs_empty.rolling("10min").mean().round(4)
d["rel135_smooth10min"] = d.rel135_vs_empty.rolling("10min").mean().round(4)
d["ratio_137_135"] = (d.rel137_smooth10min / d.rel135_smooth10min).round(4)
d["hours_since_buy"] = ((d.index - pd.Timestamp(RUN["bought"])).total_seconds() / 3600).round(4)
d["minutes_since_chicken_in"] = ((d.index - pd.Timestamp(RUN["chicken_in"])).total_seconds() / 60).round(2)
d["label"] = ""
for t, m in marks:
    d.loc[d.index >= t, "label"] = m
def zone(r):
    if pd.isna(r.rel137_smooth10min): return ""
    lo = min(r.rel137_smooth10min, r.rel135_smooth10min)
    if r.rel137_smooth10min < 0.75 and r.ratio_137_135 < 1: return "ROTTEN"
    if lo < 0.85: return "RETURN"
    if lo < 0.95: return "CHECK"
    return "FRESH"
d["zone"] = d.apply(zone, axis=1)

out_dir = Path(sys.argv[2]) if len(sys.argv) > 2 else (src.parent.parent / "processed" if src.parent.name == "raw" else src.parent)
out_dir.mkdir(parents=True, exist_ok=True)
out = out_dir / (src.stem + "_processed.csv")
d.reset_index().to_csv(out, index=False)
print(f"saved {out}  rows={len(d)}  last={d.index[-1]}")
