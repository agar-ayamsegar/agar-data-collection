#!/usr/bin/env python3
"""
Serial logger for the e-nose (enose_calibration.ino).

Saves every line the Arduino prints into a CSV file with the real clock time
in front, so it can run for hours (the Arduino Serial Monitor can't save).

Run from the agar-data-collection folder:
    caffeinate -i python3 tools/serial_log.py                 # auto-find Arduino, name "enose"
    caffeinate -i python3 tools/serial_log.py chicken_6h      # custom name
    caffeinate -i python3 tools/serial_log.py chicken_6h /dev/cu.usbmodem1101
(caffeinate keeps the Mac awake while the logger runs.)

While it runs, type then press Enter:
    c            recalibrate (enose_calibration.ino)
    k            skip warm-up (enose_calibration.ino)
    s <phase> <id>  start a phase (enose_logger.ino), e.g.  s baseline A01  /  s sample A01
    x            stop logging (enose_logger.ino)
    m <note>     save a marker in the file, e.g.  m chicken in
    q            quit (Ctrl+C also works; everything logged is already saved)

Before starting:
  * Close the Arduino IDE Serial Monitor - only one program can use the port.
  * Opening the port RESTARTS the Arduino (warm-up + auto-calibration run
    again), so start with the box EMPTY and put the chicken in after it
    prints "# rs_air137: ...".

Output: data/<name>_<YYYYmmdd_HHMMSS>.csv
"""
import glob
import sys
import threading
import time
from pathlib import Path

try:
    import serial
except ImportError:
    sys.exit("pyserial is missing. Install it with:  python3 -m pip install --user pyserial")

BAUD = 115200


def now():
    return time.strftime("%Y-%m-%d %H:%M:%S")


def find_port():
    patterns = ["/dev/cu.usbmodem*", "/dev/cu.usbserial*", "/dev/cu.wchusbserial*",
                "/dev/ttyACM*", "/dev/ttyUSB*"]
    ports = sorted(p for pat in patterns for p in glob.glob(pat))
    if not ports:
        sys.exit("No Arduino port found. Is it plugged in? Check with:  ls /dev/cu.*")
    if len(ports) > 1:
        print("Several ports found, using the first one:", ", ".join(ports))
    return ports[0]


def main():
    name = sys.argv[1] if len(sys.argv) > 1 else "enose"
    port = sys.argv[2] if len(sys.argv) > 2 else find_port()

    out_dir = Path(__file__).resolve().parent.parent / "data"
    out_dir.mkdir(exist_ok=True)
    out = out_dir / f"{name}_{time.strftime('%Y%m%d_%H%M%S')}.csv"

    try:
        ser = serial.Serial(port, BAUD, timeout=1)
    except serial.SerialException as e:
        sys.exit(f"Can't open {port}: {e}\nIs the Arduino Serial Monitor still open? Close it and retry.")

    f = open(out, "a", buffering=1)
    lock = threading.Lock()
    stop = threading.Event()

    def save(line):
        with lock:
            f.write(line + "\n")
            f.flush()

    save(f"# logging started {now()} on {port}")
    print(f"Logging {port}  ->  {out}")
    print("Type: c / k (calibration sketch), s <phase> <id> / x (logger sketch), m <note> = marker, q = quit")

    def keyboard():
        for cmd in sys.stdin:
            cmd = cmd.strip()
            if cmd == "q":
                stop.set()
                return
            if cmd.startswith("m ") and len(cmd) > 2:
                save(f"# MARK {now()} {cmd[2:]}")
                print(f">> marker saved: {cmd[2:]}")
            elif cmd in ("c", "k"):
                ser.write(cmd.encode())
                save(f"# CMD {now()} sent '{cmd}'")
                print(f">> sent '{cmd}' to the Arduino")
            elif cmd.startswith("s ") or cmd == "x":      # enose_logger.ino commands (need a newline)
                ser.write((cmd + "\n").encode())
                save(f"# CMD {now()} sent '{cmd}'")
                print(f">> sent '{cmd}' to the Arduino")
            elif cmd:
                print(">> unknown command. Use: c, k, s <phase> <id>, x, m <note>, q")

    threading.Thread(target=keyboard, daemon=True).start()

    try:
        while not stop.is_set():
            raw = ser.readline()
            if not raw:
                continue
            line = raw.decode(errors="ignore").strip()
            if not line:
                continue
            if line.startswith("#"):
                save(f"{line}  [{now()}]")          # comment lines stay comments
            elif line.startswith("t_s,"):
                save("clock," + line)               # header row gets a clock column
            else:
                save(f"{now()},{line}")             # data row: clock,t_s,raw137,...
            print(now(), line)
    except KeyboardInterrupt:
        pass
    except serial.SerialException as e:
        save(f"# serial error {now()}: {e}")
        print("Serial error (Arduino unplugged?):", e)
    finally:
        save(f"# logging stopped {now()}")
        f.close()
        ser.close()
        print(f"Saved: {out}")


if __name__ == "__main__":
    main()
