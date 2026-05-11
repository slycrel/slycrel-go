import { WhereType } from '../../model/enums.js';
import { loadInn, saveInn, saveCharacter } from '../../store/local.js';
import { TownState } from './town.js';

// Inn — port of internal/game/states/inn.go. The inn has shared mutable
// state (rooms, owner, rate, safe) stored under one localStorage key.

export class InnState {
  async enter(session) {
    session.character.location = WhereType.TheInn;
    session.io.clear();
    await session.io.showAnsiFile('inn_menu');
    session.setNext(new InnPromptState());
  }
}

export class InnPromptState {
  async enter(session) {
    const { io } = session;
    io.println('[L]ook around, [R]ent a Room, [T]alk to Innkeeper, [V]iew, [Q]uit, [*]Owner Menu, [?]Help', 1);
    io.cr();
    const choice = await io.lettersPrompt('Your choice?', 'LRTQV?*');
    io.cr();
    switch (choice) {
      case 'L':
        io.println('You look around the lobby of the inn and find it to be', 2);
        io.println('in a serious array of disorder. You wonder if you', 2);
        io.println('should really rent a room here or not.', 2);
        io.cr();
        session.setNext(new InnPromptState());
        return;
      case 'R': session.setNext(new RentRoomState()); return;
      case 'T':
        io.println('You ask the innkeep how business has been. Quickly', 2);
        io.println('you realize you should have never asked such an open', 2);
        io.println('ended question as the inkeep rambles on and on......', 2);
        io.println('and on......', 2);
        io.cr();
        session.setNext(new InnPromptState());
        return;
      case 'V':
        io.println("You look around the inn's lobby and quickly find", 2);
        io.println("there really isn't anything to view.", 2);
        io.cr();
        session.setNext(new InnPromptState());
        return;
      case 'Q':
        session.character.location = WhereType.TheTown;
        session.setNext(new TownState());
        return;
      case '*': session.setNext(new InnOwnerMenuState()); return;
      case '?': session.setNext(new InnState()); return;
    }
    session.setNext(new InnPromptState());
  }
}

export class RentRoomState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    const inn = loadInn();

    // Already have a room?
    if (inn.rooms.some(r => r.who === c.name && r.daysLeft > 0)) {
      io.println('Sorry, you can only have one room.', 3);
      session.setNext(new InnPromptState());
      return;
    }
    const emptySlot = inn.rooms.findIndex(r => !r.who || r.daysLeft <= 0);
    if (emptySlot === -1) {
      io.println('Sorry There are no more rooms Avbl. for Rent', 3);
      session.setNext(new InnPromptState());
      return;
    }
    if (!await io.yesNoQuestion('Would you like to Rent a room?')) {
      session.setNext(new InnPromptState());
      return;
    }
    const days = await io.numbersPrompt('For how many days? [0-14]:', 0, 14);
    if (days === 0) {
      io.println("FINE! Don't Rent a room, go sleep in the streets!", 3);
      session.setNext(new InnPromptState());
      return;
    }
    const rate = inn.curRate > 0 ? inn.curRate : 5;
    const totalCost = days * rate;
    if (c.coinsHand < totalCost) {
      io.println("I'm not giving the room away, come back when you have enough coins!", 3);
      io.println(`Note, it costs ${rate} coins per day to rent a room.`, 1);
      session.setNext(new InnPromptState());
      return;
    }
    const lockRating = await io.numbersPrompt('What Security Rating would you like on your lock? [0-10]:', 0, 10);
    let lockCost = lockRating * 2;
    let actualLock = lockRating;
    if (lockRating === 0) {
      io.println('"Ok, don\'t blame me if Corenne slips in late at night," grins the Innkeeper.', 3);
    } else if (c.coinsHand < totalCost + lockCost) {
      io.println("I'm not giving the Lock away!", 3);
      io.println('Note, it costs twice the lock rating for the lock.', 1);
      lockCost = 0;
      actualLock = 0;
    }

    io.println('Item:              Days:                  Cost:', 2);
    io.println(` Room               ${String(days).padEnd(23)}$${totalCost}`, 1);
    if (lockCost > 0) {
      io.println(` Lock               ${String(1).padEnd(23)}$${lockCost}`, 1);
    }
    io.println(`Total:                                     $${totalCost + lockCost}`, 3);

    c.coinsHand -= totalCost + lockCost;
    inn.rooms[emptySlot] = { who: c.name, daysLeft: days, lock: actualLock };
    inn.safe += totalCost;
    saveInn(inn);
    saveCharacter(c);
    io.cr();
    session.setNext(new InnPromptState());
  }
}

function listRooms(io, inn) {
  io.cr();
  io.println('  Room  Occupant             Days Left  Lock', 4);
  io.println('  ----  --------             ---------  ----', 1);
  inn.rooms.forEach((r, i) => {
    const who = (r.who && r.daysLeft > 0) ? r.who : '(empty)';
    io.println(`  ${String(i + 1).padStart(2)}    ${who.padEnd(20)} ${String(r.daysLeft).padStart(5)}      ${r.lock}`, 1);
  });
  io.cr();
}

export class InnOwnerMenuState {
  async enter(session) {
    const c = session.character;
    const { io } = session;
    const inn = loadInn();

    if (inn.owner && inn.owner !== c.name) {
      io.println('You are not the innkeeper!', 6);
      session.setNext(new InnPromptState());
      return;
    }
    if (!inn.owner) {
      inn.owner = c.name;
      saveInn(inn);
      io.println('You are now the innkeeper!', 3);
      io.cr();
    }
    io.cr();
    io.println('=--=-- Innkeeper\'s Menu ---=--=', 3);
    io.cr();
    io.println('1) Change Daily Rate', 1);
    io.println('2) List Rooms', 1);
    io.println('3) Kick Out Player', 1);
    io.println('4) View Inn Stats', 1);
    io.println('5) Close/Open Inn', 1);
    io.println('6) Transfer Ownership', 1);
    io.cr();
    const num = await io.numbersPrompt('[1-6, 0=Quit]:', 0, 6);
    switch (num) {
      case 0: session.setNext(new InnPromptState()); return;
      case 1: {
        const newRate = await io.numbersPrompt('Daily Charge for a Room:', 1, 32000);
        inn.curRate = newRate;
        saveInn(inn);
        io.println(`Rate set to ${newRate} coins/day.`, 3);
        session.setNext(new InnOwnerMenuState());
        return;
      }
      case 2:
        listRooms(io, inn);
        session.setNext(new InnOwnerMenuState());
        return;
      case 3: {
        listRooms(io, inn);
        const roomNum = await io.numbersPrompt('Which room to vacate? [1-10]:', 1, 10);
        inn.rooms[roomNum - 1] = { who: '', daysLeft: 0, lock: 0 };
        saveInn(inn);
        io.println('Player removed.', 3);
        session.setNext(new InnOwnerMenuState());
        return;
      }
      case 4:
        io.cr();
        io.println(`  Owner: ${inn.owner}`, 4);
        io.println(`  Rate: ${inn.curRate} coins/day`, 1);
        io.println(`  Safe: ${inn.safe} coins`, 1);
        io.println(`  Status: ${inn.open ? 'Open' : 'Closed'}`, 1);
        io.cr();
        session.setNext(new InnOwnerMenuState());
        return;
      case 5:
        inn.open = !inn.open;
        saveInn(inn);
        io.println(`The Inn is now ${inn.open ? 'OPEN.' : 'CLOSED.'}`, inn.open ? 3 : 6);
        session.setNext(new InnOwnerMenuState());
        return;
      case 6: {
        const newOwner = (await io.textPrompt('Who is to be the new Inn Owner?', 20)).trim();
        if (newOwner) {
          inn.owner = newOwner;
          saveInn(inn);
          io.println(`Ownership transferred to ${newOwner}.`, 3);
        }
        session.setNext(new InnOwnerMenuState());
        return;
      }
    }
    session.setNext(new InnOwnerMenuState());
  }
}
