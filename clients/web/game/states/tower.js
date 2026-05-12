import { WhereType } from '../../model/enums.js';
import { randBetween } from '../../mechanics/rand.js';
import { saveCharacter } from '../../store/remote.js';
import { TownState } from './town.js';

// TowerState — port of internal/game/states/tower.go. Four random outcomes
// that adjust Faith up or down.
export class TowerState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    c.location = WhereType.TheTower;

    switch (randBetween(0, 3)) {
      case 0:
        io.println('You enter the dark dingy tower, and immediately encounter', 3);
        io.println('a holy man saying a prayer in some strange language.', 3);
        io.println('After he finishes, the holy man looks up and smiles,', 3);
        io.println('happy to see a hero such as yourself. Again, he begins', 3);
        io.println('to chant out a prayer....', 3);
        io.println("Thankfully, it's a blessing. When he's finished he", 3);
        io.println("sends you on to do your hero's work.", 3);
        io.cr();
        c.faith += randBetween(1, 2);
        break;
      case 1:
        io.println('You enter the tower, hopeful that you will gain some', 3);
        io.println('much needed wisdom....', 3);
        io.cr();
        io.println('Eventually, you encounter a High Priestess. The', 3);
        io.println('priestess lets out a shriek of laugh, and casts a spell', 3);
        io.println('on you...', 3);
        io.cr();
        io.println('When you awaken, you feel dreadfully weak. When you', 3);
        io.println('eventually find the strength, you leave the tower as', 3);
        io.println('quickly as possible.', 3);
        io.cr();
        c.faith -= randBetween(1, 2);
        break;
      case 2:
        io.println('As you enter the tower, you notice an evil feeling', 3);
        io.println("has covered the area. Being a hero, you don't seem to mind.", 3);
        io.println('After a considerable search, you encounter a wizard. The', 3);
        io.println('Wizard immediately begins to cast a spell.....', 3);
        io.println('Suddenly you find yourself back in town,', 3);
        io.println('with an incredible amount of new energy.', 3);
        io.cr();
        c.faith += randBetween(1, 3);
        break;
      case 3:
        io.println('You enter the tower, and a fearful feeling', 3);
        io.println('immediately consumes your every thought. As you', 3);
        io.println('peer through the darkness, you see a faint image', 3);
        io.println('across the room. As you creep closer, you finally', 3);
        io.println("realize you've encountered a High Priest. Before", 3);
        io.println('you can even turn around, he casts a spell on you.', 3);
        io.println('An incredible dizziness immediately overtakes you....', 3);
        io.cr();
        io.println('You awaken in the center of town, so weak you can', 3);
        io.println('barely stand.', 3);
        io.cr();
        c.faith -= randBetween(1, 2);
        break;
    }

    await saveCharacter(c);
    await io.pausePrompt('-=Press A Key=-');
    session.setNext(new TownState());
  }
}
