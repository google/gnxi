export interface Prompts {
  name: string;
  prompts: {[name: string]: string};
  files: {[name: string]: string};
}

export interface PromptsList {
  prompts: string[];
  files: string[];
}

export type PromptsSet = {[name: string]: Prompts}
