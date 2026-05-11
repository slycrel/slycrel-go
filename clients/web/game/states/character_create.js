import { newBlankCharacter } from '../../model/character.js';
import { rollChar, randomClass } from '../../mechanics/leveling.js';
import { saveCharacter } from '../../store/local.js';
import { ViewCharacterState } from './view_character.js';
import { TownState } from './town.js';

// CharacterCreateState — simplified port of internal/game/states/character_create.go.
// Prompts for a name, rolls a random class, shows the sheet, then enters town.
// The full Go flow (N/F/C/G/R/S menu options to customize) lands in a later
// commit — for now you get whatever the dice give you.
export class CharacterCreateState {
  async enter(session) {
    const { io } = session;
    io.cr();
    const name = (await io.textPrompt('What is your new Name:', 20)).trim();
    if (!name) {
      io.println('A nameless wanderer cannot enter the realm.', 6);
      await io.pausePrompt();
      session.popReturn();
      return;
    }

    const char = newBlankCharacter(session.username, name);
    rollChar(char, randomClass());
    session.character = char;

    session.setReturn(new EnterTownFromCreateState());
    session.setNext(new ViewCharacterState());
  }
}

// EnterTownFromCreateState — saves the freshly-rolled character and drops
// the player into the town hub.
export class EnterTownFromCreateState {
  async enter(session) {
    const { io, character } = session;
    saveCharacter(character);
    io.cr();
    io.println(`${character.name} steps into the realm of Slycrel...`, 3);
    await io.pausePrompt();
    session.setNext(new TownState());
  }
}
