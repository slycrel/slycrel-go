import { WhereType } from '../../model/enums.js';
import { saveCharacter } from '../../store/local.js';
import { TownState } from './town.js';

// BankState — port of internal/game/states/bank.go (BankState).
// Shows the bank ANSI banner, then routes to the action prompt.
export class BankState {
  async enter(session) {
    session.character.location = WhereType.TheBank;
    session.io.clear();
    await session.io.showAnsiFile('bank_menu');
    session.setNext(new BankPromptState());
  }
}

// BankPromptState — port of BankPromptState. Shows balances + menu choices.
export class BankPromptState {
  async enter(session) {
    const c = session.character;
    const { io } = session;

    io.println(`You have ${c.coinsBank} Coins in Bank and ${c.coinsHand} Coins on Hand.`, 4);
    io.println('[D]eposit, [W]ithdraw, [Q]uit, [?]Help', 1);
    io.cr();

    const choice = await io.lettersPrompt('Your choice?', 'DWQ?');
    io.cr();

    switch (choice) {
      case 'D':
        session.setNext(new BankDepositState());
        return;
      case 'W':
        session.setNext(new BankWithdrawState());
        return;
      case 'Q':
        c.location = WhereType.TheTown;
        session.setNext(new TownState());
        return;
      case '?':
        session.setNext(new BankState());
        return;
    }
    session.setNext(new BankPromptState());
  }
}

// BankDepositState — port of BankDepositState.
export class BankDepositState {
  async enter(session) {
    const c = session.character;
    const { io } = session;

    if (c.coinsHand <= 0) {
      io.println('You have no coins to deposit!', 6);
      io.cr();
      session.setNext(new BankPromptState());
      return;
    }

    io.println(`[Max: ${c.coinsHand}, 0 = Abort]`, 5);
    io.cr();
    const amount = await io.numbersPrompt('Deposit:', 0, c.coinsHand);

    if (amount > 0) {
      c.coinsBank += amount;
      c.coinsHand -= amount;
      io.println(`You give the banker ${amount} coins.`, 3);
      io.cr();
      saveCharacter(c);
    }

    session.setNext(new BankPromptState());
  }
}

// BankWithdrawState — port of BankWithdrawState.
export class BankWithdrawState {
  async enter(session) {
    const c = session.character;
    const { io } = session;

    if (c.coinsBank <= 0) {
      io.println('You have no coins in the bank!', 6);
      io.cr();
      session.setNext(new BankPromptState());
      return;
    }

    io.println(`[Max: ${c.coinsBank}, 0 = Abort]`, 5);
    io.cr();
    const amount = await io.numbersPrompt('Withdraw:', 0, c.coinsBank);

    if (amount > 0) {
      c.coinsHand += amount;
      c.coinsBank -= amount;
      io.println(`The Banker hands you ${amount} coins.`, 3);
      io.cr();
      saveCharacter(c);
    }

    session.setNext(new BankPromptState());
  }
}
