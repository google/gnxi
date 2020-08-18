export interface Device {
  target_name: string;
  target_address: string;
}

export type Devices = {[name: string]: Device}

export function GetDevices(): Devices {
  return {"test": {target_name: "test", target_address:"test:1111"}};
}
