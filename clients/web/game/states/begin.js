import { EnterSlycrelState } from './enter_slycrel.js';

// BeginState — title screen + main entry menu.
// Mirrors internal/game/states/begin.go (BeginState).
export class BeginState {
  async enter(session) {
    const { io } = session;
    io.clear();
    await io.showAnsiFile('opening');
    io.cr();
    io.println('=--=--  Welcome to the Realm of Slycrel  ---=--=', 3);
    io.cr();
    io.println('[E]nter the Realm', 4);
    io.println('[V]iew Guild Lists', 4);
    io.println('[C]reate/View Character', 4);
    io.println('[L]eave', 4);
    io.cr();

    const choice = await io.lettersPrompt('Your Selection >', 'EVCL');
    io.cr();

    switch (choice) {
      case 'E':
      case 'C':
        session.setNext(new EnterSlycrelState());
        return;
      case 'V':
        io.println('Guild lists not yet implemented.', 1);
        await io.pausePrompt();
        session.setNext(new BeginState());
        return;
      case 'L':
        io.println('Exiting Slycrel...', 3);
        session.quit();
        return;
    }
  }
}
