export interface Target {
  target_name: string;
  target_address: string;
}

export type Targets = {[name: string]: Target};
