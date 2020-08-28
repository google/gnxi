export interface Test {
    args: object;
    doesntwant: string;
    mustfail:   boolean;
    name:       string;
    prompt:     string[];
    wait:       number;
    wants:      string;
}

export type Tests = {[name: string]: Test[]};
