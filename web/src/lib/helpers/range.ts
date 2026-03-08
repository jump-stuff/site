export const range = (min: number, max: number) =>
  [...Array(Math.max(max - min, 0)).keys()].map((i) => i + min);
