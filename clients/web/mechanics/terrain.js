import { GRID_ROWS, GRID_COLS, TerrainCell } from '../model/enums.js';

// One glyph + fg/bg per terrain cell. Chosen to mirror the Go terminal
// renderer's intent (terrain.go) without trying to reproduce a 16-color
// terminal palette exactly — the browser has full RGB so we get nicer
// hues. Water and bridges have a background tint; everything else is
// black-on-glyph for readability.
export const TERRAIN_DISPLAY = [
  { ch: ' ', fg: '#9c9', bg: '#000' }, // 0 grass/empty
  { ch: '^', fg: '#070', bg: '#000' }, // 1 plain (green)
  { ch: '.', fg: '#aa0', bg: '#000' }, // 2 plain (road/brown)
  { ch: '~', fg: '#7af', bg: '#226' }, // 3 water (instant death)
  { ch: '=', fg: '#fd6', bg: '#964' }, // 4 bridge
  { ch: 'O', fg: '#999', bg: '#000' }, // 5 boulder
  { ch: '#', fg: '#5d5', bg: '#020' }, // 6 deep forest
  { ch: '+', fg: '#070', bg: '#000' }, // 7 light forest
  { ch: '%', fg: '#dd0', bg: '#272' }, // 8 swamp
  { ch: '&', fg: '#070', bg: '#272' }, // 9 deep swamp
];

// Movement cost for entering a cell — 998 means impassable. Water is
// also a special-case (instant death) handled by moveGridPlayer.
export function numMovePoints(cell) {
  switch (cell) {
    case TerrainCell.Empty:      return 1;
    case TerrainCell.PlainGr:    return 2;
    case TerrainCell.PlainBr:    return 1;
    case TerrainCell.Water:      return 998;
    case TerrainCell.Bridge:     return 1;
    case TerrainCell.Boulder:    return 998;
    case TerrainCell.Forest:     return 3;
    case TerrainCell.DeepForest: return 4;
    case TerrainCell.Swamp:      return 3;
    case TerrainCell.DeepSwamp:  return 5;
    default: return 998;
  }
}

export function numMovePointsAt(r, c, terrain) {
  if (r < 0 || r >= GRID_ROWS || c < 0 || c >= GRID_COLS) return 998;
  return numMovePoints(terrain.cells[r][c]);
}

export function canMoveAnywhere(r, c, movement, terrain) {
  if (r > 0 && movement >= numMovePointsAt(r - 1, c, terrain)) return true;
  if (r < GRID_ROWS - 1 && movement >= numMovePointsAt(r + 1, c, terrain)) return true;
  if (c > 0 && movement >= numMovePointsAt(r, c - 1, terrain)) return true;
  if (c < GRID_COLS - 1 && movement >= numMovePointsAt(r, c + 1, terrain)) return true;
  return false;
}

// Dijkstra from (sr, sc) to (dr, dc). Returns a U/D/L/R string of moves or
// null if unreachable. 12×48 is small enough that a sorted-scan priority
// queue is fast; not worth a heap.
export function findRoute(sr, sc, dr, dc, terrain) {
  const INF = 32700;
  const w = Array.from({ length: GRID_ROWS }, () => new Array(GRID_COLS).fill(INF));
  const dir = Array.from({ length: GRID_ROWS }, () => new Array(GRID_COLS).fill(0));
  const seen = Array.from({ length: GRID_ROWS }, () => new Array(GRID_COLS).fill(false));
  w[sr][sc] = numMovePointsAt(sr, sc, terrain);
  const queue = [[sr, sc]];
  const NEIGHBORS = [[-1, 0, 'U'], [1, 0, 'D'], [0, -1, 'L'], [0, 1, 'R']];
  for (let iter = 0; iter < 1050 && queue.length; iter++) {
    let best = 0;
    for (let i = 1; i < queue.length; i++) {
      const [r, c] = queue[i], [br, bc] = queue[best];
      if (w[r][c] < w[br][bc]) best = i;
    }
    const [cr, cc] = queue.splice(best, 1)[0];
    seen[cr][cc] = true;
    if (cr === dr && cc === dc) break;
    for (const [dr2, dc2, d] of NEIGHBORS) {
      const nr = cr + dr2, nc = cc + dc2;
      if (nr < 0 || nr >= GRID_ROWS || nc < 0 || nc >= GRID_COLS) continue;
      if (seen[nr][nc]) continue;
      const cost = w[cr][cc] + numMovePointsAt(nr, nc, terrain);
      if (cost < w[nr][nc]) {
        w[nr][nc] = cost;
        dir[nr][nc] = d;
        queue.push([nr, nc]);
      }
    }
  }
  if (!seen[dr][dc] && w[dr][dc] === INF) return null;
  const route = [];
  let r = dr, c = dc;
  while (r !== sr || c !== sc) {
    const d = dir[r][c];
    route.unshift(d);
    if (d === 'U') r += 1;
    else if (d === 'D') r -= 1;
    else if (d === 'L') c += 1;
    else if (d === 'R') c -= 1;
    if (route.length > 100) break;
  }
  return route.join('');
}
