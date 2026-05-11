// Random integer in [min, max] inclusive. Matches mechanics/RandBetween in
// internal/mechanics — kept as its own module so combat/leveling/etc. share it.
export function randBetween(min, max) {
  if (min >= max) return min;
  return min + Math.floor(Math.random() * (max - min + 1));
}

export function pick(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

// Today as a YYYYMMDD integer — matches todayDate() in the Go side and is
// what we store in Character.lastOn so the new-day check on login is a
// plain integer comparison.
export function todayDate() {
  const d = new Date();
  return d.getFullYear() * 10000 + (d.getMonth() + 1) * 100 + d.getDate();
}
