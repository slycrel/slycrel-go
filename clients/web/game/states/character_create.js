import { newBlankCharacter } from '../../model/character.js';
import { CharClass, Color, COLOR_NAMES } from '../../model/enums.js';
import { rollChar, randomClass } from '../../mechanics/leveling.js';
import { saveCharacter } from '../../store/local.js';
import { ViewCharacterState } from './view_character.js';
import { TownState } from './town.js';

// CharacterCreateState — port of internal/game/states/character_create.go.
// Rolls a random initial class, displays the sheet, then drops into the
// edit menu (V/N/F/C/G/R/S/E) before the player commits to entering town.
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
    session.setReturn(new CreateCharMenuState());
    session.setNext(new ViewCharacterState());
  }
}

// CreateCharMenuState — the V/N/F/C/G/R/S/E loop. Most options branch to
// a small sub-state and return back here via the return-stack.
export class CreateCharMenuState {
  async enter(session) {
    const { io } = session;
    io.cr();
    io.println(' V)iew Character.', 4);
    io.println(' N)ew Name.', 4);
    io.println(' F)avorite Color', 4);
    io.println(' C)hange Occupation.', 4);
    io.println(' G)ender Change.', 4);
    io.println(' R)eroll Attributes. (Caution! Resets to first level)', 4);
    io.println(' S)how Scale.', 4);
    io.println(' E)nter Town!', 4);
    io.cr();
    const choice = await io.lettersPrompt('[V,N,F,C,R,S,E,G]', 'VNFCRSEG');
    io.cr();
    switch (choice) {
      case 'V':
        session.setReturn(new CreateCharMenuState());
        session.setNext(new ViewCharacterState());
        return;
      case 'N': session.setNext(new GetCharNameState()); return;
      case 'F': session.setNext(new FavoriteColorState()); return;
      case 'C': session.setNext(new OccupationMenuState()); return;
      case 'G':
        toggleGender(session);
        session.setReturn(new CreateCharMenuState());
        session.setNext(new ViewCharacterState());
        return;
      case 'R':
        rollChar(session.character, session.character.charClass);
        session.setReturn(new CreateCharMenuState());
        session.setNext(new ViewCharacterState());
        return;
      case 'S':
        showScale(io);
        await io.pausePrompt();
        session.setNext(new CreateCharMenuState());
        return;
      case 'E': session.setNext(new EnterTownFromCreateState()); return;
    }
    session.setNext(new CreateCharMenuState());
  }
}

export class GetCharNameState {
  async enter(session) {
    const { io } = session;
    const name = (await io.textPrompt('What is your new Name:', 20)).trim();
    if (name) session.character.name = name;
    session.setNext(new CreateCharMenuState());
  }
}

export class OccupationMenuState {
  async enter(session) {
    const { io } = session;
    io.println('[F]ighter', 1);
    io.println('[T]hief', 1);
    io.println('[M]age', 1);
    const choice = await io.lettersPrompt('New class -', 'FTM');
    io.cr();
    const c = session.character;
    if (choice === 'F') rollChar(c, CharClass.Fighter);
    else if (choice === 'T') rollChar(c, CharClass.Thief);
    else if (choice === 'M') rollChar(c, CharClass.Mage);
    session.setReturn(new CreateCharMenuState());
    session.setNext(new ViewCharacterState());
  }
}

// FavoriteColorState — picks favColor 0..7 from a printed 1..8 list.
// Each line is colored close to the named color so the choice is visual.
export class FavoriteColorState {
  async enter(session) {
    const { io } = session;
    io.cr();
    // Color palette index for each item; Black/Orange have no exact match
    // in the 7-color palette, so they fall through to the nearest hue.
    const lines = [
      [' 1 - Black',  2],  // bright white stand-in
      [' 2 - White',  2],
      [' 3 - Blue',   1],  // cyan
      [' 4 - Red',    6],
      [' 5 - Purple', 5],
      [' 6 - Yellow', 4],
      [' 7 - Orange', 4],
      [' 8 - Green',  3],
    ];
    for (const [text, col] of lines) io.println(text, col);
    io.cr();
    const n = await io.numbersPrompt('# of Color :', 1, 8);
    session.character.favColor = n - 1;
    session.setNext(new CreateCharMenuState());
  }
}

// Mirrors EnterTownFromCreateState in the Go side.
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

// Toggling gender re-skews a handful of stats up or down per the original.
function toggleGender(session) {
  const c = session.character;
  const { io } = session;
  if (c.gender) {
    io.println('You are Now Female.', 3);
    c.gender = false;
    c.strength -= 3;
    c.maxPsyche += 1;
    c.dexterity += 2;
    c.maxHP -= 2;
    c.hitPoints = c.maxHP;
    c.faith += 2;
  } else {
    io.println('You are Now Male.', 3);
    c.gender = true;
    c.strength += 3;
    c.maxPsyche -= 1;
    c.dexterity -= 2;
    c.maxHP += 2;
    c.hitPoints = c.maxHP;
    c.faith -= 2;
  }
}

function showScale(io) {
  io.cr();
  io.println('=--=-- Attribute Scale ---=--=', 3);
  io.println('  1-3   : Poor', 6);
  io.println('  4-6   : Below Average', 4);
  io.println('  7-9   : Average', 1);
  io.println('  10-12 : Above Average', 1);
  io.println('  13-15 : Good', 3);
  io.println('  16-18 : Excellent', 3);
  io.println('  19+   : Exceptional', 3);
  io.cr();
}
