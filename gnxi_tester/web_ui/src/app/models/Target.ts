export interface Target {
  address: string;
  ca: string;
  cakey: string;
}

export type Targets = {[name: string]: Target};
