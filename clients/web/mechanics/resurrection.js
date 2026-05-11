import { MAX_EXPLORATION, MAX_SPARS } from '../model/character.js';
import { randBetween } from './rand.js';

// Mirrors internal/mechanics/resurrection.go ResurrectUser:
// -2% TotalExperience, -2% SpendingExperience, -1 Fame, +1 Faith,
// -20% CoinsHand, then restore HP and revive.
export function resurrectUser(c) {
  c.totalExperience -= Math.round(c.totalExperience * 0.02);
  c.spendingExperience -= Math.round(c.spendingExperience * 0.02);
  c.fame -= 1;
  c.faith += 1;
  c.coinsHand -= Math.round(c.coinsHand * 0.20);
  if (c.totalExperience < 0) c.totalExperience = 0;
  if (c.spendingExperience < 0) c.spendingExperience = 0;
  if (c.coinsHand < 0) c.coinsHand = 0;
  c.alive = true;
  c.hitPoints = c.maxHP;
}

// Mirrors NewDayForUser — resets daily limits + heals + revives if dead.
export function newDayForUser(c) {
  c.psyche = c.maxPsyche;
  c.totalExperience += 15;
  c.spendingExperience += 15;
  const spread = Math.floor(c.speed / 5);
  c.movement = randBetween(c.speed - spread, c.speed + spread);
  c.exploration = MAX_EXPLORATION;
  c.spars = MAX_SPARS;
  if (!c.alive) {
    resurrectUser(c);
  } else {
    c.hitPoints = c.maxHP;
  }
}
