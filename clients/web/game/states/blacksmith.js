import { WhereType } from '../../model/enums.js';
import { randBetween } from '../../mechanics/rand.js';
import { saveCharacter } from '../../store/local.js';
import { TownState } from './town.js';

// BlacksmithState — port of internal/game/states/blacksmith.go. Three random
// outcomes affecting the primary weapon's strike.
export class BlacksmithState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    c.location = WhereType.TheBlacksmiths;
    const weapName = c.weapons[0]?.name || 'weapon';

    switch (randBetween(0, 2)) {
      case 0:
        io.println("You enter the blacksmith's shop and Smith immediately", 3);
        io.println('begins laughing. Smith picks you up by the britches,', 3);
        io.println('biceps bulging, and throws you back out on the road.', 3);
        io.cr();
        break;
      case 1:
        io.println("You enter the blacksmith's shop and Smith gives you", 3);
        io.println('a pat on the back. After you get back off the ground', 3);
        io.println(`and dust yourself off, Smith takes your ${weapName}`, 3);
        io.println('and begins to work on it. When he returns the weapon,', 3);
        io.println("you can't help but notice its fine craftsmanship. You", 3);
        io.println('immediately head for the door, thanking Smith as you leave.', 3);
        io.cr();
        c.weapons[0].strike += randBetween(1, 3);
        break;
      case 2:
        io.println("You enter the blacksmith's shop and Smith wallows up", 3);
        io.println('to shake your hand. After you get back off the ground', 3);
        io.println(`and dust yourself off, Smith takes your ${weapName}`, 3);
        io.println('and begins to work on it. When he returns the weapon,', 3);
        io.println("you can't help but notice it has several new dents.", 3);
        io.println('You slowly head for the door. Smith smiles, and asks', 3);
        io.println('you to come again soon.', 3);
        io.cr();
        c.weapons[0].strike -= randBetween(1, 2);
        if (c.weapons[0].strike < 0) c.weapons[0].strike = 0;
        break;
    }

    saveCharacter(c);
    await io.pausePrompt('-=Press A Key=-');
    session.setNext(new TownState());
  }
}
