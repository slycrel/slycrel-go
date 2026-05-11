import { Session } from './game/session.js';
import { BeginState } from './game/states/begin.js';

const terminal = document.getElementById('terminal');
const session = new Session(terminal);
terminal.focus();
session.run(new BeginState());
