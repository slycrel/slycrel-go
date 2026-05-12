import { loadInn, saveInn, writeNews, findCharacterByName } from '../store/remote.js';

// Daily upkeep — runs once per calendar day across all players.
// inn.lastUpkeep is stamped after each pass so two players logging in on
// the same day don't double-decrement the rooms.
export async function runDailyUpkeep(today) {
  await innUpkeep(today);
}

async function innUpkeep(today) {
  const inn = await loadInn();
  if ((inn.lastUpkeep ?? 0) >= today) return;
  for (const room of inn.rooms) {
    if (room.who && room.daysLeft > 0) {
      room.daysLeft -= 1;
      if (room.daysLeft <= 0) {
        const occupant = await findCharacterByName(room.who);
        if (occupant) {
          await writeNews(occupant.bbsName,
            `Your room at the Inn has expired. The innkeeper kept your deposit.`);
        }
        room.who = '';
        room.lock = 0;
      }
    }
  }
  inn.lastUpkeep = today;
  await saveInn(inn);
}
