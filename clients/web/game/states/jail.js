import { WhereType } from '../../model/enums.js';
import { randBetween } from '../../mechanics/rand.js';
import { saveCharacter } from '../../store/remote.js';
import { TownState } from './town.js';
import { DeadState } from './dead.js';

// JailState — port of internal/game/states/jail.go. Three random outcomes,
// one of which kills you for the day.
export class JailState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    c.location = WhereType.TheJail;
    const weapName = c.weapons[0]?.name || 'weapon';

    switch (randBetween(0, 2)) {
      case 0:
        io.println('As you enter the jail, in a cell you see your fellow', 3);
        io.println('hero Night Hawk. Quickly realizing this is no place', 3);
        io.println('for a hero such as yourself, you head right back out', 3);
        io.println('the door.', 3);
        io.cr();
        await io.pausePrompt('-=Press A Key=-');
        session.setNext(new TownState());
        return;

      case 1:
        io.println('As you enter the jail, the jailor walks up and shakes', 3);
        io.println("your hand, saying he's always glad to see a hero visit.", 3);
        io.println('The next thing you know you find yourself locked in a', 3);
        io.println('cell with the jailor laughing at you from outside.', 3);
        io.println("Looks like you'll have to spend the night.", 3);
        io.cr();
        c.alive = false;
        c.location = WhereType.TheTown;
        await saveCharacter(c);
        await io.pausePrompt('-=Press A Key=-');
        session.setNext(new DeadState());
        return;

      case 2:
        io.println('As you enter the jail, the jailor walks up and shakes', 3);
        io.println('your hand, saying he is glad to see such a fine hero', 3);
        io.println('visit. The jailor takes you over to his weapon rack.', 3);
        io.println(`You can't help but notice the jailors ${weapName}`, 3);
        io.println('is in much better shape than your own. The jailor cant', 3);
        io.println('help but notice your envy, and offers to loan you the', 3);
        io.println('weapon. Before he can change his mind, you grab the', 3);
        io.println('weapon and head out the door.', 3);
        io.cr();
        c.weapons[0].strike += randBetween(1, 3);
        await saveCharacter(c);
        await io.pausePrompt('-=Press A Key=-');
        session.setNext(new TownState());
        return;
    }
  }
}
