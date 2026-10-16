# python3 tools/log_to_json.py data/h1.txt
#   commands with an ID ("s baseline A01" ... "s sample A01" ... "s baseline A02" ...)
#     -> data/A01_h1.json, data/A02_h1.json  (chicken_id filled in)
#   commands without an ID ("s baseline" ... "s sample") -> one measurement -> data/h1.json
import json
import re
import sys
from pathlib import Path

INTERVAL_MS = 1000
TIMESTAMP = re.compile(r"^\d{2}:\d{2}:\d{2}\.\d{3} -> ")
NO_ID = "0"


def parse(path):
    measurements = {}  # id -> {"baseline": (mq137, mq135), "sample": (...)}, in log order
    last_t = {}
    warnings = []

    for n, line in enumerate(path.read_text().splitlines(), 1):
        line = TIMESTAMP.sub("", line.strip())
        if not line or line.startswith("#") or line.startswith("t_s,"):
            continue
        f = line.split(",")
        if len(f) != 7 or f[2] not in ("baseline", "sample"):
            warnings.append(f"line {n}: skipped, not a logger row: {line!r}")
            continue
        cid, phase = f[1], f[2]
        try:
            t, r137, r135 = float(f[0]), int(f[3]), int(f[5])
        except ValueError:
            warnings.append(f"line {n}: skipped, bad number: {line!r}")
            continue
        if not (0 <= r137 <= 1023 and 0 <= r135 <= 1023):
            warnings.append(f"line {n}: skipped, raw out of 0-1023: {line!r}")
            continue

        m = measurements.setdefault(cid, {"baseline": ([], []), "sample": ([], [])})
        if phase == "baseline" and m["sample"][0]:
            who = "" if cid == NO_ID else f" for {cid}"
            sys.exit(f"line {n}: baseline{who} after its sample — measured twice? "
                     "Use a different ID per chicken (s baseline A01 / s baseline A02) or split the .txt")

        prev = last_t.get((cid, phase))
        if prev is None and t > 1.5 * INTERVAL_MS / 1000:
            warnings.append(f"line {n}: {cid} {phase} starts at t_s={t} — rows missing at the start?")
        if prev is not None and t - prev > 1.5 * INTERVAL_MS / 1000:
            warnings.append(f"line {n}: {cid} {phase} jumps t_s {prev} -> {t} — rows missing?")
        last_t[(cid, phase)] = t

        m[phase][0].append(r137)
        m[phase][1].append(r135)

    return measurements, warnings


def to_json(cid, m):
    (b137, b135), (s137, s135) = m["baseline"], m["sample"]
    out = json.dumps({
        "date": "",
        "chicken_id": "" if cid == NO_ID else cid,
        "supplier": "",
        "part": "WHOLE",
        "weight_g": None,
        "hours_since_buy": None,
        "label": "",
        "author": "",
        "interval_ms": INTERVAL_MS,
        "lid_closed_at_s": len(b137) * INTERVAL_MS // 1000,
        "baseline_137": b137,
        "baseline_135": b135,
        "mq_137": s137,
        "mq_135": s135,
        "temp": None,
        "humidity": None,
    }, indent=2)
    return re.sub(r"\[[\d,\s]+\]", lambda a: re.sub(r"\s+", "", a.group()), out) + "\n"


def main():
    if len(sys.argv) != 2:
        sys.exit("usage: python3 tools/log_to_json.py <log.txt>")
    src = Path(sys.argv[1])
    measurements, warnings = parse(src)
    if not measurements:
        sys.exit(f"{src}: no logger rows found")
    if NO_ID in measurements and len(measurements) > 1:
        sys.exit(f"{src}: some commands have an ID and some don't — use an ID on every s command")

    incomplete = [cid for cid, m in measurements.items() if not m["baseline"][0] or not m["sample"][0]]
    if incomplete:
        sys.exit(f"{src}: missing baseline or sample rows for: {', '.join(incomplete)}")

    for w in warnings:
        print("warning:", w)
    for cid, m in measurements.items():
        name = src.stem if cid == NO_ID else f"{cid}_{src.stem}"
        dst = src.with_name(name + ".json")
        dst.write_text(to_json(cid, m))
        print(f"{dst}: baseline {len(m['baseline'][0])} rows, sample {len(m['sample'][0])} rows")


if __name__ == "__main__":
    main()
