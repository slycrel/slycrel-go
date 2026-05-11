// Random integer in [min, max] inclusive. Matches mechanics/RandBetween in
// internal/mechanics — kept as its own module so combat/leveling/etc. share it.
export function randBetween(min, max) {
  if (min >= max) return min;
  return min + Math.floor(Math.random() * (max - min + 1));
}

export function pick(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}
