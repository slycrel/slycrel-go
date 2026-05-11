import { EnterSlycrelState } from './enter_slycrel.js';
import { listCharacters } from '../../store/local.js';

// BeginState — title screen + main entry menu.
// Mirrors internal/game/states/begin.go (BeginState).
export class BeginState {
  async enter(session) {
    const { io } = session;
    io.clear();
    await io.showAnsiFile('opening');
    io.cr();
    io.println('=--=--  Welcome to the Realm of Slycrel  ---=--=', 3);
    io.cr();
    io.println('[E]nter the Realm', 4);
    io.println('[V]iew Guild Lists', 4);
    io.println('[C]reate/View Character', 4);
    io.println('[L]eave', 4);
    io.cr();

    const choice = await io.lettersPrompt('Your Selection >', 'EVCL');
    io.cr();

    switch (choice) {
      case 'E':
      case 'C':
        session.setNext(new EnterSlycrelState());
        return;
      case 'V':
        await renderGuildLists(io);
        await io.pausePrompt();
        session.setNext(new BeginState());
        return;
      case 'L':
        io.println('Exiting Slycrel...', 3);
        session.quit();
        return;
    }
  }
}

// Shared guild-rankings table. The tavern's ViewGuildsState prints
// effectively the same thing — kept inline rather than dragged into a
// new module because it's the same ~15 lines and the duplication is
// easier to read than the indirection would be.
async function renderGuildLists(io) {
  const chars = listCharacters();
  if (!chars.length) {
    io.println('No adventurers have registered yet.', 1);
    return;
  }
  io.cr();
  io.println('=--=-- Guild Rankings ---=--=', 3);
  io.cr();
  io.println(' Name                 Class      F.Lvl T.Lvl M.Lvl       XP', 4);
  io.println(' ------------------------------------------------------------', 1);
  for (const ch of chars) {
    const name = (ch.name ?? '').padEnd(20).slice(0, 20);
    const cls = (ch.charClass ?? '').padEnd(10).slice(0, 10);
    const f = String(ch.fighterLvl ?? 0).padStart(5);
    const t = String(ch.thiefLvl ?? 0).padStart(5);
    const m = String(ch.mageLvl ?? 0).padStart(5);
    const xp = String(ch.totalExperience ?? 0).padStart(8);
    io.println(` ${name} ${cls} ${f} ${t} ${m} ${xp}`, 1);
  }
  io.cr();
}
