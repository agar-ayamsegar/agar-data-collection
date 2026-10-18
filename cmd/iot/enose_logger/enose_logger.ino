// MQ-137: AO -> A0 | MQ-135: AO -> A1 | VCC -> 5V, GND -> GND | VREF -> 1.1

const uint8_t PIN_MQ137 = A0;
const uint8_t PIN_MQ135 = A1;

const float VCC   = 5.0;
const int   N_AVG = 20;
const unsigned long INTERVAL_MS = 1000;
const float VREF = 1.1; // for measurement

bool logging = true;
String phase = "none";
String id = "0";
unsigned long startMs = 0, lastMs = 0;

int readAvg(uint8_t pin) {
  long sum = 0;
  for (int i = 0; i < N_AVG; i++) {
    sum += analogRead(pin);
    delay(2);
  }
  return sum / N_AVG;
}

void handleSerial() {
  if (!Serial.available()) return;
  String cmd = Serial.readStringUntil('\n');
  cmd.trim();

  if (cmd.startsWith("s ")) {
    int sp = cmd.indexOf(' ', 2);
    phase = (sp > 0) ? cmd.substring(2, sp) : cmd.substring(2);
    id    = (sp > 0) ? cmd.substring(sp + 1) : "0";
    logging = true;
    startMs = millis();
    Serial.println("# START");
  } else if (cmd == "x") {
    logging = false;
    Serial.println("# STOP");
  }
}

void printRow() {
  int raw137 = readAvg(PIN_MQ137);
  int raw135 = readAvg(PIN_MQ135);

  Serial.print((millis() - startMs) / 1000.0, 1);
  Serial.print(',');
  Serial.print(id);
  Serial.print(',');
  Serial.print(phase);
  Serial.print(',');
  Serial.print(raw137);
  Serial.print(',');
  Serial.print(raw137 * VREF / 1023, 3);
  Serial.print(',');
  Serial.print(raw135);
  Serial.print(',');
  Serial.println(raw135 * VREF / 1023, 3);
}

void setup() {
  Serial.begin(115200);
  analogReference(INTERNAL);
  Serial.println("# READY");
  Serial.println("t_s,sample_id,label,raw137,v137,raw135,v135");
}

void loop() {
  handleSerial();
  if (!logging || millis() - lastMs < INTERVAL_MS) return;
  lastMs = millis();
  printRow();
}
