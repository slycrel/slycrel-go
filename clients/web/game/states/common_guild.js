import { WhereType, CharClass } from '../../model/enums.js';
import { level } from '../../model/character.js';
import { nextLevelUp, giveNewLevel } from '../../mechanics/leveling.js';
import { saveCharacter } from '../../store/remote.js';
import { TownState } from './town.js';

// The Go server reads this from TownConfig at runtime; we don't have a
// server-side config, so hardcode the same fallback default (4).
const LEVEL_UP_DIFFICULTY = 4;
const MAX_LEVEL = 60;

export class CommonGuildState {
  async enter(session) {
    session.character.location = WhereType.TheCommonGuild;
    session.io.clear();
    await session.io.showAnsiFile('common_guild_menu');
    session.setNext(new GuildPromptState());
  }
}

export class GuildPromptState {
  async enter(session) {
    const { io } = session;
    io.println('[L]evel Up, [Q]uit, [?]Help', 1);
    io.cr();
    const choice = await io.lettersPrompt('Your choice?', 'LQ?');
    io.cr();
    switch (choice) {
      case 'L': session.setNext(new TryNewLevelState()); return;
      case 'Q':
        session.character.location = WhereType.TheTown;
        session.setNext(new TownState());
        return;
      case '?': session.setNext(new CommonGuildState()); return;
    }
    session.setNext(new GuildPromptState());
  }
}

export class TryNewLevelState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    const currentLevel = level(c);
    const required = nextLevelUp(currentLevel + 1, LEVEL_UP_DIFFICULTY);

    io.cr();
    io.println(`  Your Current Level: ${currentLevel}`, 4);
    io.println(`  Your Total Experience: ${c.totalExperience}`, 4);
    io.println(`  Required Experience: ${required}`, 4);
    io.cr();

    if (currentLevel >= MAX_LEVEL) {
      io.println("The GuildMaster uncomfortably says, \"I'm sorry, you've maxed out.", 2);
    } else if (c.totalExperience >= required) {
      io.println('"Finally someone who is ready to go up," grins the GuildMaster.', 5);
      giveNewLevel(c);
      io.println(`You have gained one level! You are now level ${level(c)}!`, 3);
      io.cr();
      io.println(`  HP: ${c.maxHP}  Str: ${c.strength}  Dex: ${c.dexterity}  Spd: ${c.speed}`, 1);
      if (c.charClass === CharClass.Mage) {
        io.println(`  Psyche: ${c.maxPsyche}`, 1);
      }
      await saveCharacter(c);
    } else {
      io.println('The GuildMaster looks at you and Laughs.', 6);
      const needed = required - c.totalExperience;
      io.println(`  (You need ${needed} more experience)`, 1);
    }
    io.cr();
    session.setNext(new GuildPromptState());
  }
}
