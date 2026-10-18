#!/usr/bin/env python3
"""
Analyze a logger CSV (from tools/serial_log.py) against reference baselines.

    python3 tools/analyze_run.py                     # newest data/*.csv
    python3 tools/analyze_run.py data/chicken_6h_X.csv

Prints a 10-minute summary and saves <csv>_analysis.png next to the CSV.
rel_vs_start = Rs / rs_air from the logger's own calibration
rel_vs_empty = Rs / empty-closed-box Rs from data/reference_baselines.json
  -> rel_vs_empty clearly below ~0.95 = more gas than an empty box (ammonia/amines likely)
"""
import json, sys
from pathlib import Path
import pandas as pd

root = Path(__file__).resolve().parent.parent
csv = Path(sys.argv[1]) if len(sys.argv) > 1 else max(list((root / "data").glob("*_2*.csv")) + list((root / "data" / "runs").glob("*/raw/*.csv")), key=lambda p: p.stat().st_mtime)
ref = json.loads((root / "data" / "reference_baselines.json").read_text())["empty_closed_box_20261017"]
E137, E135 = ref["rs137_mean"], ref["rs135_mean"]

rows, marks, rsair = [], [], None
for line in csv.read_text().splitlines():
    if line.startswith("# MARK"):
        p = line.split(" ", 4); marks.append((pd.Timestamp(p[2] + " " + p[3]), p[4]))
    elif line.startswith("# rs_air137"):
        rsair = line
    elif line and not line.startswith(("#", "clock")):
        p = line.split(",")
        if len(p) == 8:
            try: rows.append([pd.Timestamp(p[0]), int(p[1]), int(p[2]), float(p[3]), float(p[4]), int(p[5]), float(p[6]), float(p[7])])
            except ValueError: pass
df = pd.DataFrame(rows, columns=["clock","t_s","raw137","rs137","rel137","raw135","rs135","rel135"])
_proc = (csv.parent.parent / "processed" / (csv.stem + "_processed.csv")) if csv.parent.name == "raw" else csv.with_name(csv.stem + "_processed.csv")
if _proc.exists():   # use shake-corrected Rs from tools/process_run.py when available
    _p = pd.read_csv(_proc, parse_dates=["clock"])[["clock", "rs137_corr", "rs135_corr"]]
    df = df.merge(_p, on="clock", how="left")
    df["rs137"] = df.rs137_corr.fillna(df.rs137); df["rs135"] = df.rs135_corr.fillna(df.rs135)
    print("using shake-corrected Rs from", _proc.name)
df["rel137_vs_empty"] = df.rs137 / E137
df["rel135_vs_empty"] = df.rs135 / E135
start = df.clock.iloc[0]
df["min"] = (df.clock - start).dt.total_seconds() / 60

print(f"file: {csv.name}\n{rsair}\nempty-box reference: rs137={E137}  rs135={E135}")
print("markers:"); [print(f"  {(t-start).total_seconds()/60:6.1f} min  {m}") for t, m in marks]
g = df.groupby((df["min"] // 10) * 10)
s = g.agg(raw137=("raw137","mean"), rs137=("rs137","mean"), rel137=("rel137","mean"), r137e=("rel137_vs_empty","mean"),
          raw135=("raw135","mean"), rs135=("rs135","mean"), rel135=("rel135","mean"), r135e=("rel135_vs_empty","mean"))
print("\n min | raw137  rs137  rel_start  rel_EMPTY | raw135  rs135  rel_start  rel_EMPTY")
for m, r in s.iterrows():
    print(f"{m:4.0f} | {r.raw137:6.1f} {r.rs137:6.2f}   {r.rel137:.3f}     {r.r137e:.3f}   | {r.raw135:6.1f} {r.rs135:6.2f}   {r.rel135:.3f}     {r.r135e:.3f}")

import matplotlib; matplotlib.use("Agg"); import matplotlib.pyplot as plt
fig, ax = plt.subplots(2, 1, figsize=(10, 7), sharex=True)
for a, col, base, name in [(ax[0], "rs137", E137, "MQ-137"), (ax[1], "rs135", E135, "MQ-135")]:
    a.plot(df["min"], df[col].rolling(30, min_periods=1).mean(), lw=1.5, label=f"{name} Rs (30 s avg)")
    a.axhline(base, color="gray", ls="--", label=f"empty closed box = {base}")
    a.axhline(base * 0.95, color="red", ls=":", label="5% below empty box")
    for t, m in marks:
        x = (t - start).total_seconds() / 60; a.axvline(x, color="orange", alpha=.5)
        if a is ax[0]: a.text(x, a.get_ylim()[1], " " + m[:28], rotation=90, va="top", fontsize=7)
    a.set_ylabel("Rs (kOhm)  lower = more gas"); a.legend(fontsize=8, loc="lower left")
ax[1].set_xlabel("minutes since logger start")
fig.suptitle(csv.name); fig.tight_layout()
out_dir = Path(sys.argv[2]) if len(sys.argv) > 2 else (csv.parent.parent / "processed" if csv.parent.name == "raw" else csv.parent)
out_dir.mkdir(parents=True, exist_ok=True)
out = out_dir / (csv.stem + "_analysis.png"); fig.savefig(out, dpi=120)
print(f"\nplot: {out}")
