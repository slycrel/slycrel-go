import { isWeaponEmpty, isArmorEmpty, genderString } from '../../model/character.js';
import { COLOR_NAMES } from '../../model/enums.js';

// ViewCharacterState — port of internal/game/states/character_create.go.
// Renders the character sheet, then pops the return stack.
export class ViewCharacterState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    if (!c) {
      io.println('No character to display.', 6);
      session.popReturn();
      return;
    }

    io.cr();
    io.println('=--=-- Character Sheet ---=--=', 3);
    io.cr();
    io.println(`  Name:       ${c.name}`, 4);
    io.println(`  Class:      ${c.charClass}`, 1);
    io.println(`  Gender:     ${genderString(c)}`, 1);
    io.println(`  Level:      F:${c.fighterLvl}  T:${c.thiefLvl}  M:${c.mageLvl}`, 1);
    io.cr();
    io.println(`  Hit Points: ${c.hitPoints}/${c.maxHP}`, 2);
    io.println(`  Psyche:     ${c.psyche}/${c.maxPsyche}`, 2);
    io.cr();
    io.println(`  Strength:   ${c.strength}`, 1);
    io.println(`  Dexterity:  ${c.dexterity}`, 1);
    io.println(`  Speed:      ${c.speed}`, 1);
    io.cr();
    io.println(`  Coins:      ${c.coinsHand} (Bank: ${c.coinsBank})`, 4);
    io.println(`  Experience: ${c.totalExperience} (Spending: ${c.spendingExperience})`, 1);
    io.cr();
    io.println(`  Fame: ${c.fame}  Honor: ${c.honor}  Faith: ${c.faith}`, 1);
    io.println(`  Color:      ${COLOR_NAMES[c.favColor] ?? 'Unknown'}`, 1);

    for (let i = 0; i < c.weapons.length; i++) {
      const w = c.weapons[i];
      if (!isWeaponEmpty(w)) {
        io.println(`  Weapon ${i + 1}:   ${w.name} (Strike:${w.strike} Range:${w.range})`, 1);
      }
    }
    for (let i = 0; i < c.armor.length; i++) {
      const a = c.armor[i];
      if (!isArmorEmpty(a)) {
        io.println(`  Armor ${i + 1}:    ${a.name} (Defense:${a.defense})`, 1);
      }
    }

    io.cr();
    session.popReturn();
  }
}
