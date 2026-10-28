// Agar sensor box v2 — parametric model of all seven layouts + printable parts.
// Open in OpenSCAD (https://openscad.org), pick `design` and `part` in the Customizer,
// press F5 to preview, F6 + "Export STL" for a printable part.
//
// Units: millimetres. Axes: x = box length (left → right), y = box width (front → back),
// z = up. The box floor (inside) is z = 0, centred on x = y = 0.
//
// Designs (same letters as design/box-airflow-3d.html):
//   NOW  still box, sensors under the lid, chicken on the floor
//   A    stir fan in a top corner, sensors on the side wall, rack
//   B    side loop tube, sensors + fan in the tube outside the box
//   C    through-flow: filter inlet low, outlet high → sensor chamber → fan
//   D    hold · mix · flush: stir fan + exhaust fan + filter inlet (recommended to build first)
//   E    fan under the rack blowing up + V-hood, sensors in the hood throat
//   F    V-hood chimney through the lid, fan on top after the sensors (best signal)

/* [What to show] */
design = "F"; // [NOW, A, B, C, D, E, F]
part = "assembly"; // [assembly, throat_collar, throat_collar_E, sensor_wall_mount, rack_foot, filter_port, tube_port, fan_mount_40, hood_net]
show_chicken = true;
cutaway = true;      // hide the front wall so you can see inside

/* [Box, inside size] */
box_L = 400;   // length (x)
box_W = 300;   // width  (y)
box_H = 300;   // height (z)
wall  = 3;

/* [Rack and tray] */
rack_h = 30;       // grid height above the floor
rack_L = 320;
rack_W = 240;
rod = 3;           // grid rod diameter
rack_pitch = 20;
tray_h = 15;

/* [Chicken placeholder, whole ~1.1 kg] */
chick = [250, 150, 110];

/* [Sensors] */
mq_pcb = [32, 22, 1.6];  // MQ-137 / MQ-135 module board
mq_can_d = 20;           // sensor can (mesh) diameter
mq_can_h = 15;
dht = [16, 12, 6];       // DHT11 body
standoff = 25;           // wall mount distance from the wall

/* [Fans] */
fan = 40;          // 40 mm 5 V fan
fan_t = 10;
fan_hole = 32;     // screw spacing

/* [Hood (E, F)] */
hood_gap = 25;     // clearance above the chicken
hood_h = 70;       // funnel height
mouth = [270, 190];
throat = 56;       // square throat, inside
collar_wall = 2.4;
collar_h = 80;     // printed collar for F (runs up through the lid)
collar_h_E = 45;   // shorter collar for E (stays inside the box)

/* [Ports and filter] */
port_d = 16;       // tube outside diameter (B)
filter_d = 40;     // carbon filter cartridge diameter
filter_len = 60;

$fn = 48;

// ---------- derived ----------
chick_z0  = rack_h + rod;                 // chicken sits on the rack
chick_top = chick_z0 + chick[2];
hood_z0   = chick_top + hood_gap;          // funnel mouth
hood_z1   = hood_z0 + hood_h;              // funnel top / throat start
collar_o  = throat + 2*collar_wall;
xL = -box_L/2; xR = box_L/2; yF = -box_W/2; yB = box_W/2;

has_rack = design != "NOW";

// ---------- colours ----------
C_BOX   = [0.62, 0.78, 0.86, 0.18];
C_LID   = [0.62, 0.78, 0.86, 0.12];
C_RACK  = [0.72, 0.74, 0.76];
C_TRAY  = [0.86, 0.88, 0.90];
C_CHICK = [0.90, 0.68, 0.52];
C_PCB   = [0.10, 0.45, 0.25];
C_CAN   = [0.55, 0.57, 0.58];
C_DHT   = [0.20, 0.45, 0.85];
C_FAN   = [0.05, 0.42, 0.38];
C_PRINT = [0.95, 0.60, 0.20];   // printed parts
C_HOOD  = [0.35, 0.70, 0.60, 0.55];
C_TUBE  = [0.55, 0.60, 0.62, 0.55];
C_FILT  = [0.15, 0.15, 0.15];

// =====================================================================
// Shared hardware
// =====================================================================
module box_shell() {
  color(C_BOX) difference() {
    translate([xL-wall, yF-wall, -wall]) cube([box_L+2*wall, box_W+2*wall, box_H+wall]);
    translate([xL, yF, 0]) cube([box_L, box_W, box_H+1]);
    if (cutaway) translate([xL-wall-1, yF-wall-1, -wall-1]) cube([box_L+2*wall+2, wall+1.01, box_H+wall+2]);
    wall_holes();
  }
}

module lid() {
  color(C_LID) difference() {
    translate([xL-wall, yF-wall, box_H]) cube([box_L+2*wall, box_W+2*wall, wall]);
    lid_holes();
  }
}

module rack() {
  color(C_RACK) {
    for (x = [-rack_L/2 : rack_pitch : rack_L/2])
      translate([x, -rack_W/2, rack_h]) rotate([-90,0,0]) cylinder(d=rod, h=rack_W, $fn=10);
    for (y = [-rack_W/2 : rack_pitch*2 : rack_W/2])
      translate([-rack_L/2, y, rack_h+rod]) rotate([0,90,0]) cylinder(d=rod, h=rack_L, $fn=10);
  }
  for (sx=[-1,1], sy=[-1,1]) translate([sx*(rack_L/2-15), sy*(rack_W/2-10), 0]) color(C_PRINT) rack_foot();
}

module drip_tray(hole=false) {
  color(C_TRAY) difference() {
    translate([-rack_L/2-5, -rack_W/2-5, 0]) cube([rack_L+10, rack_W+10, tray_h]);
    translate([-rack_L/2-3, -rack_W/2-3, 1.5]) cube([rack_L+6, rack_W+6, tray_h]);
    if (hole) translate([0,0,-1]) cylinder(d=fan+6, h=5);
  }
}

module chicken(z0) {
  if (show_chicken) color(C_CHICK) translate([0, 0, z0 + chick[2]/2]) scale([chick[0]/2, chick[1]/2, chick[2]/2]) sphere(1, $fn=40);
}

module mq_module() { // board flat in x-y, can pointing +z
  color(C_PCB) translate([-mq_pcb[0]/2, -mq_pcb[1]/2, 0]) cube(mq_pcb);
  color(C_CAN) translate([0,0,mq_pcb[2]]) cylinder(d=mq_can_d, h=mq_can_h);
}
module dht11() { color(C_DHT) translate([-dht[0]/2, -dht[1]/2, 0]) cube(dht); }

module fan40() { // airflow along +z
  color(C_FAN) difference() {
    translate([-fan/2, -fan/2, 0]) cube([fan, fan, fan_t]);
    translate([0,0,-1]) cylinder(d=fan-3, h=fan_t+2);
  }
  color(C_FAN) cylinder(d=14, h=fan_t);
  color(C_FAN) for (a=[0:72:359]) rotate([0,0,a]) translate([0,-2,2]) rotate([0,20,0]) cube([fan/2-3, 4, 1.2]);
}

module tube(p1, p2, d=port_d) {
  color(C_TUBE) hull() { translate(p1) sphere(d=d, $fn=16); translate(p2) sphere(d=d, $fn=16); }
}

module sensor_cluster() { // two MQ + DHT on a wall mount, facing +x (into the box from the left wall)
  rotate([0,90,0]) {
    translate([0, -20, 0]) mq_module();
    translate([0,  20, 0]) mq_module();
    translate([-28, 0, 0]) dht11();
  }
}

// =====================================================================
// Printable parts
// =====================================================================

// Collar between the V-hood and the fan: holds both MQ cans and the DHT11 inside the throat.
// Print upright, no supports (the windows are vertical slots).
module throat_collar(collar_h=collar_h, fan_plate=true) {
  difference() {
    union() {
      translate([-collar_o/2, -collar_o/2, 0]) cube([collar_o, collar_o, collar_h]);
      // bottom flange: the hood panels tape or glue under it
      translate([-collar_o/2-8, -collar_o/2-8, 0]) cube([collar_o+16, collar_o+16, 3]);
      // top fan plate
      // (the collar's own top is the fan plate when fan_plate = true)
    }
    translate([-throat/2, -throat/2, -1]) cube([throat, throat, (fan_plate ? collar_h-3 : collar_h) + 1.01]);
    if (fan_plate) translate([0,0,collar_h-5]) cylinder(d=fan-3, h=10);   // air goes up into the fan
    // MQ cans poke in through round holes on two opposite walls, boards stay outside
    for (s=[-1,1]) translate([s*(collar_o/2), 0, collar_h*0.45]) rotate([0,90,0]) cylinder(d=mq_can_d+1, h=10, center=true);
    // DHT11 window on a third wall
    translate([-dht[0]/2-0.5, collar_o/2-collar_wall-1, collar_h*0.45-dht[1]/2-0.5]) cube([dht[0]+1, collar_wall+2, dht[1]+1]);
    // fan screw holes
    for (sx=[-1,1], sy=[-1,1]) translate([sx*fan_hole/2, sy*fan_hole/2, collar_h-5]) cylinder(d=3.2, h=8);
  }
}

// Side-wall bracket for designs NOW (under the lid), A and D: two MQ cans + DHT, held off the wall.
// Print flat (plate down).
module sensor_wall_mount() {
  difference() {
    union() {
      translate([-50, -22, 0]) cube([100, 44, 3]);                       // sensor plate
      for (sx=[-1,1]) translate([sx*44-4, -22, 0]) cube([8, 44, standoff]); // standoff legs
      for (sx=[-1,1]) translate([sx*44-4-6, -22, standoff-3]) cube([20, 44, 3]); // wall feet
    }
    for (sx=[-1,1]) translate([sx*22, 0, -1]) cylinder(d=mq_can_d+1, h=5);   // MQ cans
    translate([-dht[0]/2-0.5, -dht[1]/2-0.5, -1]) cube([dht[0]+1, dht[1]+1, 5]); // DHT window
    for (sx=[-1,1], sy=[-1,1]) translate([sx*(22+mq_pcb[0]/2-3), sy*(mq_pcb[1]/2-3), -1]) cylinder(d=2.5, h=5); // board screws M2
    for (sx=[-1,1], sy=[-1,1]) translate([sx*48, sy*14, standoff-4]) cylinder(d=3.4, h=6);  // wall screws M3
  }
}

// Foot that lifts a kitchen cooling rack to rack_h. Print 4.
module rack_foot() {
  difference() {
    cylinder(d1=22, d2=12, h=rack_h + rod/2);
    translate([-15, -rod/2-0.2, rack_h - 0.5]) cube([30, rod+0.4, 10]); // slot for the rack rod
  }
}

// Wall port that holds an activated-carbon cartridge (C, D, F) or a cap. Glue through a 41 mm hole.
module filter_port() {
  difference() {
    union() {
      cylinder(d=filter_d+4, h=filter_len);
      cylinder(d=filter_d+20, h=3);                       // flange inside the box
    }
    translate([0,0,3]) cylinder(d=filter_d, h=filter_len);   // carbon goes here
    // grid at the inside end keeps carbon in
    for (a=[0,90]) rotate([0,0,a]) for (o=[-15:6:15]) translate([o-1.5, -filter_d/2, -1]) cube([3, filter_d, 5]);
  }
}

// Pass-through for the B loop tube (or any 16 mm hose). Glue through a 20 mm hole.
module tube_port() {
  difference() {
    union() { cylinder(d=port_d+5, h=25); cylinder(d=port_d+18, h=3); }
    translate([0,0,-1]) cylinder(d=port_d+0.4, h=30);
  }
}

// Holds a 40 mm fan over a 38 mm hole in a wall or the lid.
module fan_mount_40() {
  difference() {
    translate([-fan/2-6, -fan/2-6, 0]) cube([fan+12, fan+12, 4]);
    translate([0,0,-1]) cylinder(d=fan-2, h=6);
    for (sx=[-1,1], sy=[-1,1]) translate([sx*fan_hole/2, sy*fan_hole/2, -1]) cylinder(d=3.2, h=6);
    for (sx=[-1,1], sy=[-1,1]) translate([sx*(fan/2+2), sy*(fan/2+2), -1]) cylinder(d=2.5, h=6);
  }
}

// Flat cutting template for the V-hood: 4 panels from a 0.5–1 mm PET/PP sheet, taped or glued
// to the collar flange. Export as SVG/DXF (part = hood_net, then F6, File > Export > SVG).
module hood_net() {
  a1 = mouth[0]; b1 = mouth[1]; a2 = throat + 4; b2 = throat + 4;
  sa = sqrt(hood_h*hood_h + pow((b1-b2)/2,2));  // slant of the long panels
  sb = sqrt(hood_h*hood_h + pow((a1-a2)/2,2));  // slant of the short panels
  module trap(bot, top, h) polygon([[-bot/2,0],[bot/2,0],[top/2,h],[-top/2,h]]);
  // 10 mm tab on the narrow (top) edge glues under the collar flange; tape the slanted edges together.
  module tab(len, h) translate([-len/2, h]) square([len, 10]);
  translate([0, 0]) { trap(a1, a2, sa); tab(a2, sa); }
  translate([0, sa + 40]) { trap(a1, a2, sa); tab(a2, sa); }
  translate([a1/2 + b1/2 + 30, 0]) { trap(b1, b2, sb); tab(b2, sb); }
  translate([a1/2 + b1/2 + 30, sb + 40]) { trap(b1, b2, sb); tab(b2, sb); }
}

// =====================================================================
// Hood (E, F)
// =====================================================================
module vhood() {
  color(C_HOOD) difference() {
    hull() {
      translate([-mouth[0]/2, -mouth[1]/2, hood_z0]) cube([mouth[0], mouth[1], 0.8]);
      translate([-throat/2-2, -throat/2-2, hood_z1-0.8]) cube([throat+4, throat+4, 0.8]);
    }
    hull() {
      translate([-mouth[0]/2+1, -mouth[1]/2+1, hood_z0-0.1]) cube([mouth[0]-2, mouth[1]-2, 0.8]);
      translate([-throat/2, -throat/2, hood_z1-0.6]) cube([throat, throat, 0.8]);
    }
  }
}

// Collar on top of the hood; MQ boards sit outside two walls with the cans pointing into the throat.
module collar_with_sensors(z, h, with_fan) {
  translate([0,0,z]) {
    color(C_PRINT) throat_collar(h, with_fan);
    for (s=[-1,1]) translate([s*(collar_o/2+0.5), 0, h*0.45]) rotate([0, -s*90, 0]) mq_module();
    translate([0, collar_o/2, h*0.45]) rotate([-90,0,0]) dht11();
    if (with_fan) translate([0,0,h]) fan40();
  }
}

// =====================================================================
// Holes per design
// =====================================================================
module wall_holes() {
  if (design=="B") {
    translate([xR-1, 0, 25]) rotate([0,90,0]) cylinder(d=port_d+4, h=wall+2);
    translate([xR-1, 0, box_H-30]) rotate([0,90,0]) cylinder(d=port_d+4, h=wall+2);
    translate([xL-wall-1, 0, box_H-30]) rotate([0,90,0]) cylinder(d=filter_d+1, h=wall+2);
  }
  if (design=="C" || design=="D" || design=="F")
    translate([xL-wall-1, 0, 45]) rotate([0,90,0]) cylinder(d=filter_d+1, h=wall+2);
  if (design=="C" || design=="D")
    translate([xR-1, 0, box_H-45]) rotate([0,90,0]) cylinder(d=fan-2, h=wall+2);
}
module lid_holes() {
  if (design=="F") {
    translate([-collar_o/2, -collar_o/2, box_H-1]) cube([collar_o, collar_o, wall+2]);
    translate([xR-60, yB-60, box_H-1]) cylinder(d=port_d+4, h=wall+2);   // return
  }
}

// =====================================================================
// Assemblies
// =====================================================================
module assembly() {
  box_shell();
  lid();

  if (has_rack) { drip_tray(hole = design=="E"); rack(); chicken(chick_z0); }
  else chicken(0);

  if (design=="NOW") {
    // sensors on a bracket under the lid, facing down
    translate([60, 0, box_H - standoff]) {
      color(C_PRINT) sensor_wall_mount();
      for (sx=[-1,1]) translate([sx*22,0,0]) rotate([180,0,0]) mq_module();
    }
  }

  if (design=="A" || design=="D") {
    // sensors on the left wall, mid-height, in the return flow of the stir loop
    translate([xL + standoff, 0, box_H*0.55]) rotate([0,-90,0]) {
      color(C_PRINT) sensor_wall_mount();
      for (sx=[-1,1]) translate([sx*22,0,0]) rotate([180,0,0]) mq_module();
    }
    // stir fan, top-left corner, blowing along the lid toward +x
    translate([xL+30, 0, box_H-35]) rotate([0,90,0]) fan40();
    translate([xL+28, 0, box_H-35]) rotate([0,90,0]) color(C_PRINT) fan_mount_40();
  }

  if (design=="B") {
    // ports
    translate([xR, 0, 25]) rotate([0,90,0]) color(C_PRINT) tube_port();
    translate([xR, 0, box_H-30]) rotate([0,90,0]) color(C_PRINT) tube_port();
    // loop outside the right wall
    tube([xR+25, 0, 25], [xR+60, 0, 25]);
    tube([xR+60, 0, 25], [xR+60, 0, box_H-30]);
    tube([xR+60, 0, box_H-30], [xR+25, 0, box_H-30]);
    // sensor chamber on the riser, then the fan (after the sensors)
    translate([xR+60, 0, 110]) {
      color(C_PRINT, 0.6) translate([-25,-25,-30]) cube([50,50,60]);
      translate([-26, 0, 0]) rotate([0,90,0]) mq_module();
      translate([26, 0, 0]) rotate([0,-90,0]) mq_module();
    }
    translate([xR+60, 0, 200]) fan40();
    // purge valve (tee) + carbon filter on the left wall
    color(C_FILT) translate([xR+60, 0, box_H-30]) sphere(d=26);
    tube([xR+60, 0, box_H-30], [xR+110, 0, box_H-30]);
    translate([xL, 0, box_H-30]) rotate([0,-90,0]) { color(C_PRINT) filter_port(); }
  }

  if (design=="C" || design=="D") {
    // filter inlet low on the left wall
    translate([xL, 0, 45]) rotate([0,-90,0]) color(C_PRINT) filter_port();
    // outlet high on the right wall
    translate([xR+wall, 0, box_H-45]) rotate([0,90,0]) { color(C_PRINT) fan_mount_40(); }
  }
  if (design=="C") {
    // sensor chamber outside the outlet, fan after it
    translate([xR+wall+4, 0, box_H-45]) {
      color(C_PRINT, 0.6) translate([0,-28,-28]) cube([70,56,56]);
      translate([35, 0, 29]) rotate([180,0,0]) mq_module();
      translate([35, 0, -29]) mq_module();
      translate([70, 0, 0]) rotate([0,90,0]) fan40();
    }
  }
  if (design=="D") {
    translate([xR+wall+4, 0, box_H-45]) rotate([0,90,0]) fan40();   // exhaust fan outside the outlet
  }

  if (design=="E") {
    // fan in the floor under the rack, blowing up, with a drip shield above it
    translate([0,0,2]) fan40();
    color(C_PRINT) translate([0,0,rack_h-6]) cylinder(d=fan+30, h=1.5);   // drip shield
    for (a=[0:120:359]) rotate([0,0,a]) color(C_PRINT) translate([fan/2+8,0,12]) cylinder(d=4, h=rack_h-18);
    vhood();
    collar_with_sensors(hood_z1, collar_h_E, false);
  }

  if (design=="F") {
    vhood();
    collar_with_sensors(hood_z1, collar_h, true);   // collar runs up through the lid; fan sits on top
    // flap box on the fan: loop back (return tube) or out (flush)
    translate([0,0,hood_z1+collar_h+fan_t]) color(C_PRINT, 0.7) translate([-25,-25,0]) cube([50,50,25]);
    tube([25, 0, hood_z1+collar_h+fan_t+12], [xR-60, yB-60, hood_z1+collar_h+fan_t+12]);
    tube([xR-60, yB-60, hood_z1+collar_h+fan_t+12], [xR-60, yB-60, box_H-5]);
    tube([-25, 0, hood_z1+collar_h+fan_t+12], [-70, 0, hood_z1+collar_h+fan_t+12]);  // to outside (flush)
    // filter inlet low on the left wall, capped while sampling
    translate([xL, 0, 45]) rotate([0,-90,0]) color(C_PRINT) filter_port();
  }
}

// =====================================================================
// Entry
// =====================================================================
if (part=="assembly")           assembly();
else if (part=="throat_collar") throat_collar(collar_h, true);
else if (part=="throat_collar_E") throat_collar(collar_h_E, false);
else if (part=="sensor_wall_mount") sensor_wall_mount();
else if (part=="rack_foot")     rack_foot();
else if (part=="filter_port")   filter_port();
else if (part=="tube_port")     tube_port();
else if (part=="fan_mount_40")  fan_mount_40();
else if (part=="hood_net")      hood_net();
