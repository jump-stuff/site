const formats = ['monthly', 'motw', 'archive', 'test'];

export function match(value: string) {
  return formats.includes(value);
}
