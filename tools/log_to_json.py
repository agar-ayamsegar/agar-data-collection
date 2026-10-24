# python3 tools/log_to_json.py data/h1.txt
#   commands with an ID ("s baseline A01" ... "s sample A01" ... "s baseline A02" ...)
#     -> data/A01_h1.json, data/A02_h1.json  (chicken_id filled in)
#   commands without an ID ("s baseline" ... "s sample") -> one measurement -> data/h1.json
import json
import re
import statistics
from datetime import datetime, timedelta, timezone
import sys
from pathlib import Path

INTERVAL_MS = 1000   # fallback only; the real interval is measured from the rows (DHT11 reads block ~0.5 s)
TZ = timezone(timedelta(hours=7))                                   # WIB, where the data is collected
TIMESTAMP = re.compile(r"^\d{2}:\d{2}:\d{2}\.\d{3} -> ")            # Arduino Serial Monitor "show timestamp"
CLOCK = re.compile(r"^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})[ ,]")   # tools/serial_log.py file (,) or its console output ( )
MARKER = re.compile(r"^(?:>> marker saved: |m )(.+)$")              # the "m <note>" command and its enose_logger echo
NO_ID = "0"


def parse(path):
    measurements = {}  # id -> {"baseline": (mq137, mq135), "sample": (...)}, in log order
    last_t = {}
    last_clock = None
    warnings = []
    steps = []           # (line, cid, phase, prev_t, t) — gap check needs the measured interval first
    stops = []           # (line, clock, last_clock)
    vrefs = []           # ADC reference, back-computed from the raw + volt columns
    clipped = []
    cur_cid = None
    pending_notes = []   # markers seen before the first row of a measurement

    for n, line in enumerate(path.read_text().splitlines(), 1):
        line = TIMESTAMP.sub("", line.strip())
        clock = None
        if line[:1] == "x" and CLOCK.match(line[1:]):   # "x" echo glued onto the next row
            line = line[1:]
        cm = CLOCK.match(line)
        if cm:
            clock, line = cm.group(1), line[cm.end():]
        if line.startswith("# STOP") and clock and last_clock:
            stops.append((n, clock, last_clock))
        mk = MARKER.match(line)
        if mk:
            notes = measurements[cur_cid]["notes"] if cur_cid else pending_notes
            note = mk.group(1).strip()
            if not notes or notes[-1] != note:   # the command and its echo are the same note
                notes.append(note)
            continue
        if not line or line.startswith("#") or line.startswith(("t_s,", "clock,")):
            continue
        f = [x.strip() for x in line.split(",")]
        if len(f) not in (7, 9) or f[2] not in ("baseline", "sample"):
            warnings.append(f"line {n}: skipped, not a logger row: {line!r}")
            continue
        cid, phase = f[1], f[2]
        try:
            t, r137, r135 = float(f[0]), int(f[3]), int(f[5])
            v137, v135 = float(f[4]), float(f[6])
            # DHT11 columns print as "32 C, 49 %" -> two fields with units
            temp, hum = (float(f[7].rstrip(" C")), float(f[8].rstrip(" %"))) if len(f) == 9 else (None, None)
        except ValueError:
            warnings.append(f"line {n}: skipped, bad number: {line!r}")
            continue
        if not (0 <= r137 <= 1023 and 0 <= r135 <= 1023):
            warnings.append(f"line {n}: skipped, raw out of 0-1023: {line!r}")
            continue

        m = measurements.setdefault(cid, {"baseline": ([], [], [], []), "sample": ([], [], [], []),
                                          "date": clock or "", "notes": []})
        if cid != cur_cid:
            cur_cid = cid
            m["notes"].extend(pending_notes)
            pending_notes.clear()
        if phase == "baseline" and m["sample"][0]:
            who = "" if cid == NO_ID else f" for {cid}"
            sys.exit(f"line {n}: baseline{who} after its sample — measured twice? "
                     "Use a different ID per chicken (s baseline A01 / s baseline A02) or split the .txt")

        if 1023 in (r137, r135):
            clipped.append(n)
        vrefs += [v * 1023 / r for r, v in ((r137, v137), (r135, v135)) if r > 0]

        steps.append((n, cid, phase, last_t.get((cid, phase)), t))
        last_t[(cid, phase)] = t
        last_clock = clock or last_clock

        m[phase][0].append(r137)
        m[phase][1].append(r135)
        if temp is not None:
            m[phase][2].append(temp)
            m[phase][3].append(hum)

    deltas = [t - prev for _, _, _, prev, t in steps if prev is not None]
    interval_ms = round(statistics.median(deltas) * 1000, -1) if deltas else INTERVAL_MS
    gap = 1.5 * interval_ms / 1000
    for n, cid, phase, prev, t in steps:
        if prev is None and t > gap:
            warnings.append(f"line {n}: {cid} {phase} starts at t_s={t} — rows missing at the start?")
        elif prev is not None and t - prev > gap:
            warnings.append(f"line {n}: {cid} {phase} jumps t_s {prev} -> {t} — rows missing?")
    for n, clock, prev_clock in stops:
        lost = (datetime.fromisoformat(clock) - datetime.fromisoformat(prev_clock)).total_seconds()
        if lost > gap:
            warnings.append(f"line {n}: {lost:.0f} s between the last row ({prev_clock}) and STOP "
                            "— rows missing at the end (copy truncated?)")
    warnings.sort(key=lambda w: int(w.split()[1].rstrip(":")))

    # the Uno prints volts as raw * VREF / 1023, so VREF comes back out of any row
    vref = round(statistics.median(vrefs), 1) if vrefs else None
    if vrefs and max(vrefs) - min(vrefs) > 0.3:
        warnings.append(f"ADC reference is not constant ({min(vrefs):.2f}-{max(vrefs):.2f} V) "
                        "— two sketches in one .txt? split it")
    if clipped:
        warnings.append(f"{len(clipped)} row(s) hit raw 1023 (first at line {clipped[0]}) — the sensor "
                        f"clipped against the {vref} V reference; those readings are floored, not real")
    for cid, m in measurements.items():
        for ph in ("baseline", "sample"):
            a, b = m[ph][0], m[ph][1]
            if len(a) < 10:
                continue
            if min(statistics.pstdev(a), statistics.pstdev(b)) < 0.5:
                continue  # a flat channel says nothing either way
            d = [y - x for x, y in zip(a, b)]
            # shape, not magnitude: at vref 5.0 both sensors sit in the bottom 6% of the ADC,
            # so "a few counts apart" is normal. Crosstalk means the same curve twice.
            if statistics.correlation(a, b) > 0.99 and abs(statistics.mean(d)) < 10 * vref / 1.1:
                warnings.append(f"{cid} {ph}: MQ-135 is the same curve as MQ-137 (r>0.99, offset "
                                f"{statistics.mean(d):.1f} counts) — check A1 is really wired, "
                                "mq_135 may not be real data")
    return measurements, warnings, int(interval_ms), vref


def rfc3339(clock):
    # "2026-10-18 13:40:03" (local, from tools/serial_log.py) -> "2026-10-18T13:40:03+07:00"
    if not clock:
        return ""
    return datetime.fromisoformat(clock).astimezone(TZ).isoformat()


def to_json(cid, m, interval_ms, vref):
    (b137, b135, _, _), (s137, s135, temp, hum) = m["baseline"], m["sample"]
    out = json.dumps({
        "date": rfc3339(m.get("date", "")),
        "chicken_id": "" if cid == NO_ID else cid,
        "supplier": "",
        "part": "WHOLE",
        "weight_g": None,
        "hours_since_buy": None,
        "label": "",
        "author": "",
        "interval_ms": interval_ms,
        "vref": vref,
        "lid_closed_at_s": round(len(b137) * interval_ms / 1000),
        "baseline_137": b137,
        "baseline_135": b135,
        "mq_137": s137,
        "mq_135": s135,
        "temp": temp or None,
        "humidity": hum or None,
        "notes": m.get("notes", []),
        "taste_notes": "",   # emitted so the key is never hand-typed (it is taste_notes, not tasting_notes)
        "texture": "",       # BOUNCY / SOFT / MUSHY
        "color": "",         # PINK_PEACH / PEACH_YELLOWISH / PINK_YELLOWISH / PALE / GREYISH
        "smell_level": None, # 1-5
    }, indent=2)
    return re.sub(r"\[[\d,\s]+\]", lambda a: re.sub(r"\s+", "", a.group()), out) + "\n"


def main():
    if len(sys.argv) != 2:
        sys.exit("usage: python3 tools/log_to_json.py <log.txt>")
    src = Path(sys.argv[1])
    measurements, warnings, interval_ms, vref = parse(src)
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
        dst.write_text(to_json(cid, m, interval_ms, vref))
        print(f"{dst}: vref {vref} V, interval {interval_ms} ms, baseline {len(m['baseline'][0])} rows, sample {len(m['sample'][0])} rows")


if __name__ == "__main__":
    main()
