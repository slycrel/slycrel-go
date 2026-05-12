import { WhereType } from '../../model/enums.js';
import { randBetween } from '../../mechanics/rand.js';
import { saveCharacter, listCharacters } from '../../store/remote.js';
import { TownState } from './town.js';
import { DeadState } from './dead.js';

// Drink prices indexed by menu number (1-15). 0 unused.
const DRINK_COSTS = [0, 0, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 7, 9, 10, 10];

// Flavor text for each drink (verbatim from internal/game/states/tavern.go).
// Multi-line strings — emit one line per div via println.
const DRINK_TEXTS = [
  '',
  "Lynx gives you a hard look and shifts his eyes slightly.\n\"Here. But don't let anyone know I still have this stuff...\"",
  'The bartender suppresses a smile as he hands you the drink.',
  "Lynx hands you an ale. When you don't tip him, he gets an\nangry look on his face, and stalks off.",
  "Lynx hands you the drink. You think he winks as you turn,\nbut you can't be certain.",
  "Lynx stirs the cocktail slightly and hands it to you.\n\nEWW!!! He licked the spoon and stuck it in that guy's drink!",
  'Mmmm! Delicious. (You hope no faeries were still in it...)',
  'The Bartender hands you the wine, and tells you:\n"That woman over in the corner... I think she likes you!"',
  'Lynx pours whiskey into a small glass. He licks the rim of\nthe bottle as he walks off.',
  "The bartender tells you this stuff's expensive, and how hard it's\ngoing to be to replace.\n\nYou just shrug, and gleefully chug down your drink.",
  'You shudder as the life essence of the poor, innocent,\nunsuspecting little dwarf rolls down your throat.',
  'The bartender steps into the back for a moment. When he\nreturns, he is carrying a slightly thicker mug with a lid.\n"I ain\'t responsible," he tells you as he hands it over.\n\nThe drink oozes down your throat, nearly suffocating you.',
  'The barkeep reaches under the counter, and pulls out a large, black\njug. He pours you a tiny glass.\n\nThe drink burns as it goes down.',
  'The bartender hands you a steaming glass of amber liquid.\n\nThe drink runs smoothly down your throat.\nYou suddenly feel the need to use the restroom.',
  'The bartender laughs, and says, "You HONESTLY thought that\nI\'d even let Josepi IN here?" He throws back his head and laughs.\n\nwell, you DID ask... (And pay, too!)',
  'Lynx shrugs as if to say "hey, it\'s not my life" and\nhands it to you.\n\nYou suddenly feel as if the world turns upside down.\nYou sway and pass out...\n\nYou\'d better check your stats bud!!!',
];

const FLIRT_FAIL_TEXTS = [
  '',
  "Corenne doesn't seem to see you, but Grego perks right up\nand runs over, eagerly awaiting your order...",
  'You take her hand as she turns away, but just as you lift it to your\nlips, she jerks free, quickly walking to a much better looking customer.',
  'You bravely announce that Corenne has the best looking elbows that\nyou have ever seen. She goes red all over and you congratulate\nyourself. You do wonder, however, why Lynx and the other bar\npatrons are laughing at you...',
  "You stand, gathering everyone's attention with a manly <ahem>.\nSomeone in the back snickers as the bar goes silent. You start\nrambling off miscellaneous heroic deeds that you have done. After\nless than a minute, you feel slightly sick. You wonder at fate,\nand all its quirks as you spill your lunch on the floor...",
  'You very obviously wink seductively at Corenne several times.\nShe hurries over to you and asks if your eye is okay. You\nnearly choke as you mumble you\'re fine...',
  "You compliment Corenne's attributes and how bouncy they are.\nYou suddenly find yourself staring at the ceiling with a\nsplit lip.",
  'As Corenne walks by, you ask her to dance. She stares at you\nfor a moment, and reminds you that there is no music playing.\nGrego quickly snakes in, and informs you he will dance.\n\nYou shudder at the thought...',
  'You ask Corenne if she would like to sit a while and share a meal.\nShe replies that she is on a diet, and that she needs to keep\nher figure for the MEN...\n\nYour pride is hurt....',
];

// println every line of a multi-line string. Many Go calls pass a single
// Outln with embedded \n that the Go IO splits internally.
function printLines(io, text, color) {
  for (const line of text.split('\n')) io.println(line, color);
}

export class TavernState {
  async enter(session) {
    session.character.location = WhereType.TheTavern;
    session.io.clear();
    await session.io.showAnsiFile('tavern_menu');
    session.setNext(new TavernPromptState());
  }
}

export class TavernPromptState {
  async enter(session) {
    const { io } = session;
    io.println('[F]lirt, [H]ang Around, [L]isten, [O]rder Drink, [Q]uit, [T]alk to Lynx, [V]iew Guilds, [?]', 1);
    io.cr();
    const choice = await io.lettersPrompt('Well? :', 'FHLOQTV?');
    io.cr();
    switch (choice) {
      case 'F': session.setNext(new CorennePromptState()); return;
      case 'H': session.setNext(new HangAroundState()); return;
      case 'L':
        io.println('You sit on a bar stool for a while trying to listen', 2);
        io.println('to every whisper from across the room. After some time', 2);
        io.println("you decide you can't hear anything and wonder if you", 2);
        io.println('should just leave the tavern.', 2);
        io.cr();
        session.setNext(new TavernPromptState());
        return;
      case 'O': session.setNext(new OrderDrinkState()); return;
      case 'Q':
        session.character.location = WhereType.TheTown;
        session.setNext(new TownState());
        return;
      case 'T': session.setNext(new TalkToLynxState()); return;
      case 'V': session.setNext(new ViewGuildsState()); return;
      case '?': session.setNext(new TavernState()); return;
    }
    session.setNext(new TavernPromptState());
  }
}

export class OrderDrinkState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    io.cr();
    io.println('Whaddya want?', 1);
    io.cr();
    const drinks = [
      ' 1- Fruit Juice                  0',
      ' 2- Elven Ale                    1',
      ' 3- Ale                          2',
      ' 4- Mead                         2',
      ' 5- Cocktail                     3',
      ' 6- Faery Water                  3',
      ' 7- Fine Wine                    4',
      ' 8- Whiskey                      4',
      ' 9- Ogre Sweat                   5',
      '10- Dwarf Spirits                5',
      '11- Dragon Blood                 6',
      '12- Potency                      7',
      '13- Kwick                        9',
      "14- Josepi's Molotov Cocktail   10",
      "15- Auneletho's Surprise        10",
    ];
    for (const d of drinks) io.println(d, 2);
    io.cr();
    io.println(`$${c.coinsHand} in Hand`, 3);
    io.cr();

    const num = await io.numbersPrompt('[1-15, 0=Quit]:', 0, 16);
    if (num === 0) {
      session.setNext(new TavernPromptState());
      return;
    }
    // Easter egg: ordering 16 kills you (preserved from the original).
    if (num === 16) {
      io.println("You just can't follow directions, can you? I said 1 to 15, NOT 16...", 6);
      io.println("But that's okay. I won't mind killing you.", 2);
      io.cr();
      io.println('                   <ZZZZZZZZZZZZZZZT!!!>', 4);
      c.alive = false;
      await saveCharacter(c);
      session.setNext(new DeadState());
      return;
    }

    const cost = DRINK_COSTS[num];
    if (c.coinsHand < cost) {
      io.println("You can't afford that drink!", 6);
      io.cr();
    } else {
      c.coinsHand -= cost;
      io.cr();
      printLines(io, DRINK_TEXTS[num], randBetween(0, 7));
      io.cr();
      // Fruit juice cuts the flirt cooldown by one.
      if (num === 1) {
        c.flirt1 -= 1;
        if (c.flirt1 < 1) c.flirt1 = 1;
      }
      await saveCharacter(c);
    }
    io.cr();
    session.setNext(new TavernPromptState());
  }
}

export class TalkToLynxState {
  async enter(session) {
    const { io } = session;
    io.println("Sorry, Lynx doesn't feel up to talking right now. He's busy", randBetween(0, 7));
    io.println('playing with the barmaid...', randBetween(0, 7));
    io.cr();
    session.setNext(new TavernPromptState());
  }
}

export class HangAroundState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    io.println('You lounge around the bar for a while.', randBetween(0, 7));
    io.cr();
    if (randBetween(1, 100) <= 50) {
      io.println("This is boring. Let's do something else.", randBetween(0, 7));
      io.cr();
      await io.pausePrompt('ZZZZZzzzzzz...');
    } else {
      io.println('You nod off after a while....', randBetween(0, 7));
      io.cr();
      io.println('ZZzzzZZzzzZZ', 1);
      io.cr();
      io.println('As you wake up, your money purse feels lighter... You must have spent more than you thought...', 2);
      const divisor = randBetween(10, 25);
      if (divisor > 0) c.coinsHand -= Math.floor(c.coinsHand / divisor);
      await saveCharacter(c);
    }
    session.setNext(new TavernPromptState());
  }
}

export class CorennePromptState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    if (c.flirt1 < 0) {
      io.println("Corenne will probably think you're desperate, and you wouldn't want that.", 3);
      session.setNext(new TavernPromptState());
      return;
    }
    io.cr();
    const options = [
      '1] Grin Foolishly', '2] Kiss her hand', '3] Compliment and tip her',
      '4] Brag about yourself', '5] Wink seductively', '6] Compliment her',
      '7] Ask her to dance', '8] Share a meal',
    ];
    for (const o of options) io.println(o, 6);
    io.cr();
    const num = await io.numbersPrompt("Whats it gonna be Cutie?", 1, 8);
    io.cr();
    if (randBetween(1, 100) < 18 + c.flirt1) {
      io.println('Corenne smiles at you as she serves you...', 5);
      io.cr();
      io.println('Success!', 5);
      c.flirt1 += 1;
    } else {
      if (num >= 1 && num <= 8) printLines(io, FLIRT_FAIL_TEXTS[num], 2);
      // Failure penalties per choice.
      if (num === 3 || num === 4 || num === 8) c.honor -= 1;
      else if (num === 6) c.hitPoints -= randBetween(1, Math.max(1, Math.floor(c.hitPoints / 4)));
      c.flirt1 -= 2;
      if (c.flirt1 <= 0) c.flirt1 = 1;
    }
    // Flip sign — you can only flirt once per visit.
    c.flirt1 = -c.flirt1;
    await saveCharacter(c);
    io.cr();
    session.setNext(new TavernPromptState());
  }
}

export class ViewGuildsState {
  async enter(session) {
    const { io } = session;
    const chars = await listCharacters();
    if (chars.length === 0) {
      io.println('No adventurers have registered yet.', 1);
      session.setNext(new TavernPromptState());
      return;
    }
    io.cr();
    io.println('=--=-- Guild Rankings ---=--=', 3);
    io.cr();
    io.println(' Name                 Class      F.Lvl T.Lvl M.Lvl       XP', 4);
    io.println(' ------------------------------------------------------------', 1);
    for (const ch of chars) {
      const name = ch.name.padEnd(20).slice(0, 20);
      const cls = (ch.charClass || '').padEnd(10).slice(0, 10);
      const f = String(ch.fighterLvl ?? 0).padStart(5);
      const t = String(ch.thiefLvl ?? 0).padStart(5);
      const m = String(ch.mageLvl ?? 0).padStart(5);
      const xp = String(ch.totalExperience ?? 0).padStart(8);
      io.println(` ${name} ${cls} ${f} ${t} ${m} ${xp}`, 1);
    }
    io.cr();
    await io.pausePrompt();
    session.setNext(new TavernPromptState());
  }
}
