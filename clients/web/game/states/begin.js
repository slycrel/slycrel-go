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

    // TODO: real transitions once enter_slycrel / character_create / quit
    // states are ported. For now, log and loop back to the title.
    io.cr();
    io.println(`(stub) you chose ${choice} — transitions land in the next commit.`, 6);
    io.cr();
    await io.lettersPrompt('--press any key--', 'ABCDEFGHIJKLMNOPQRSTUVWXYZ');
    return new BeginState();
  }
}
