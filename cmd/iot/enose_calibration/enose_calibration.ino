// MQ-137: AO -> A0 | MQ-135: AO -> A1 | VCC -> 5V, GND -> GND
// Calibrate with the box EMPTY. Send "c" to re-calibrate.

const uint8_t PIN_MQ137 = A0;
const uint8_t PIN_MQ135 = A1;

const float VCC    = 5.0;
const float RL_137 = 1.0;
const float RL_135 = 2.0;

const int WARMUP_SEC  = 600; // warm up harusnya 10 menit biar sensor panas, kalau pertamakali beli malah dibiarin 48jam (katanya)
const int CAL_SAMPLES = 50;

float rsAir137 = 0;
float rsAir135 = 0;
unsigned long startMs = 0;

float rawToRs(int raw, float rl) {
  float v = raw * VCC / 1023;
  if (v < 0.01) v = 0.01;
  return rl * (VCC - v) / v;
}

float readRs(uint8_t pin, float rl) {
  return rawToRs(analogRead(pin), rl);
}

void calibrate() {
  Serial.println("# calibrating, keep box empty.");
  float sum137 = 0, sum135 = 0;
  for (int i = 0; i < CAL_SAMPLES; i++) {
    sum137 += readRs(PIN_MQ137, RL_137);
    sum135 += readRs(PIN_MQ135, RL_135);
    delay(100);
  }
  rsAir137 = sum137 / CAL_SAMPLES;
  rsAir135 = sum135 / CAL_SAMPLES;

  Serial.print("# rs_air137: ");
  Serial.print(rsAir137, 2);
  Serial.print("  rs_air135: ");
  Serial.println(rsAir135, 2);
  Serial.println("t_s,raw137,rs137,rel137,raw135,rs135,rel135");
  startMs = millis();
}

void warmUp() {
  for (int s = WARMUP_SEC; s > 0; s--) {
    if (s % 60 == 0) {
      Serial.print("# warm-up ");
      Serial.print(s / 60);
      Serial.println(" min left (k = skip)");
    }
    if (Serial.available() && Serial.read() == 'k') break;
    delay(1000);
  }
}

void setup() {
  Serial.begin(115200);
  warmUp();
  calibrate();
}

void loop() {
  if (Serial.available() && Serial.read() == 'c') calibrate();

  int raw137 = analogRead(PIN_MQ137);
  int raw135 = analogRead(PIN_MQ135);
  float rs137 = rawToRs(raw137, RL_137);
  float rs135 = rawToRs(raw135, RL_135);

  Serial.print((millis() - startMs) / 1000);
  Serial.print(',');
  Serial.print(raw137);
  Serial.print(',');
  Serial.print(rs137, 2);
  Serial.print(',');
  Serial.print(rs137 / rsAir137, 3);
  Serial.print(',');
  Serial.print(raw135);
  Serial.print(',');
  Serial.print(rs135, 2);
  Serial.print(',');
  Serial.println(rs135 / rsAir135, 3);

  delay(1000);
}
