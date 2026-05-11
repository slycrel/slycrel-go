import { loadInn, saveInn, writeNews, findCharacterByName } from '../store/local.js';

// Daily upkeep — runs once per calendar day across all players.
// inn.lastUpkeep is stamped after each pass so concurrent logins don't
// double-decrement. (Browser-only is single-user; this still matters
// if the player logs in twice on the same day.)
export function runDailyUpkeep(today) {
  innUpkeep(today);
}

function innUpkeep(today) {
  const inn = loadInn();
  if ((inn.lastUpkeep ?? 0) >= today) return;
  for (const room of inn.rooms) {
    if (room.who && room.daysLeft > 0) {
      room.daysLeft -= 1;
      if (room.daysLeft <= 0) {
        const occupant = findCharacterByName(room.who);
        if (occupant) {
          writeNews(occupant.bbsName,
            `Your room at the Inn has expired. The innkeeper kept your deposit.`);
        }
        room.who = '';
        room.lock = 0;
      }
    }
  }
  inn.lastUpkeep = today;
  saveInn(inn);
}
